package cmd

import (
	"github.com/rokuosan/github-issue-cms/pkg/converter"
	"log/slog"
	"strconv"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var githubToken string

var generateCmd = &cobra.Command{
	Use:   "generate",
	Short: "Generate articles from GitHub issues",
	Long: `Generate articles from GitHub issues.

This command will get issues from GitHub and create articles from them.
The articles will be saved in the "content" directory.`,
	Run: func(cmd *cobra.Command, args []string) {
		username := viper.GetString("github.username")
		repository := viper.GetString("github.repository")
		if username == "" || repository == "" {
			slog.Error("Please set username and repository in gic.config.yaml")
			return
		}
		url := "https://github.com/" + username + "/" + repository
		slog.Info("Target Repository: " + url)

		// Create articles
		c := converter.NewConverter(githubToken)
		issues := c.GetIssues()
		slog.Info("Found Issues: " + strconv.Itoa(len(issues)))
		slog.Info("Converting articles...")
		for _, issue := range issues {
			article, images := c.IssueToArticle(issue)
			article.ExportToMarkdown(article.Key)
			c.SaveImages(images)
		}

		slog.Info("Complete")
	},
}

func init() {
	rootCmd.AddCommand(generateCmd)

	// GitHub Token
	generateCmd.Flags().StringVarP(&githubToken, "token", "t", "", "GitHub API Token")
	_ = generateCmd.MarkFlagRequired("token")

}
