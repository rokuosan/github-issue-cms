package cmd

import (
	"os"

	"github.com/rokuosan/github-issue-cms/internal"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var rootCmd = &cobra.Command{
	Use:   "github-issue-cms",
	Short: "Generate articles from GitHub issues for Hugo",
	Run: func(cmd *cobra.Command, args []string) {
	},
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	internal.SetupLogger()

	// Read config file
	viper.SetConfigName("gic.config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	err := viper.ReadInConfig()
	if err != nil {
		panic(err)
	}

	// Debug
	rootCmd.PersistentFlags().BoolVarP(&internal.Debug, "debug", "d", false, "Debug mode")
}
