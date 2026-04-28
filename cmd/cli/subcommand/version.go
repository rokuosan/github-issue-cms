package subcommand

import (
	"fmt"

	"github.com/spf13/cobra"
)

// NewVersionCommand creates the version subcommand.
func NewVersionCommand(version *string) *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print the application version",
		Long: `Print the application version.

Examples:
  github-issue-cms version`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if version == nil {
				return fmt.Errorf("version is not set")
			}

			_, err := fmt.Fprintln(cmd.OutOrStdout(), *version)
			return err
		},
	}
}
