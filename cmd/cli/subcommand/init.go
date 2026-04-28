package subcommand

import (
	"fmt"
	"log/slog"

	"github.com/rokuosan/github-issue-cms/pkg/config"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// NewInitCommand creates the init subcommand.
func NewInitCommand() *cobra.Command {
	var (
		username   string
		repository string
	)

	cmd := &cobra.Command{
		Use:   "init",
		Short: "Generate config file",
		Long: `Generate config file named "gic.config.yaml" in the current directory.

This command creates a configuration file with default output settings.
You can specify GitHub username and repository using flags.

Examples:
  # Generate config with prompts
  github-issue-cms init

  # Generate config with username and repository
  github-issue-cms init --username yourname --repository yourrepo

  # Short form
  github-issue-cms init -u yourname -r yourrepo`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runInit(cmd, username, repository)
		},
	}

	// Define flags.
	cmd.Flags().StringVarP(&username, "username", "u", "", "GitHub username")
	cmd.Flags().StringVarP(&repository, "repository", "r", "", "GitHub repository name")

	return cmd
}

func runInit(cmd *cobra.Command, username, repository string) error {
	slog.Info("Generating configuration file...")

	// Generate the config file.
	if err := config.Generate(); err != nil {
		return fmt.Errorf("failed to generate config file: %w", err)
	}

	// Load configuration.
	conf, err := config.Reload()
	if err != nil {
		return fmt.Errorf("failed to load config file: %w", err)
	}

	// Override config values when flags are provided.
	if username != "" {
		viper.Set("github.username", username)
		conf.GitHub.Username = username
	}
	if repository != "" {
		viper.Set("github.repository", repository)
		conf.GitHub.Repository = repository
	}

	// Save configuration.
	if err := viper.WriteConfig(); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	configPath := config.GetConfigPath()
	slog.Info("Configuration file created: " + configPath)

	// Show the resulting configuration values.
	if conf.GitHub.Username != "" && conf.GitHub.Username != "<YOUR_USERNAME>" {
		slog.Info("GitHub Username: " + conf.GitHub.Username)
	} else {
		slog.Warn("Please update 'github.username' in " + configPath)
	}

	if conf.GitHub.Repository != "" && conf.GitHub.Repository != "<YOUR_REPOSITORY>" {
		slog.Info("GitHub Repository: " + conf.GitHub.Repository)
	} else {
		slog.Warn("Please update 'github.repository' in " + configPath)
	}

	slog.Info("You can now run 'github-issue-cms generate --token YOUR_TOKEN'")

	return nil
}
