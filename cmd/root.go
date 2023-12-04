package cmd

import (
	"fmt"
	"os"

	"github.com/rokuosan/github-issue-cms/internal"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var rootCmd = &cobra.Command{
	Use:   "github-issue-cms",
	Short: "Generate articles from GitHub issues for Hugo",
	Run: func(cmd *cobra.Command, args []string) {
		// Initialize
		internal.Init()
		Logger := internal.Logger

		// Print info
		username := viper.GetString("github.username")
		repository := viper.GetString("github.repository")
		if username == "" || repository == "" {
			Logger.Error("Please set username and repository in gic.config.yaml")
			return
		}
		URL := fmt.Sprintf("https://github.com/%s/%s", username, repository)
		Logger.Info(fmt.Sprintf("Target Repository: %s", URL))

		// Get issues
		Logger.Info("Getting issues...")
		issues := internal.GetIssues()
		Logger.Infof("Found %d issues", len(issues))

		// Create articles
		Logger.Info("Creating articles...")
		for _, issue := range issues {
			article := internal.IssueToArticle(issue)
			if article != nil {
				internal.ExportArticle(article, fmt.Sprintf("%d", issue.GetID()))
			}
		}
	},
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	// Read config file
	viper.SetConfigName("gic.config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	err := viper.ReadInConfig()
	if err != nil {
		panic(err)
	}

	// Token
	rootCmd.PersistentFlags().StringVarP(&internal.GitHubToken, "token", "t", "", "GitHub API Token")
	rootCmd.MarkPersistentFlagRequired("token")

	// Debug
	rootCmd.PersistentFlags().BoolVarP(&internal.Debug, "debug", "d", false, "Debug mode")
}
