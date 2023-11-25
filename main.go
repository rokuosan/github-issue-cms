package main

import (
	"fmt"
	"time"

	"github.com/rokuosan/github-issue-cms/cmd"
	"github.com/rokuosan/github-issue-cms/internal"
)

func main() {
	// Measure the time it takes to run the program
	startTime := time.Now()
	defer func() {
		internal.Logger.Info(fmt.Sprintf("Finished in %f seconds\n", time.Since(startTime).Seconds()))
	}()

	// Execute the root command
	cmd.Execute()
}
