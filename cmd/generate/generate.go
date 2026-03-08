package generate

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/rokuosan/github-issue-cms/pkg/config"
	converter_v2 "github.com/rokuosan/github-issue-cms/pkg/converter-v2"

	"github.com/spf13/cobra"
)

func validateConfig(c config.Config) error {
	if c.GitHub.Username == "" {
		return fmt.Errorf("username is required")
	}
	if c.GitHub.Repository == "" {
		return fmt.Errorf("repository is required")
	}
	if c.Hugo.Directory.Articles == "" {
		return fmt.Errorf("hugo.directory.articles is required")
	}
	if c.Hugo.Directory.Images == "" {
		return fmt.Errorf("hugo.directory.images is required")
	}
	if c.Hugo.Filename.Articles == "" {
		return fmt.Errorf("hugo.filename.articles is required")
	}
	if c.Hugo.Filename.Images == "" {
		return fmt.Errorf("hugo.filename.images is required")
	}

	return nil
}

func handleRun(cmd *cobra.Command, args []string) {
	ctx := cmd.Context()

	conf := config.Get()
	if config.GitHubToken == "" {
		config.GitHubToken = os.Getenv("GITHUB_TOKEN")
	}
	if err := validateConfig(conf); err != nil {
		slog.Error("Invalid configuration", "error", err)
		return
	}

	conv := converter_v2.NewConverter(
		converter_v2.WithRepository(conf.GitHub.Username, conf.GitHub.Repository),
		converter_v2.WithToken(config.GitHubToken),
	)
	for issues, err := range conv.WalkIssues(ctx, converter_v2.WalkIssuesOptions{
		IgnorePullRequests: true,
		PerPage:            200,
	}) {
		if err != nil {
			slog.Error("Failed to walk issues", "error", err)
			return
		}
		for _, iss := range issues {
			art, err := converter_v2.NewIssueArticle(converter_v2.Markdown, iss)
			if err != nil {
				slog.Error("Failed to create article", "error", err)
				return
			}
			slog.Info("Generated article", "title", art.Title())

		export_file:
			dest, err := os.Create(prepareDestination(art, conf))
			if err != nil {
				if os.IsNotExist(err) {
					// If the directory does not exist, create it and retry
					dir := filepath.Dir(prepareDestination(art, conf))
					if err := os.MkdirAll(dir, 0755); err != nil {
						slog.Error("Failed to create directory", "error", err)
						return
					}
					goto export_file
				}
				slog.Error("Failed to create file", "error", err)
				return
			}
			defer dest.Close()

			if err := art.Export(dest); err != nil {
				slog.Error("Failed to export article", "error", err)
				return
			}
		}
	}
}

func prepareDestination(art converter_v2.Article, conf config.Config) string {
	filename := strings.ReplaceAll(conf.Hugo.Filename.Articles, "[:id]", art.ID())
	return filepath.Join(
		config.CompileTimeTemplate(art.Date(), conf.Hugo.Directory.Articles),
		config.CompileTimeTemplate(art.Date(), filename),
	)
}

func Command() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "gen",
		Short: "Generate articles from GitHub issues",
		Run:   handleRun,
	}
	cmd.Flags().StringVarP(&config.GitHubToken, "token", "t", "", "GitHub API Token")

	return cmd
}
