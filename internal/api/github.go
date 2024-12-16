package api

import (
	"context"
	"fmt"
	"github.com/google/go-github/v67/github"
	"log/slog"
)

const (
	stateAll       = "all"
	defaultPerPage = 200
)

type GitHub interface {
	GetIssues() []*github.Issue
}

type githubAPI struct {
	*github.Client
	Username   string
	Repository string
}

func NewGitHubAPI(client *github.Client, username, repository string) GitHub {
	if client == nil {
		panic("client is nil")
	}
	if username == "" || repository == "" {
		panic("username and repository is must be set")
	}

	return &githubAPI{
		Client:     client,
		Username:   username,
		Repository: repository,
	}
}

func (g *githubAPI) GetIssues() []*github.Issue {
	var issues []*github.Issue
	var limit github.Rate
	next := 1
	for next != 0 {
		slog.Debug(fmt.Sprintf("Attempt to get issues from GitHub: page %d (per %d)", next, defaultPerPage))
		all, resp, err := g.Issues.ListByRepo(
			context.Background(),
			g.Username,
			g.Repository,
			&github.IssueListByRepoOptions{
				State: stateAll,
				ListOptions: github.ListOptions{
					PerPage: defaultPerPage,
					Page:    next,
				},
			},
		)
		if err != nil {
			panic(err)
		}

		for _, item := range all {
			if item.IsPullRequest() {
				slog.Debug(fmt.Sprintf("Skip PR: %s", item.GetTitle()))
				continue
			}
			issues = append(issues, item)
		}

		next = resp.NextPage
		limit = resp.Rate
	}
	slog.Debug(fmt.Sprintf("Found issues: %d", len(issues)))
	slog.Debug(fmt.Sprintf("Remaining Rate Limit: %d/%d (Reset: %s)", limit.Remaining, limit.Limit, limit.Reset))
	return issues
}
