package core

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"github.com/google/go-github/v86/github"
	"github.com/rokuosan/github-issue-cms/pkg/config"
	"github.com/stretchr/testify/require"
)

func TestContentSnapshotRemainsUsableAfterGitHubGoesOffline(t *testing.T) {
	issueRequests, imageRequests := 0, 0
	var source *httptest.Server
	source = httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/v3/repos/owner/repo/issues":
			issueRequests++
			require.Equal(t, "publish", r.URL.Query().Get("labels"))
			w.Header().Set("Content-Type", "application/json")
			if r.URL.Query().Get("page") == "1" {
				w.Header().Set("Link", fmt.Sprintf(`<%s/api/v3/repos/owner/repo/issues?page=2>; rel="next"`, source.URL))
				_, _ = fmt.Fprintf(w, `[
				  {"number":1,"title":"Published","state":"closed","created_at":"2026-01-01T00:00:00Z","labels":[{"name":"publish"},{"name":"Go"}],"body":"![image](%s/attachments/image)"},
				  {"number":2,"title":"Draft","state":"open","created_at":"2026-01-02T00:00:00Z"},
				  {"number":3,"title":"PR","state":"closed","created_at":"2026-01-03T00:00:00Z","pull_request":{}}
				]`, source.URL)
			} else {
				_, _ = fmt.Fprint(w, `[{"number":4,"title":"Another","state":"closed","created_at":"2026-01-04T00:00:00Z"}]`)
			}
		case "/attachments/image":
			imageRequests++
			w.Header().Set("Content-Type", "image/png")
			_, _ = w.Write([]byte("fixture-image"))
		default:
			http.NotFound(w, r)
		}
	}))
	defer source.Close()
	client, err := github.NewClient(source.Client()).WithEnterpriseURLs(source.URL, source.URL)
	require.NoError(t, err)
	store := &GitHubIssueRepository{client: client, logger: slog.Default()}
	conf := *config.NewConfig()
	conf.GitHub.Username, conf.GitHub.Repository = "owner", "repo"
	conf.GitHub.Labels = []string{"publish"}
	conf.Output.Images.Targets = []string{source.URL + "/attachments/"}
	assets := &HTTPImageRepository{client: source.Client(), logger: slog.Default()}
	snapshot, err := LoadContentSnapshot(context.Background(), store, assets, conf, 1, false)
	require.NoError(t, err)
	require.Len(t, snapshot.Articles, 2)
	require.Equal(t, 2, issueRequests)
	require.Equal(t, 1, imageRequests)
	source.Close()

	// Every API resource and static export must keep working after the source is gone.
	handler := NewContentHandler(snapshot)
	for _, path := range []string{
		"/data/articles?page=1&tag=Go", "/data/articles?q=Published", "/data/article/1",
		"/data/metadata.json", "/data/search/index.json", "/data/articles/page/2.json", "/images/1/0.png",
	} {
		response := httptest.NewRecorder()
		handler.ServeHTTP(response, httptest.NewRequest(http.MethodGet, path, nil))
		require.Equal(t, http.StatusOK, response.Code, path)
	}
	require.NoError(t, ExportContentSnapshot(context.Background(), snapshot, filepath.Join(t.TempDir(), "dist")))
}
