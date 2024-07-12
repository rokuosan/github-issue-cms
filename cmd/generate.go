package cmd

import (
	"log/slog"
	"strconv"

	"github.com/rokuosan/github-issue-cms/pkg/config"
	"github.com/rokuosan/github-issue-cms/pkg/converter"

	"github.com/spf13/cobra"
)

var generateCmd = &cobra.Command{
	Use:   "generate",
	Short: "Generate articles from GitHub issues",
	Long: `Generate articles from GitHub issues.

This command will get issues from GitHub and create articles from them.
The articles will be saved in the "content" directory.`,
	Run: func(cmd *cobra.Command, args []string) {
		config := config.Get()
		username := config.GitHub.Username
		repository := config.GitHub.Repository
		if username == "" || repository == "" {
			slog.Error("Please set username and repository in gic.config.yaml")
			return
		}
		url := config.GitHub.RepositoryURL()
		slog.Info("Target Repository: " + url)

		// Create articles
		c := converter.NewConverter()
		issues := c.GetIssues()
		slog.Info("Found Issues: " + strconv.Itoa(len(issues)))
		slog.Info("Converting articles...")
		for _, issue := range issues {
			article := c.IssueToArticle(issue)
			article.Export()
		}

		slog.Info("Complete")
	},
}

func init() {
	rootCmd.AddCommand(generateCmd)

	// GitHub Token
	generateCmd.Flags().StringVarP(&config.GitHubToken, "token", "t", "", "GitHub API Token")
	_ = generateCmd.MarkFlagRequired("token")

}
