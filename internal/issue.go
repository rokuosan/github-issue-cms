package internal

import (
	"context"
	"log/slog"

	"github.com/google/go-github/v56/github"
	"github.com/spf13/viper"
)

func GetIssues() {
	// Get Issues
	client := GitHubClient
	username := viper.GetString("github.username")
	repository := viper.GetString("github.repository")
	issues, _, err := client.Issues.ListByRepo(
		context.Background(),
		username,
		repository,
		&github.IssueListByRepoOptions{
			State: "all",
		},
	)

	// Check Rate Limits
	if _, ok := err.(*github.RateLimitError); ok {
		slog.Error("hit rate limit")
	}
	if _, ok := err.(*github.AbuseRateLimitError); ok {
		slog.Error("hit secondary rate limit")
	}
	if err != nil {
		panic(err)
	}

}
