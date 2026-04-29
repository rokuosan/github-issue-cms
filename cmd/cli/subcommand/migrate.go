package subcommand

import (
	"fmt"
	"log/slog"

	"github.com/rokuosan/github-issue-cms/pkg/config"
	"github.com/spf13/cobra"
)

// NewMigrateCommand creates the migrate subcommand.
func NewMigrateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "migrate",
		Short: "Rewrite config to the latest schema",
		Long: `Rewrite gic.config.yaml to the latest schema.

This command reads the current configuration, keeps legacy keys readable,
and writes the canonical config format back to disk.

Examples:
  github-issue-cms migrate`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runMigrate()
		},
	}

	return cmd
}

func runMigrate() error {
	conf, err := config.Reload()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	hadLegacyHugo := conf.Hugo != nil
	if err := config.Write(conf); err != nil {
		return fmt.Errorf("failed to write migrated config: %w", err)
	}

	if hadLegacyHugo {
		slog.Info("Configuration migrated from 'hugo' to 'output': " + config.GetConfigPath())
		return nil
	}

	slog.Info("Configuration rewritten in the latest schema: " + config.GetConfigPath())
	return nil
}
