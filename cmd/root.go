package cmd

import (
	"log/slog"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
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
	// Read config file
	viper.SetConfigName("gic.config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	err := viper.ReadInConfig()
	if err != nil {
		panic(err)
	}

	// Debug
	rootCmd.PersistentFlags().BoolVarP(&debug, "debug", "d", false, "Debug mode")
}
