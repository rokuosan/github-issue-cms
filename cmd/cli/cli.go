package cli

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/rokuosan/github-issue-cms/cmd/cli/subcommand"
	"github.com/spf13/cobra"
)

var verbosity int

// Version is the application version and can be overridden at build time with -ldflags.
var Version = "dev"

// NewRootCommand creates the root command.
func NewRootCommand() *cobra.Command {
	verbosity = 0
	configureLogger(0)

	rootCmd := &cobra.Command{
		Use:           "github-issue-cms",
		Short:         "Generate articles from GitHub issues for Hugo",
		SilenceErrors: true,
		SilenceUsage:  true,
		Long: `GitHub Issue-based headless CMS for Hugo.

This tool converts GitHub Issues into Hugo-compatible markdown articles
with frontmatter and downloads attached images.`,
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			configureLogger(verbosity)
		},
	}

	// Global flags.
	rootCmd.PersistentFlags().CountVarP(&verbosity, "verbose", "v", "Increase log verbosity (-v: info, -vv: debug)")

	// Register subcommands.
	rootCmd.AddCommand(subcommand.NewGenerateCommand())
	rootCmd.AddCommand(subcommand.NewInitCommand())
	rootCmd.AddCommand(subcommand.NewVersionCommand(&Version))

	return rootCmd
}

// Execute runs the CLI application.
func Execute() {
	rootCmd := NewRootCommand()
	if err := rootCmd.Execute(); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func configureLogger(verbosity int) {
	level := new(slog.LevelVar)
	level.Set(logLevelForVerbosity(verbosity))
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: level})))
}

func logLevelForVerbosity(verbosity int) slog.Level {
	switch {
	case verbosity >= 2:
		return slog.LevelDebug
	case verbosity == 1:
		return slog.LevelInfo
	default:
		return slog.LevelError
	}
}
