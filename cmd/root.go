package cmd

import (
	"log/slog"
	"os"

	"github.com/spf13/cobra"
)

var debug bool

var rootCmd = &cobra.Command{
	Use:   "github-issue-cms",
	Short: "Generate articles from GitHub issues for Hugo",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		if debug {
			slog.SetLogLoggerLevel(slog.LevelDebug)
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
	// Debug
	rootCmd.PersistentFlags().BoolVarP(&debug, "debug", "d", false, "Debug mode")
}
