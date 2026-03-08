package converter_v2

import (
	"context"
	"errors"
	"log/slog"
	"net/http"

	"github.com/google/go-github/v67/github"
)

var defaultGitHubClient = github.NewClient(http.DefaultClient)
var defaultLogger = func(method string) *slog.Logger {
	return slog.Default().With("package", "converter-v2", "method", method)
}

type Converter interface {
	HTTPClient() *http.Client
	GitHubClient() *github.Client

	WalkIssues(ctx context.Context, options WalkIssuesOptions, onPage func([]*github.Issue) error) error
}
type Option func(*converterImpl)

type converterImpl struct {
	Token string

	Username   string
	Repository string

	http   *http.Client
	github *github.Client
}

func NewConverter(options ...Option) Converter {
	c := &converterImpl{}
	for _, opt := range options {
		if opt == nil {
			continue
		}
		opt(c)
	}
	if c.github == nil {
		c.github = defaultGitHubClient
	}
	if c.http == nil {
		c.http = http.DefaultClient
	}
	return c
}

func WithToken(token string) Option {
	return func(c *converterImpl) {
		c.Token = token
	}
}

func WithRepository(owner, repository string) Option {
	return func(c *converterImpl) {
		c.Username = owner
		c.Repository = repository
	}
}

func WithHTTPClient(client *http.Client) Option {
	return func(c *converterImpl) {
		c.http = client
	}
}

func WithGitHubClient(client *github.Client) Option {
	return func(c *converterImpl) {
		c.github = client
	}
}

func (c *converterImpl) CheckRequirements() error {
	if c.Token == "" {
		return errors.New("missing token")
	}
	if c.Username == "" || c.Repository == "" {
		return errors.New("missing repository information")
	}
	return nil
}

func (c *converterImpl) HTTPClient() *http.Client {
	if c.http != nil {
		return c.http
	}
	return c.http
}

func (c *converterImpl) GitHubClient() *github.Client {
	if c.github != nil {
		return c.github
	}
	return defaultGitHubClient.WithAuthToken(c.Token)
}

type WalkIssuesOptions struct {
	IgnorePullRequests bool
	PerPage            int
}

func (c *converterImpl) WalkIssues(ctx context.Context, options WalkIssuesOptions, onPage func([]*github.Issue) error) error {
	if err := c.CheckRequirements(); err != nil {
		return err
	}

	if ctx == nil {
		ctx = context.Background()
	}
	perPage := options.PerPage
	if perPage <= 0 {
		perPage = 200
	}
	if onPage == nil {
		return nil
	}

	logger := defaultLogger("walkIssues")
	logger.Debug(
		"start walking issues",
		"owner", c.Username,
		"repo", c.Repository,
		"ignore_pull_requests", options.IgnorePullRequests,
		"per_page", perPage,
	)

	page := 1
	for {
		issues, resp, err := c.getIssuesByPage(ctx, page, perPage)
		if err != nil {
			logger.Debug("failed to fetch page", "page", page, "err", err.Error())
			return err
		}

		filtered := c.filterIssues(issues, options)
		logger.Debug("fetched page", "page", page, "count", len(filtered))
		if len(filtered) > 0 {
			if err := onPage(filtered); err != nil {
				return err
			}
		}

		if resp.NextPage <= 0 {
			return nil
		}
		page = resp.NextPage
	}
}

func (c *converterImpl) getIssuesByPage(ctx context.Context, page int, perPage int) ([]*github.Issue, *github.Response, error) {
	logger := defaultLogger("getIssuesByPage").With("page", page, "per_page", perPage)
	logger.Debug("requesting issues page")

	issues, resp, err := c.GitHubClient().Issues.ListByRepo(
		ctx,
		c.Username,
		c.Repository,
		&github.IssueListByRepoOptions{
			State: "all",
			ListOptions: github.ListOptions{
				PerPage: perPage,
				Page:    page,
			},
		},
	)
	if err != nil {
		logger.Debug("request failed", "err", err.Error())
		return nil, nil, err
	}
	logger.Debug("request succeeded", "count", len(issues), "next_page", resp.NextPage, "last_page", resp.LastPage)
	return issues, resp, nil
}

func (c *converterImpl) filterIssues(issues []*github.Issue, options WalkIssuesOptions) []*github.Issue {
	// Filter out pull requests by checking GitHub's issue field for pull request metadata.
	if !options.IgnorePullRequests {
		return issues
	}

	filtered := make([]*github.Issue, 0, len(issues))
	for _, issue := range issues {
		if issue.PullRequestLinks != nil {
			continue
		}
		filtered = append(filtered, issue)
	}
	return filtered
}
