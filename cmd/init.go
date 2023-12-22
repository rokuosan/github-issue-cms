package cmd

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Generate config file.",
	Long:  `Generate config file named "gic.config.yaml" in the current directory.`,
	Run: func(cmd *cobra.Command, args []string) {
		username := cmd.Flag("username").Value.String()
		repository := cmd.Flag("repository").Value.String()

		viper.Set("github.username", username)
		viper.Set("github.repository", repository)

		err := viper.WriteConfigAs("gic.config.yaml")
		if err != nil {
			panic(err)
		}
	},
}

func init() {
	rootCmd.AddCommand(initCmd)

	initCmd.Flags().StringP("username", "u", "<YOUR_USERNAME>", "GitHub username")
	initCmd.Flags().StringP("repository", "r", "<YOUR_REPOSITORY>", "GitHub repository")
}
