package core

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func contentHTTPFixture(t *testing.T) *ContentSnapshot {
	t.Helper()
	snapshot, err := NewContentSnapshot([]ContentArticle{
		{ID: 1, Title: "Go", Content: "Cloudflare Workers", Date: "2026-01-01", Category: "tech", Tags: []string{"Go", "Web"}},
		{ID: 2, Title: "Web", Content: "Cloudflare", Date: "2026-01-02", Category: "tech", Tags: []string{"Web"}},
		{ID: 3, Title: "Travel", Content: "Japan", Date: "2026-01-03", Category: "travel", Tags: []string{"Web"}},
	}, 1)
	require.NoError(t, err)
	snapshot.Assets = map[string]ContentAsset{"/images/1/0.png": {Data: []byte("image bytes"), ContentType: "image/png"}}
	return snapshot
}

func TestContentHandlerQuery(t *testing.T) {
	handler := NewContentHandler(contentHTTPFixture(t))
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/data/articles?q=cloudflare&tag=Web&category=tech&page=2&per_page=1", nil))
	require.Equal(t, http.StatusOK, recorder.Code)
	var page ArticlePage
	require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &page))
	require.Equal(t, 2, page.Pagination.TotalItems)
	require.Equal(t, 2, page.Pagination.TotalPages)
	require.Len(t, page.Items, 1)
	require.Equal(t, 1, page.Items[0].ID)

	recorder = httptest.NewRecorder()
	handler.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/data/articles?tag=Go&tag=Web", nil))
	require.Equal(t, http.StatusOK, recorder.Code)
	require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &page))
	require.Equal(t, 1, page.Pagination.TotalItems)
}

func TestContentHandlerErrors(t *testing.T) {
	handler := NewContentHandler(contentHTTPFixture(t))
	for _, test := range []struct {
		path   string
		status int
	}{
		{"/data/articles?page=0", 400},
		{"/data/articles?page=01", 400},
		{"/data/articles?page=1&page=2", 400},
		{"/data/articles?per_page=101", 400},
		{"/data/articles?tag=", 400},
		{"/data/articles?unknown=true", 400},
		{"/data/articles?sort=unknown", 400},
		{"/data/articles?order=unknown", 400},
		{"/data/articles?q=%zz", 400},
		{"/data/article/0", 400},
		{"/data/article/nope", 400},
		{"/data/article/99", 404},
		{"/data/articles/page/0.json", 400},
		{"/data/articles/page/99.json", 404},
		{"/data/articles/page/1", 404},
		{"/data/articles/page/1.json?tag=Go", 400},
		{"/images/missing.png", 404},
		{"/missing", 404},
	} {
		t.Run(test.path, func(t *testing.T) {
			recorder := httptest.NewRecorder()
			handler.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, test.path, nil))
			require.Equal(t, test.status, recorder.Code, recorder.Body.String())
			require.Contains(t, recorder.Header().Get("Content-Type"), "application/json")
			var body map[string]string
			require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &body))
			require.NotEmpty(t, body["error"])
		})
	}
}

func TestContentHandlerSnapshotRoutesAndHead(t *testing.T) {
	handler := NewContentHandler(contentHTTPFixture(t))
	for _, path := range []string{
		"/data/articles", "/data/article/1", "/data/article/1.json", "/data/articles/page/1.json",
		"/data/metadata.json", "/data/search/index.json", "/images/1/0.png",
	} {
		t.Run(path, func(t *testing.T) {
			get := httptest.NewRecorder()
			handler.ServeHTTP(get, httptest.NewRequest(http.MethodGet, path, nil))
			require.Equal(t, http.StatusOK, get.Code, get.Body.String())
			require.NotEmpty(t, get.Body.String())
			require.Equal(t, "no-cache", get.Header().Get("Cache-Control"))
			if strings.HasPrefix(path, "/images/") {
				require.Equal(t, "image/png", get.Header().Get("Content-Type"))
				require.Equal(t, "image bytes", get.Body.String())
			} else {
				require.True(t, json.Valid(get.Body.Bytes()))
			}
			head := httptest.NewRecorder()
			handler.ServeHTTP(head, httptest.NewRequest(http.MethodHead, path, nil))
			require.Equal(t, get.Code, head.Code)
			require.Equal(t, get.Header(), head.Header())
			require.Empty(t, head.Body.String())
		})
	}
}

func TestContentHandlerMethodAndEmptyPages(t *testing.T) {
	handler := NewContentHandler(contentHTTPFixture(t))
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, httptest.NewRequest(http.MethodPost, "/data/articles", nil))
	require.Equal(t, http.StatusMethodNotAllowed, response.Code)
	require.Equal(t, "GET, HEAD", response.Header().Get("Allow"))

	snapshot, err := NewContentSnapshot(nil, 20)
	require.NoError(t, err)
	handler = NewContentHandler(snapshot)
	response = httptest.NewRecorder()
	handler.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/data/articles?page=99", nil))
	require.Equal(t, http.StatusOK, response.Code)
	require.Contains(t, response.Body.String(), `"items":[]`)
	response = httptest.NewRecorder()
	handler.ServeHTTP(response, httptest.NewRequest(http.MethodHead, "/data/articles/page/1.json", nil))
	require.Equal(t, http.StatusNotFound, response.Code)
	require.Empty(t, response.Body.String())
}
