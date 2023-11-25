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
	Short: "A brief description of your application",
	Run: func(cmd *cobra.Command, args []string) {
		// Initialize
		internal.Init()
		logger := internal.Logger

		// Print info
		username := viper.GetString("github.username")
		repository := viper.GetString("github.repository")
		if username == "" || repository == "" {
			logger.Error("Please set username and repository in gic.config")
			return
		}
		URL := fmt.Sprintf("https://github.com/%s/%s", username, repository)
		logger.Info(fmt.Sprintf("Target Repository: %s\n", URL))

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

	// Token
	rootCmd.PersistentFlags().StringVarP(&internal.GitHubToken, "token", "t", "", "GitHub API Token")
	rootCmd.MarkPersistentFlagRequired("token")

	// Debug
	rootCmd.PersistentFlags().BoolVarP(&internal.Debug, "debug", "d", false, "Debug mode")
}
