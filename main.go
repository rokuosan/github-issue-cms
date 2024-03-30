package main

import (
	"fmt"
	"github.com/rokuosan/github-issue-cms/cmd"
	"log/slog"
	"time"
)

func main() {
	// Measure the time it takes to run the program
	startTime := time.Now()
	defer func() {
		slog.Debug(fmt.Sprintf("Finished in %f seconds\n", time.Since(startTime).Seconds()))
	}()

	// Execute the root command
	cmd.Execute()
}
