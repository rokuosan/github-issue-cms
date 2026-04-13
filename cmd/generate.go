package cmd

import (
	"fmt"
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
	RunE: func(cmd *cobra.Command, args []string) error {
		conf, err := config.Get()
		if err != nil {
			return err
		}
		if conf.GitHub.Username == "" || conf.GitHub.Repository == "" {
			return fmt.Errorf("please set username and repository in gic.config.yaml")
		}
		url := conf.GitHub.RepositoryURL()
		slog.Info("Target Repository: " + url)

		// Create articles
		c := converter.NewConverter(converter.Config{
			Config: conf,
			Token:  config.GitHubToken,
		}, config.GitHubToken)
		issues := c.GetIssues()
		slog.Info("Found Issues: " + strconv.Itoa(len(issues)))
		slog.Info("Converting articles...")
		for _, issue := range issues {
			article := c.IssueToArticle(issue)
			article.Export(conf)
		}

		slog.Info("Complete")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(generateCmd)

	// GitHub Token
	generateCmd.Flags().StringVarP(&config.GitHubToken, "token", "t", "", "GitHub API Token")
	_ = generateCmd.MarkFlagRequired("token")

}
