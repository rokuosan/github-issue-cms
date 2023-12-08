package cmd

import (
	"fmt"

	"github.com/rokuosan/github-issue-cms/internal"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var generateCmd = &cobra.Command{
	Use:   "generate",
	Short: "Generate articles from GitHub issues",
	Long: `Generate articles from GitHub issues.

This command will get issues from GitHub and create articles from them.
The articles will be saved in the "content" directory.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Initialize
		logger := internal.Logger

		username := viper.GetString("github.username")
		repository := viper.GetString("github.repository")
		if username == "" || repository == "" {
			logger.Error("Please set username and repository in gic.config.yaml")
			return
		}
		url := "https://github.com/" + username + "/" + repository
		logger.Info("Target Repository: " + url)

		// Get issues
		logger.Info("Getting issues...")
		issues := internal.GetIssues()
		if len(issues) == 0 {
			logger.Info("No issues found")
			return
		} else {
			logger.Infof("Found %d issues", len(issues))
		}

		// Create articles
		logger.Info("Creating articles...")
		for _, issue := range issues {
			article := internal.IssueToArticle(issue)
			if article != nil {
				internal.ExportArticle(article, fmt.Sprintf("%d", issue.GetID()))
			}
		}

	},
}

func init() {
	rootCmd.AddCommand(generateCmd)

	// GitHub Token
	generateCmd.Flags().StringVarP(&internal.GitHubToken, "token", "t", "", "GitHub API Token")
	generateCmd.MarkFlagRequired("token")

}
