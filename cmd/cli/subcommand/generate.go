package subcommand

import (
	"fmt"
	"log/slog"
	"strconv"

	"github.com/rokuosan/github-issue-cms/pkg/config"
	"github.com/rokuosan/github-issue-cms/pkg/core"
	"github.com/spf13/cobra"
)

// NewGenerateCommand creates the generate subcommand.
func NewGenerateCommand() *cobra.Command {
	var githubToken string

	cmd := &cobra.Command{
		Use:   "generate",
		Short: "Generate articles from GitHub issues",
		Long: `Generate articles from GitHub issues.

This command will get issues from GitHub and create articles from them.
The articles will be saved in the Hugo-compatible directory structure
specified in gic.config.yaml.

Examples:
  # Generate articles with GitHub token
  github-issue-cms generate --token YOUR_GITHUB_TOKEN

  # Generate with info logging
  github-issue-cms -v generate --token YOUR_GITHUB_TOKEN

  # Generate with debug logging
  github-issue-cms -vv generate --token YOUR_GITHUB_TOKEN`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runGenerate(cmd, githubToken)
		},
	}

	// Define flags.
	cmd.Flags().StringVarP(&githubToken, "token", "t", "", "GitHub API Token (required)")
	_ = cmd.MarkFlagRequired("token")

	return cmd
}

func runGenerate(cmd *cobra.Command, githubToken string) error {
	// Load configuration.
	conf, err := config.Get()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}
	if conf.GitHub.Username == "" || conf.GitHub.Repository == "" {
		slog.Info("Run 'github-issue-cms init' to create a config file")
		return fmt.Errorf("please set username and repository in gic.config.yaml")
	}

	url := conf.GitHub.RepositoryURL()
	slog.Info("Target Repository: " + url)

	// Create the article generator.
	generator, err := core.NewArticleGeneratorWithLogger(conf, githubToken, slog.Default())
	if err != nil {
		return fmt.Errorf("failed to create generator: %w", err)
	}

	// Generate articles.
	slog.Info("Generating articles...")
	count, err := generator.Generate(cmd.Context(), conf.GitHub.Username, conf.GitHub.Repository)
	if err != nil {
		return fmt.Errorf("failed to generate articles: %w", err)
	}

	slog.Info("Complete: " + strconv.Itoa(count) + " articles generated")
	return nil
}
