package core

import (
	"bytes"
	"context"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewHTTPImageRepository(t *testing.T) {
	tests := []struct {
		name  string
		token string
	}{
		{
			name:  "with token",
			token: "test-token",
		},
		{
			name:  "without token",
			token: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := NewHTTPImageRepository(tt.token)
			assert.NotNil(t, repo)
		})
	}
}

func TestHTTPImageRepository_Download(t *testing.T) {
	t.Run("successful image download", func(t *testing.T) {
		// Test HTTP server.
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "image/png")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("FAKE_PNG_DATA"))
		}))
		defer server.Close()

		repo := NewHTTPImageRepository("")

		image := &Image{
			URL:  server.URL,
			Time: "2021-01-01_000000",
			ID:   0,
		}

		asset, err := repo.Fetch(context.Background(), image)
		assert.NoError(t, err)
		requireBody := func() string {
			defer asset.Body.Close()
			data, err := io.ReadAll(asset.Body)
			assert.NoError(t, err)
			return string(data)
		}
		assertEqualCmp(t, "image/png", asset.ContentType)
		assertEqualCmp(t, "FAKE_PNG_DATA", requireBody())
	})

	t.Run("HTTP 404 error", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
		}))
		defer server.Close()

		repo := NewHTTPImageRepository("")

		image := &Image{
			URL:  server.URL,
			Time: "2021-01-01_000000",
			ID:   0,
		}

		_, err := repo.Fetch(context.Background(), image)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "bad response")
	})

	t.Run("invalid content type", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "text/html")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("<html></html>"))
		}))
		defer server.Close()

		repo := NewHTTPImageRepository("")
		image := &Image{
			URL:  server.URL,
			Time: "2021-01-01_000000",
			ID:   0,
		}

		_, err := repo.Fetch(context.Background(), image)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "bad response")
	})

	t.Run("request with token", func(t *testing.T) {
		var authHeader string
		server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader = r.Header.Get("Authorization")
			w.Header().Set("Content-Type", "image/png")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("PNG"))
		}))
		defer server.Close()

		repo := NewHTTPImageRepository("test-token").(*HTTPImageRepository)
		repo.client = server.Client()
		image := &Image{
			URL:  server.URL,
			Time: "2021-01-01_000000",
			ID:   0,
		}

		asset, err := repo.Fetch(context.Background(), image)
		assert.NoError(t, err)
		defer asset.Body.Close()
		assertEqualCmp(t, "token test-token", authHeader)
	})

	t.Run("does not send token to HTTP URL", func(t *testing.T) {
		var authHeader string
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader = r.Header.Get("Authorization")
			w.Header().Set("Content-Type", "image/png")
			_, _ = w.Write([]byte("PNG"))
		}))
		defer server.Close()

		repo := NewHTTPImageRepository("test-token")
		asset, err := repo.Fetch(context.Background(), NewImage(server.URL, "2021-01-01_000000", 0))
		assert.NoError(t, err)
		defer asset.Body.Close()
		assert.Empty(t, authHeader)
	})

	t.Run("strips token on redirect from HTTPS to HTTP", func(t *testing.T) {
		var httpsAuthHeader string
		var httpAuthHeader string
		httpServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			httpAuthHeader = r.Header.Get("Authorization")
			w.Header().Set("Content-Type", "image/png")
			_, _ = w.Write([]byte("PNG"))
		}))
		defer httpServer.Close()

		httpsServer := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			httpsAuthHeader = r.Header.Get("Authorization")
			http.Redirect(w, r, httpServer.URL, http.StatusFound)
		}))
		defer httpsServer.Close()

		repo := NewHTTPImageRepository("test-token").(*HTTPImageRepository)
		repo.client = httpsServer.Client()
		asset, err := repo.Fetch(context.Background(), NewImage(httpsServer.URL, "2021-01-01_000000", 0))
		assert.NoError(t, err)
		defer asset.Body.Close()
		assert.Equal(t, "token test-token", httpsAuthHeader)
		assert.Empty(t, httpAuthHeader)
	})
}

func TestHTTPImageRepository_Fetch_ReportsAuthenticatedAndFallbackFailures(t *testing.T) {
	var logs bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&logs, &slog.HandlerOptions{Level: slog.LevelDebug}))
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	repo := NewHTTPImageRepositoryWithLogger("test-token", logger).(*HTTPImageRepository)
	repo.client = server.Client()
	_, err := repo.Fetch(context.Background(), NewImage(server.URL, "2021-01-01_000000", 0))

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "authenticated request failed")
	assert.Contains(t, err.Error(), "unauthenticated fallback failed")
	assert.Contains(t, logs.String(), "authenticated image download failed; retrying without token")
	assert.NotContains(t, logs.String(), "test-token")
}

func TestHTTPImageRepository_Download_InvalidURL(t *testing.T) {
	repo := NewHTTPImageRepository("")

	image := &Image{
		URL:  "://invalid-url",
		Time: "2021-01-01_000000",
		ID:   0,
	}

	_, err := repo.Fetch(context.Background(), image)
	assert.Error(t, err)
}
