package cmd

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/google/go-github/v56/github"
	"github.com/rokuosan/github-issue-cms/internal"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var rootCmd = &cobra.Command{
	Use:   "github-issue-cms",
	Short: "A brief description of your application",
	Run: func(cmd *cobra.Command, args []string) {
		// Print info
		username := viper.GetString("github.username")
		repository := viper.GetString("github.repository")
		if username == "" || repository == "" {
			slog.Error("Please set username and repository in gic.config")
			return
		}
		URL := fmt.Sprintf("https://github.com/%s/%s", username, repository)
		slog.Info(fmt.Sprintf("Target Repository: %s\n", URL))

		// Prepare Client
		slog.Info("Preparing GitHub Client...")
		internal.GitHubClient = github.NewClient(nil).WithAuthToken(internal.GitHubToken)
		slog.Info("Done\n")

		// Get issues

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

	// Flags
	rootCmd.PersistentFlags().StringVarP(&internal.GitHubToken, "token", "t", "", "GitHub API Token")
	rootCmd.MarkPersistentFlagRequired("token")
}
