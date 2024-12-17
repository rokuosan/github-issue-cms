package api

import (
	"encoding/json"
	"fmt"
	"github.com/google/go-github/v67/github"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

const baseURLPath = "/api-v3"

var issuesDB = map[string][]*github.Issue{}

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

func mockIssueEndpoint(mux *http.ServeMux, username, repository string) {
	mux.HandleFunc(
		fmt.Sprintf("/repos/%s/%s/issues", username, repository),
		func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)

			// Get issues from DB
			issues := issuesDB[username+"/"+repository]
			_ = json.NewEncoder(w).Encode(issues)
		},
	)
}

func TestGithubAPI_GetIssues(t *testing.T) {
	client, mux, _ := mockServer(t)
	username, repository := "owner", "repo"
	mockIssueEndpoint(mux, username, repository)

	issuesDB[username+"/"+repository] = []*github.Issue{
		{
			Title: github.String("Example Issue"),
		},
		{
			Title: github.String("Example Issue 2"),
		},
		{
			Title:            github.String("Example Pull Request"),
			PullRequestLinks: &github.PullRequestLinks{},
		},
	}

	api := NewGitHubAPI(client, username, repository)
	issues := api.GetIssues()

	if len(issues) != 2 {
		t.Errorf("Expected 2 issues, got %d", len(issues))
	}
}

func TestNewGitHubAPI(t *testing.T) {
	client := github.NewClient(nil)
	_ = NewGitHubAPI(client, "owner", "repo")
	// Test panic
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("The code did not panic")
		}
	}()

	_ = NewGitHubAPI(nil, "owner", "repo")
	_ = NewGitHubAPI(client, "", "repo")
	_ = NewGitHubAPI(client, "owner", "")
	_ = NewGitHubAPI(client, "", "")
	_ = NewGitHubAPI(nil, "", "")
}
