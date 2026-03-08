package main

import (
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/rokuosan/github-issue-cms/cmd"
)

func main() {
	// Measure the time it takes to run the program
	startTime := time.Now()
	defer func() {
		slog.Debug(fmt.Sprintf("Finished in %f seconds\n", time.Since(startTime).Seconds()))
	}()
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})))

	// Execute the root command
	cmd.Execute()
}
