package subcommand

import (
	"log/slog"
	"strconv"

	"github.com/rokuosan/github-issue-cms/pkg/config"
	"github.com/rokuosan/github-issue-cms/pkg/core"
	"github.com/spf13/cobra"
)

var githubToken string

// NewGenerateCommand creates the generate subcommand.
func NewGenerateCommand() *cobra.Command {
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

  # Generate with debug logging
  github-issue-cms generate --token YOUR_GITHUB_TOKEN --debug`,
		RunE: runGenerate,
	}

	// Define flags.
	cmd.Flags().StringVarP(&githubToken, "token", "t", "", "GitHub API Token (required)")
	_ = cmd.MarkFlagRequired("token")

	return cmd
}

func runGenerate(cmd *cobra.Command, args []string) error {
	// Load configuration.
	conf, err := config.Get()
	if err != nil {
		slog.Error("Failed to load config: " + err.Error())
		return err
	}
	if conf.GitHub.Username == "" || conf.GitHub.Repository == "" {
		slog.Error("Please set username and repository in gic.config.yaml")
		slog.Info("Run 'github-issue-cms init' to create a config file")
		return nil
	}

	url := conf.GitHub.RepositoryURL()
	slog.Info("Target Repository: " + url)

	// Create the article generator.
	generator, err := core.NewArticleGeneratorWithLogger(conf, githubToken, slog.Default())
	if err != nil {
		slog.Error("Failed to create generator: " + err.Error())
		return err
	}

	// Generate articles.
	slog.Info("Generating articles...")
	count, err := generator.Generate(cmd.Context(), conf.GitHub.Username, conf.GitHub.Repository)
	if err != nil {
		slog.Error("Failed to generate articles: " + err.Error())
		return err
	}

	slog.Info("Complete: " + strconv.Itoa(count) + " articles generated")
	return nil
}
