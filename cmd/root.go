package cmd

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/rokuosan/github-issue-cms/internal"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var rootCmd = &cobra.Command{
	Use:   "github-issue-cms",
	Short: "A brief description of your application",
	RunE: func(cmd *cobra.Command, args []string) error {
		// Print info
		username := viper.GetString("github.username")
		repository := viper.GetString("github.repository")
		URL := fmt.Sprintf("https://github.com/%s/%s", username, repository)
		slog.Info(fmt.Sprintf("Target Repository: %s\n", URL))

		// Download images
		internal.DownloadImage("https://github.com/rokuosan/tcardgen/blob/master/example/blogpost2.png", "test", 1)

		return nil
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
	rootCmd.PersistentFlags().StringVarP(&internal.Token, "token", "t", "", "GitHub API Token")
	rootCmd.MarkPersistentFlagRequired("token")
}
