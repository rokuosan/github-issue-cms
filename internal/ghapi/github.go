//go:generate mockgen -source=$GOFILE -package=${GOPACKAGE}_mock -destination=./mock/$GOFILE
package ghapi

import (
	"context"
	"time"

	"github.com/google/go-github/v67/github"
	"github.com/rokuosan/github-issue-cms/internal/util"
)

var defaultTimeout = 5 * time.Second

type Repository github.Repository
type Issue github.Issue

type API interface {
	Issues(ctx context.Context, input IssuesInput) ([]Issue, error)
}

type client struct {
	*github.Client
}

var _ API = &client{}

func NewClient(c *github.Client) API {
	return &client{Client: c}
}

type IssuesInput struct {
	Owner   string
	Name    string
	Page    *int
	PerPage *int
}

func (c *client) Issues(ctx context.Context, input IssuesInput) ([]Issue, error) {
	ctx, cancel := context.WithTimeout(ctx, defaultTimeout)
	defer cancel()

	page := util.DefaultOrItself(input.Page, 1)
	perPage := util.DefaultOrItself(input.PerPage, 100)

	issues, _, err := c.Client.Issues.ListByRepo(ctx, input.Owner, input.Name, &github.IssueListByRepoOptions{
		State: "all",
		ListOptions: github.ListOptions{
			Page:    page,
			PerPage: perPage,
		},
	})
	if err != nil {
		return nil, err
	}

	result := make([]Issue, len(issues))
	for i, issue := range issues {
		if issue == nil || issue.IsPullRequest() {
			continue
		}
		result[i] = Issue(*issue)
	}

	return result, nil
}
