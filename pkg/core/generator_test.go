package core

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/go-github/v67/github"
	"github.com/rokuosan/github-issue-cms/pkg/config"
	"github.com/stretchr/testify/assert"
)

func TestNewArticleGenerator(t *testing.T) {
	conf := *config.NewConfig()

	t.Run("有効なトークン", func(t *testing.T) {
		gen, err := NewArticleGenerator(conf, "valid-token")
		assert.NoError(t, err)
		assert.NotNil(t, gen)
	})

	t.Run("空のトークン", func(t *testing.T) {
		gen, err := NewArticleGenerator(conf, "")
		assert.Error(t, err)
		assert.Nil(t, gen)
	})
}

func TestArticleGenerator_ConvertIssueToArticle(t *testing.T) {
	conf := *config.NewConfig()
	conf.Hugo.Url.Images = "/images"
	conf.Hugo.Filename.Images = "[:id].png"

	gen, err := NewArticleGenerator(conf, "test-token")
	assert.NoError(t, err)

	issue := &github.Issue{
		Title:     generatorStringPtr("Test Issue"),
		Body:      generatorStringPtr("Test content"),
		CreatedAt: generatorParseTime("2021-01-01T00:00:00Z"),
		User:      &github.User{Login: generatorStringPtr("testuser")},
		State:     generatorStringPtr("closed"),
		Labels:    []*github.Label{},
	}

	article := gen.ConvertIssueToArticle(issue)
	assert.NotNil(t, article)
	assertEqualCmp(t, "Test Issue", article.Title)
	assertEqualCmp(t, "testuser", article.Author)
	assertEqualCmp(t, "Test content\n", article.Content)
	assert.False(t, article.Draft)
}

func TestArticleGenerator_ConvertPullRequest(t *testing.T) {
	conf := *config.NewConfig()
	gen, err := NewArticleGenerator(conf, "test-token")
	assert.NoError(t, err)

	pr := &github.Issue{
		Title:            generatorStringPtr("PR Title"),
		Body:             generatorStringPtr("PR content"),
		CreatedAt:        generatorParseTime("2021-01-01T00:00:00Z"),
		User:             &github.User{Login: generatorStringPtr("user")},
		State:            generatorStringPtr("open"),
		Labels:           []*github.Label{},
		PullRequestLinks: &github.PullRequestLinks{},
	}

	article := gen.ConvertIssueToArticle(pr)
	assert.Nil(t, article, "Pull Request should return nil")
}

func TestArticleGenerator_SaveArticle(t *testing.T) {
	tempDir := t.TempDir()
	conf := *config.NewConfig()
	conf.Hugo.Directory.Articles = tempDir + "/articles"
	conf.Hugo.Directory.Images = tempDir + "/images"
	conf.Hugo.Filename.Articles = "%Y-%m-%d.md"
	conf.Hugo.Filename.Images = "[:id].png"

	gen, err := NewArticleGenerator(conf, "test-token")
	assert.NoError(t, err)

	issue := &github.Issue{
		Title:     generatorStringPtr("Test Article"),
		Body:      generatorStringPtr("Content"),
		CreatedAt: generatorParseTime("2021-01-01T00:00:00Z"),
		User:      &github.User{Login: generatorStringPtr("testuser")},
		State:     generatorStringPtr("closed"),
		Labels:    []*github.Label{},
	}

	article := gen.ConvertIssueToArticle(issue)
	assert.NotNil(t, article)

	err = gen.SaveArticle(context.Background(), article)
	assert.NoError(t, err)
}

func TestArticleGenerator_GetIssues(t *testing.T) {
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

	conf := *config.NewConfig()

	// Create a mock IssueRepository.
	issueRepo := newMockIssueRepository(server.URL)
	imageRepo := NewHTTPImageRepository("")
	articleRepo := NewFileSystemArticleRepository(imageRepo)
	articleService := NewArticleService(conf)

	gen := &ArticleGenerator{
		issueRepo:   issueRepo,
		articleRepo: articleRepo,
		service:     articleService,
		config:      conf,
		logger:      slog.Default(),
	}

	issues, err := gen.GetIssues(context.Background(), "testuser", "testrepo")
	assert.NoError(t, err)
	assert.NotNil(t, issues)
	assertEqualCmp(t, []string{"Test Issue 1", "Test Issue 2"}, []string{issues[0].GetTitle(), issues[1].GetTitle()})
}

func TestArticleGenerator_Generate(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/v3/repos/testuser/testrepo/issues" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`[
				{
					"id": 1,
					"number": 1,
					"title": "Test Issue",
					"body": "Test content",
					"state": "closed",
					"user": {"login": "testuser"},
					"created_at": "2021-01-01T00:00:00Z",
					"labels": []
				}
			]`))
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	tempDir := t.TempDir()
	conf := *config.NewConfig()
	conf.Hugo.Directory.Articles = tempDir + "/articles"
	conf.Hugo.Directory.Images = tempDir + "/images"
	conf.Hugo.Filename.Articles = "%Y-%m-%d.md"
	conf.Hugo.Filename.Images = "[:id].png"

	// Create a mock IssueRepository.
	issueRepo := newMockIssueRepository(server.URL)
	imageRepo := NewHTTPImageRepository("")
	articleRepo := NewFileSystemArticleRepository(imageRepo)
	articleService := NewArticleService(conf)

	gen := &ArticleGenerator{
		issueRepo:   issueRepo,
		articleRepo: articleRepo,
		service:     articleService,
		config:      conf,
		logger:      slog.Default(),
	}

	count, err := gen.Generate(context.Background(), "testuser", "testrepo")
	assert.NoError(t, err)
	assertEqualCmp(t, 1, count)
}

// Helper functions.

func generatorStringPtr(s string) *string {
	return &s
}

func generatorParseTime(s string) *github.Timestamp {
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		panic(err)
	}
	return &github.Timestamp{Time: t}
}

// newMockIssueRepository creates an IssueRepository for the mock server.
func newMockIssueRepository(serverURL string) IssueStore {
	client := github.NewClient(nil)
	client, _ = client.WithEnterpriseURLs(serverURL, serverURL)

	// Use mockIssueRepository.
	return &mockIssueRepository{client: client}
}

// mockIssueRepository is a test mock implementation.
type mockIssueRepository struct {
	client *github.Client
}

func (m *mockIssueRepository) ListIssues(ctx context.Context, username, repository string) ([]*github.Issue, error) {
	if username == "" || repository == "" {
		return nil, fmt.Errorf("username and repository name are required")
	}

	opts := &github.IssueListByRepoOptions{
		State: "all",
		ListOptions: github.ListOptions{
			PerPage: 100,
		},
	}

	var allIssues []*github.Issue
	for {
		issues, resp, err := m.client.Issues.ListByRepo(ctx, username, repository, opts)
		if err != nil {
			return nil, err
		}

		// Filter out pull requests.
		for _, issue := range issues {
			if issue.PullRequestLinks == nil {
				allIssues = append(allIssues, issue)
			}
		}

		if resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}

	return allIssues, nil
}
