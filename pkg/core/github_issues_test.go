package core

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/go-github/v67/github"
	"github.com/stretchr/testify/assert"
)

func TestNewGitHubIssueRepository(t *testing.T) {
	tests := []struct {
		name    string
		token   string
		wantErr bool
	}{
		{
			name:    "有効なトークン",
			token:   "valid-token",
			wantErr: false,
		},
		{
			name:    "空のトークン",
			token:   "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo, err := NewGitHubIssueRepository(tt.token)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, repo)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, repo)
			}
		})
	}
}

func TestGitHubIssueRepository_ListIssues(t *testing.T) {
	t.Run("正常なIssue取得", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == "/api/v3/repos/testuser/testrepo/issues" {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte(`[
					{
						"id": 1,
						"number": 1,
						"title": "Test Issue 1",
						"body": "Test body 1",
						"state": "open",
						"user": {"login": "testuser"},
						"created_at": "2021-01-01T00:00:00Z"
					},
					{
						"id": 2,
						"number": 2,
						"title": "Test Issue 2",
						"body": "Test body 2",
						"state": "closed",
						"user": {"login": "testuser"},
						"created_at": "2021-01-02T00:00:00Z"
					}
				]`))
				return
			}
			w.WriteHeader(http.StatusNotFound)
		}))
		defer server.Close()

		client := github.NewClient(nil)
		client, err := client.WithEnterpriseURLs(server.URL, server.URL)
		assert.NoError(t, err)
		repo := &GitHubIssueRepository{client: client, logger: slog.Default()}

		issues, err := repo.ListIssues(context.Background(), "testuser", "testrepo")
		assert.NoError(t, err)
		assert.NotNil(t, issues)
		assertEqualCmp(t, []string{"Test Issue 1", "Test Issue 2"}, []string{issues[0].GetTitle(), issues[1].GetTitle()})
	})

	t.Run("空のユーザー名", func(t *testing.T) {
		client := github.NewClient(nil)
		repo := &GitHubIssueRepository{client: client, logger: slog.Default()}

		issues, err := repo.ListIssues(context.Background(), "", "testrepo")
		assert.Error(t, err)
		assert.Nil(t, issues)
		assert.Contains(t, err.Error(), "username and repository name are required")
	})

	t.Run("空のリポジトリ名", func(t *testing.T) {
		client := github.NewClient(nil)
		repo := &GitHubIssueRepository{client: client, logger: slog.Default()}

		issues, err := repo.ListIssues(context.Background(), "testuser", "")
		assert.Error(t, err)
		assert.Nil(t, issues)
		assert.Contains(t, err.Error(), "username and repository name are required")
	})
}

func TestGitHubIssueRepository_ListIssues_FiltersPRs(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/v3/repos/testuser/testrepo/issues" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			// Response containing both issues and pull requests.
			_, _ = w.Write([]byte(`[
				{
					"id": 1,
					"number": 1,
					"title": "Real Issue",
					"body": "This is an issue",
					"state": "open",
					"user": {"login": "testuser"},
					"created_at": "2021-01-01T00:00:00Z"
				},
				{
					"id": 2,
					"number": 2,
					"title": "Pull Request",
					"body": "This is a PR",
					"state": "open",
					"user": {"login": "testuser"},
					"created_at": "2021-01-02T00:00:00Z",
					"pull_request": {
						"url": "https://api.github.com/repos/testuser/testrepo/pulls/2"
					}
				},
				{
					"id": 3,
					"number": 3,
					"title": "Another Issue",
					"body": "This is another issue",
					"state": "closed",
					"user": {"login": "testuser"},
					"created_at": "2021-01-03T00:00:00Z"
				}
			]`))
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	client := github.NewClient(nil)
	client, err := client.WithEnterpriseURLs(server.URL, server.URL)
	assert.NoError(t, err)
	repo := &GitHubIssueRepository{client: client, logger: slog.Default()}

	issues, err := repo.ListIssues(context.Background(), "testuser", "testrepo")
	assert.NoError(t, err)
	assert.NotNil(t, issues)
	assertEqualCmp(t, []string{"Real Issue", "Another Issue"}, []string{issues[0].GetTitle(), issues[1].GetTitle()})
}

func TestGitHubIssueRepository_ListIssues_Pagination(t *testing.T) {
	requestCount := 0
	var serverURL string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v3/repos/testuser/testrepo/issues" {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		w.Header().Set("Content-Type", "application/json")

		// Change the response based on the page number.
		page := r.URL.Query().Get("page")
		requestCount++

		switch page {
		case "", "1":
			// First page: include a Link header pointing to the next page.
			// Use an absolute URL in the expected format.
			nextURL := serverURL + "/api/v3/repos/testuser/testrepo/issues?page=2"
			w.Header().Set("Link", `<`+nextURL+`>; rel="next"`)
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`[
				{"id": 1, "number": 1, "title": "Issue 1", "state": "open", "user": {"login": "testuser"}, "created_at": "2021-01-01T00:00:00Z"}
			]`))
		case "2":
			// Second page: last page, so no Link header.
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`[
				{"id": 2, "number": 2, "title": "Issue 2", "state": "open", "user": {"login": "testuser"}, "created_at": "2021-01-02T00:00:00Z"}
			]`))
		default:
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`[]`))
		}
	}))
	defer server.Close()
	serverURL = server.URL

	client := github.NewClient(nil)
	client, err := client.WithEnterpriseURLs(server.URL, server.URL)
	assert.NoError(t, err)
	repo := &GitHubIssueRepository{client: client, logger: slog.Default()}

	issues, err := repo.ListIssues(context.Background(), "testuser", "testrepo")
	assert.NoError(t, err)
	assert.NotNil(t, issues)
	assertEqualCmp(t, []string{"Issue 1", "Issue 2"}, []string{issues[0].GetTitle(), issues[1].GetTitle()})
	assert.GreaterOrEqual(t, requestCount, 2, "Should make multiple requests for pagination")
}

func TestGitHubIssueRepository_ListIssues_InvalidCredentials(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"message": "Bad credentials"}`))
	}))
	defer server.Close()

	client := github.NewClient(nil).WithAuthToken("invalid-token")
	client, err := client.WithEnterpriseURLs(server.URL, server.URL)
	assert.NoError(t, err)
	repo := &GitHubIssueRepository{client: client, logger: slog.Default()}

	issues, err := repo.ListIssues(context.Background(), "testuser", "testrepo")
	assert.Error(t, err)
	assert.Nil(t, issues)
	assert.Contains(t, err.Error(), "invalid API token")
}

func TestNormalizeGitHubIssueError(t *testing.T) {
	err := &github.ErrorResponse{
		Response: &http.Response{StatusCode: http.StatusUnauthorized},
		Message:  "Bad credentials",
	}

	got := normalizeGitHubIssueError(err)
	assert.EqualError(t, got, "invalid API token; please check your GitHub token")

	other := errors.New("some other error")
	assert.Same(t, other, normalizeGitHubIssueError(other))
}
