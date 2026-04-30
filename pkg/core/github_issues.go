package core

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/google/go-github/v85/github"
)

// GitHubIssueRepository retrieves issues via the GitHub API.
type GitHubIssueRepository struct {
	client *github.Client
	logger *slog.Logger
}

// NewGitHubIssueRepository creates a new GitHubIssueRepository.
func NewGitHubIssueRepository(token string) (IssueStore, error) {
	return NewGitHubIssueRepositoryWithLogger(token, nil)
}

// NewGitHubIssueRepositoryWithLogger creates a new GitHubIssueRepository with an injected logger.
func NewGitHubIssueRepositoryWithLogger(token string, logger *slog.Logger) (IssueStore, error) {
	if token == "" {
		return nil, fmt.Errorf("GitHub token is required")
	}

	client := github.NewClient(nil).WithAuthToken(token)
	if client == nil {
		return nil, fmt.Errorf("failed to create GitHub client")
	}

	return &GitHubIssueRepository{
		client: client,
		logger: defaultLogger(logger),
	}, nil
}

// ListIssues retrieves all issues from the specified repository.
func (r *GitHubIssueRepository) ListIssues(ctx context.Context, query IssueListQuery) ([]*github.Issue, error) {
	if query.Username == "" || query.Repository == "" {
		return nil, fmt.Errorf("username and repository name are required")
	}

	r.logger.Debug("Collecting Issues...")
	var (
		issues []*github.Issue
		rate   github.Rate
		page   = 1
	)

	for page != 0 {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		issuesAndPRs, resp, err := r.listIssuesPage(ctx, query, page)

		if err != nil {
			return nil, normalizeGitHubIssueError(err)
		}

		issues = append(issues, filterOutPullRequests(issuesAndPRs)...)
		page = resp.NextPage
		rate = resp.Rate
	}

	r.logger.Debug("Collected issues", "count", len(issues))
	r.logger.Debug("GitHub rate limit", "remaining", rate.Remaining, "limit", rate.Limit, "reset", rate.Reset)

	return issues, nil
}

func (r *GitHubIssueRepository) listIssuesPage(ctx context.Context, query IssueListQuery, page int) ([]*github.Issue, *github.Response, error) {
	return r.client.Issues.ListByRepo(
		ctx,
		query.Username,
		query.Repository,
		&github.IssueListByRepoOptions{
			State:  "all",
			Labels: query.Labels,
			ListOptions: github.ListOptions{
				PerPage: 100,
				Page:    page,
			},
		},
	)
}

func filterOutPullRequests(items []*github.Issue) []*github.Issue {
	issues := make([]*github.Issue, 0, len(items))
	for _, item := range items {
		if item.IsPullRequest() {
			continue
		}
		issues = append(issues, item)
	}
	return issues
}

func normalizeGitHubIssueError(err error) error {
	var ghErr *github.ErrorResponse
	if errors.As(err, &ghErr) && ghErr.Response != nil && ghErr.Response.StatusCode == http.StatusUnauthorized {
		return fmt.Errorf("invalid API token; please check your GitHub token")
	}
	return err
}
