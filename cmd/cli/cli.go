package cli

import (
	"log/slog"
	"os"

	"github.com/rokuosan/github-issue-cms/cmd/cli/subcommand"
	"github.com/spf13/cobra"
)

var debug bool

// NewRootCommand creates the root command.
func NewRootCommand() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "github-issue-cms",
		Short: "Generate articles from GitHub issues for Hugo",
		Long: `GitHub Issue-based headless CMS for Hugo.

This tool converts GitHub Issues into Hugo-compatible markdown articles
with frontmatter and downloads attached images.`,
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			if debug {
				slog.SetLogLoggerLevel(slog.LevelDebug)
			}
		},
	}

	// Global flags.
	rootCmd.PersistentFlags().BoolVarP(&debug, "debug", "d", false, "Enable debug mode")

	// Register subcommands.
	rootCmd.AddCommand(subcommand.NewGenerateCommand())
	rootCmd.AddCommand(subcommand.NewInitCommand())

	return rootCmd
}

// Execute runs the CLI application.
func Execute() {
	rootCmd := NewRootCommand()
	if err := rootCmd.Execute(); err != nil {
		slog.Error("Command execution failed: " + err.Error())
		os.Exit(1)
	}
}
