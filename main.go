package main

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/rokuosan/github-issue-cms/cmd/cli"
)

func main() {
	// Measure the time it takes to run the program
	startTime := time.Now()
	defer func() {
		slog.Debug(fmt.Sprintf("Finished in %f seconds\n", time.Since(startTime).Seconds()))
	}()

	// Execute the CLI
	cli.Execute()
}
