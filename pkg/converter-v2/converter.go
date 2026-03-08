package converter_v2

import (
	"context"
	"errors"
	"net/http"

	"github.com/google/go-github/v67/github"
)

var defaultGitHubClient = github.NewClient(http.DefaultClient)

type Converter interface {
	HTTPClient() *http.Client
	GitHubClient() *github.Client

	GetIssues(ctx context.Context, options GetIssuesOptions) ([]*github.Issue, error)
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

type GetIssuesOptions struct {
	IgnorePullRequests bool
}

func (c *converterImpl) GetIssues(ctx context.Context, options GetIssuesOptions) ([]*github.Issue, error) {
	if err := c.CheckRequirements(); err != nil {
		return nil, err
	}

	if ctx == nil {
		ctx = context.Background()
	}
	return c.getIssues(ctx, options)
}

func (c *converterImpl) getIssues(ctx context.Context, options GetIssuesOptions) ([]*github.Issue, error) {
	all := []*github.Issue{}
	page := 1
	for {
		issues, resp, err := c.GitHubClient().Issues.ListByRepo(
			ctx,
			c.Username,
			c.Repository,
			&github.IssueListByRepoOptions{
				State: "all",
				ListOptions: github.ListOptions{
					PerPage: 200,
					Page:    page,
				},
			},
		)
		if err != nil {
			return nil, err
		}
		if options.IgnorePullRequests {
			filtered := make([]*github.Issue, 0, len(issues))
			for _, issue := range issues {
				if issue.PullRequestLinks != nil {
					continue
				}
				filtered = append(filtered, issue)
			}
			issues = filtered
		}
		all = append(all, issues...)
		if resp.NextPage <= 0 {
			return all, nil
		}
		page = resp.NextPage
	}
}
