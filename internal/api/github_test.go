package api

import (
	"github.com/google/go-github/v67/github"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

const baseURLPath = "/api-v3"

func mockServer(t *testing.T) (*github.Client, *http.ServeMux, string) {
	t.Helper()

	mux := http.NewServeMux()

	m := http.NewServeMux()
	m.Handle(baseURLPath+"/", http.StripPrefix(baseURLPath, mux))
	server := httptest.NewServer(m)

	client := github.NewClient(nil)
	uri, err := url.Parse(server.URL + baseURLPath + "/")
	if err != nil {
		t.Error(err)
	}

	client.BaseURL = uri
	client.UploadURL = uri

	t.Cleanup(server.Close)
	return client, mux, server.URL
}

func mockIssueEndpoint(mux *http.ServeMux) {
	mux.HandleFunc("/repos/owner/repo/issues", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`[{"title":"test issue"}]`))
	})
}

func TestGithubAPI_GetIssues(t *testing.T) {
	client, mux, _ := mockServer(t)
	mockIssueEndpoint(mux)

	api := NewGitHubAPI(client, "owner", "repo")
	issues := api.GetIssues()

	if len(issues) != 1 {
		t.Errorf("Expected 1 issue, got %d", len(issues))
	}
	if issues[0].GetTitle() != "test issue" {
		t.Errorf("Expected title 'test issue', got %s", issues[0].GetTitle())
	}
}
