package core

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestContentExportMatchesHTTP(t *testing.T) {
	snapshot, err := NewContentSnapshot([]ContentArticle{
		{ID: 1, Title: "First", Date: "2026-01-01", Content: strings.Repeat("日", 210), Tags: []string{"Go"}},
		{ID: 2, Title: "Second", Date: "2026-01-02", Content: "Body", Metadata: map[string]any{"custom": "value"}},
	}, 1)
	require.NoError(t, err)
	snapshot.Assets["/images/1/0.png"] = ContentAsset{Data: []byte("image"), ContentType: "image/png"}
	output := filepath.Join(t.TempDir(), "dist")
	require.NoError(t, ExportContentSnapshot(context.Background(), snapshot, output))
	for _, path := range []string{
		"/data/metadata.json", "/data/articles/page/1.json", "/data/articles/page/2.json",
		"/data/article/1.json", "/data/article/2.json", "/data/search/index.json", "/images/1/0.png",
	} {
		t.Run(path, func(t *testing.T) {
			data, err := os.ReadFile(filepath.Join(output, filepath.FromSlash(strings.TrimPrefix(path, "/"))))
			require.NoError(t, err)
			response := httptest.NewRecorder()
			NewContentHandler(snapshot).ServeHTTP(response, httptest.NewRequest(http.MethodGet, path, nil))
			require.Equal(t, http.StatusOK, response.Code)
			require.Equal(t, string(data), response.Body.String())
		})
	}
	index := snapshot.SearchIndex()
	require.Len(t, index, 2)
	require.Equal(t, strings.Repeat("日", 200)+"…", index[1].Summary)
	index[1].Tags[0] = "changed"
	require.Equal(t, []string{"Go"}, snapshot.Articles[1].Tags)
	data, err := os.ReadFile(filepath.Join(output, "data", "metadata.json"))
	require.NoError(t, err)
	var metadata ContentMetadata
	require.NoError(t, json.Unmarshal(data, &metadata))
	require.Equal(t, []string{"/data/articles/page/1.json", "/data/articles/page/2.json"}, metadata.Pages)
}

func TestContentExportReplacesOnlyManagedOutput(t *testing.T) {
	ctx := context.Background()
	snapshot, err := NewContentSnapshot([]ContentArticle{{ID: 1, Date: "2026-01-01"}}, 20)
	require.NoError(t, err)
	output := filepath.Join(t.TempDir(), "dist")
	require.NoError(t, os.MkdirAll(output, 0o755))
	userFile := filepath.Join(output, "keep.txt")
	require.NoError(t, os.WriteFile(userFile, []byte("user content"), 0o644))
	require.ErrorContains(t, ExportContentSnapshot(ctx, snapshot, output), "non-export")
	data, err := os.ReadFile(userFile)
	require.NoError(t, err)
	require.Equal(t, "user content", string(data))
	require.NoError(t, os.Remove(userFile))
	require.NoError(t, ExportContentSnapshot(ctx, snapshot, output))
	empty, err := NewContentSnapshot(nil, 20)
	require.NoError(t, err)
	require.NoError(t, ExportContentSnapshot(ctx, empty, output))
	_, err = os.Stat(filepath.Join(output, "data", "article", "1.json"))
	require.True(t, os.IsNotExist(err), "removed articles must not remain public")
	data, err = os.ReadFile(filepath.Join(output, "data", "search", "index.json"))
	require.NoError(t, err)
	require.JSONEq(t, "[]", string(data))
	require.Empty(t, empty.Metadata().Pages)
	_, err = os.Stat(filepath.Join(output, "data", "articles", "page", "1.json"))
	require.True(t, os.IsNotExist(err))
}

func TestContentExportFailurePreservesPreviousExport(t *testing.T) {
	snapshot, err := NewContentSnapshot([]ContentArticle{{ID: 1, Date: "2026-01-01"}}, 20)
	require.NoError(t, err)
	parent := t.TempDir()
	output := filepath.Join(parent, "dist")
	require.NoError(t, ExportContentSnapshot(context.Background(), snapshot, output))
	path := filepath.Join(output, "data", "article", "1.json")
	before, err := os.ReadFile(path)
	require.NoError(t, err)
	snapshot.Articles[0].Metadata["invalid"] = make(chan int)
	require.Error(t, ExportContentSnapshot(context.Background(), snapshot, output))
	after, err := os.ReadFile(path)
	require.NoError(t, err)
	require.Equal(t, before, after)
	entries, err := os.ReadDir(parent)
	require.NoError(t, err)
	require.Len(t, entries, 1, "staging directories should be cleaned up")
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	require.ErrorIs(t, ExportContentSnapshot(ctx, snapshot, output), context.Canceled)
}

func TestContentExportRejectsUnsafePaths(t *testing.T) {
	snapshot, err := NewContentSnapshot(nil, 20)
	require.NoError(t, err)
	for _, path := range []string{"/images/../../escape", "/images/a/../escape", "/images/a\\escape", "/other/file"} {
		t.Run(path, func(t *testing.T) {
			snapshot.Assets = map[string]ContentAsset{path: {Data: []byte("no")}}
			output := filepath.Join(t.TempDir(), "dist")
			require.ErrorContains(t, ExportContentSnapshot(context.Background(), snapshot, output), "invalid asset path")
			_, err := os.Stat(output)
			require.True(t, os.IsNotExist(err))
		})
	}
	snapshot.Assets = nil
	parent := t.TempDir()
	target := filepath.Join(parent, "target")
	require.NoError(t, os.Mkdir(target, 0o755))
	link := filepath.Join(parent, "link")
	if err := os.Symlink(target, link); err != nil {
		t.Skipf("symlinks unavailable: %v", err)
	}
	require.ErrorContains(t, ExportContentSnapshot(context.Background(), snapshot, link), "symlink")
}
