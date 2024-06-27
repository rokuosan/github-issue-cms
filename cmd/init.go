package cmd

import (
	"github.com/rokuosan/github-issue-cms/pkg/config"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Generate config file.",
	Long:  `Generate config file named "gic.config.yaml" in the current directory.`,
	Run: func(cmd *cobra.Command, args []string) {
		config.Generate()
		config.Get()

		if username := cmd.Flag("username").Value.String(); username != "<YOUR_USERNAME>" {
			viper.Set("github.username", username)
		}
		if repository := cmd.Flag("repository").Value.String(); repository != "<YOUR_REPOSITORY>" {
			viper.Set("github.repository", repository)
		}

		if err := viper.WriteConfig(); err != nil {
			panic(err)
		}
	},
}

func init() {
	rootCmd.AddCommand(initCmd)

	initCmd.Flags().StringP("username", "u", "<YOUR_USERNAME>", "GitHub username")
	initCmd.Flags().StringP("repository", "r", "<YOUR_REPOSITORY>", "GitHub repository")
}
