package converter_v2

import (
	"context"
	"errors"
	"net/http"

	"github.com/google/go-github/v67/github"
)

var defaultGitHubClient = github.NewClient(http.DefaultClient)

type Converter interface{}
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

func (c *converterImpl) GetIssues(ctx context.Context) ([]*github.Issue, error) {
	if err := c.CheckRequirements(); err != nil {
		return nil, err
	}

	if ctx == nil {
		ctx = context.Background()
	}
	issues := []*github.Issue{}
	issues, err := c.getIssuesRecursively(ctx, issues, 1)
	if err != nil {
		return nil, err
	}

	return issues, nil
}

func (c *converterImpl) getIssuesRecursively(ctx context.Context, current []*github.Issue, page int) ([]*github.Issue, error) {
	issues, resp, err := c.getIssuesByPage(ctx, page)
	if err != nil {
		return nil, err
	}
	current = append(current, issues...)
	if resp.NextPage <= 0 {
		return current, nil
	}

	return c.getIssuesRecursively(ctx, current, resp.NextPage)
}

func (c *converterImpl) getIssuesByPage(ctx context.Context, page int) ([]*github.Issue, *github.Response, error) {
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
		return nil, nil, err
	}
	return issues, resp, nil
}
