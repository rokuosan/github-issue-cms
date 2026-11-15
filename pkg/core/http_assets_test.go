package core

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"

	"github.com/stretchr/testify/require"
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

// imageRoundTripper keeps the original URL visible to the authentication and
// redirect logic without contacting GitHub.
type imageRoundTripper func(*http.Request) (*http.Response, error)

func (f imageRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func TestHTTPImageRepository_TokenDestinations(t *testing.T) {
	tests := []struct {
		name     string
		urls     []string
		wantAuth []bool
	}{
		{"GitHub attachment", []string{"https://github.com/user-attachments/assets/id"}, []bool{true}},
		{"legacy image", []string{"https://user-images.githubusercontent.com/id.png"}, []bool{true}},
		{"private image", []string{"https://private-user-images.githubusercontent.com/id.png"}, []bool{true}},
		{"case insensitive", []string{"https://GitHub.com/user-attachments/assets/id"}, []bool{true}},
		{"external HTTPS", []string{"https://cdn.example.com/id.png"}, []bool{false}},
		{"HTTP", []string{"http://github.com/user-attachments/assets/id"}, []bool{false}},
		{"subdomain", []string{"https://evil.github.com/id.png"}, []bool{false}},
		{"host suffix", []string{"https://github.com.example.com/id.png"}, []bool{false}},
		{"custom port", []string{"https://github.com:8443/id.png"}, []bool{false}},
		{"userinfo", []string{"https://user@github.com/id.png"}, []bool{false}},
		{"same host redirect", []string{"https://github.com/start", "https://github.com/image"}, []bool{true, true}},
		{"known host redirect", []string{"https://github.com/start", "https://private-user-images.githubusercontent.com/image"}, []bool{true, true}},
		{"external redirect", []string{"https://github.com/start", "https://cdn.example.com/image"}, []bool{true, false}},
		{"subdomain redirect", []string{"https://github.com/start", "https://evil.github.com/image"}, []bool{true, false}},
		{"HTTP redirect", []string{"https://github.com/start", "http://github.com/image"}, []bool{true, false}},
		{"external initial request", []string{"https://cdn.example.com/start", "https://github.com/image"}, []bool{false, false}},
		{"redirect chain", []string{"https://github.com/start", "https://cdn.example.com/image", "https://user-images.githubusercontent.com/image"}, []bool{true, false, true}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := NewHTTPImageRepository("test-token").(*HTTPImageRepository)
			calls := 0
			repo.client.Transport = imageRoundTripper(func(req *http.Request) (*http.Response, error) {
				i := calls
				calls++
				require.Less(t, i, len(tt.urls))
				assert.Equal(t, tt.urls[i], req.URL.String())
				want := ""
				if tt.wantAuth[i] {
					want = "token test-token"
				}
				if req.URL.User != nil && !tt.wantAuth[i] {
					assert.NotContains(t, req.Header.Get("Authorization"), "test-token")
				} else {
					assert.Equal(t, want, req.Header.Get("Authorization"))
				}
				resp := &http.Response{StatusCode: http.StatusOK, Header: http.Header{"Content-Type": {"image/png"}}, Body: io.NopCloser(strings.NewReader("PNG"))}
				if i+1 < len(tt.urls) {
					resp.StatusCode = http.StatusFound
					resp.Header.Set("Location", tt.urls[i+1])
				}
				return resp, nil
			})
			asset, err := repo.Fetch(context.Background(), NewImage(tt.urls[0], "", 0))
			require.NoError(t, err)
			asset.Body.Close()
			assert.Equal(t, len(tt.urls), calls)
		})
	}
}

func TestHTTPImageRepository_Fetch_AuthFallback(t *testing.T) {
	for _, fallbackOK := range []bool{true, false} {
		t.Run(fmt.Sprint(fallbackOK), func(t *testing.T) {
			var logs bytes.Buffer
			logger := slog.New(slog.NewTextHandler(&logs, nil))
			repo := NewHTTPImageRepositoryWithLogger("test-token", logger).(*HTTPImageRepository)
			calls := 0
			repo.client.Transport = imageRoundTripper(func(req *http.Request) (*http.Response, error) {
				calls++
				status := http.StatusNotFound
				if calls == 1 {
					assert.Equal(t, "token test-token", req.Header.Get("Authorization"))
				} else {
					assert.Empty(t, req.Header.Get("Authorization"))
					if fallbackOK {
						status = http.StatusOK
					}
				}
				return &http.Response{StatusCode: status, Header: http.Header{"Content-Type": {"image/png"}}, Body: io.NopCloser(strings.NewReader("PNG"))}, nil
			})
			asset, err := repo.Fetch(context.Background(), NewImage("https://github.com/user-attachments/assets/id", "", 0))
			if fallbackOK {
				require.NoError(t, err)
				asset.Body.Close()
			} else {
				require.Error(t, err)
				assert.Contains(t, err.Error(), "authenticated request failed")
				assert.Contains(t, err.Error(), "unauthenticated fallback failed")
			}
			assert.Equal(t, 2, calls)
			assert.Contains(t, logs.String(), "authenticated image download failed; retrying without token")
			assert.NotContains(t, logs.String(), "test-token")
		})
	}
}

func TestHTTPImageRepository_RedirectLimit(t *testing.T) {
	repo := NewHTTPImageRepository("test-token").(*HTTPImageRepository)
	calls := 0
	repo.client.Transport = imageRoundTripper(func(req *http.Request) (*http.Response, error) {
		calls++
		return &http.Response{StatusCode: http.StatusFound, Header: http.Header{"Location": {"/loop"}}, Body: io.NopCloser(strings.NewReader(""))}, nil
	})
	_, err := repo.Fetch(context.Background(), NewImage("https://github.com/loop", "", 0))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "stopped after 10 redirects")
	assert.Equal(t, 2*maxRedirects, calls)
}
