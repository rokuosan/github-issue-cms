package subcommand

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/rokuosan/github-issue-cms/pkg/config"
	"github.com/rokuosan/github-issue-cms/pkg/core"
	"github.com/rokuosan/github-issue-cms/pkg/ogimage"
	"github.com/spf13/cobra"
)

// NewGenerateCommand creates the generate subcommand.
func NewGenerateCommand() *cobra.Command {
	var (
		githubToken string
		withOGImage bool
	)

	cmd := &cobra.Command{
		Use:   "generate",
		Short: "Generate articles from GitHub issues",
		Long: `Generate articles from GitHub issues.

This command will get issues from GitHub and create articles from them.
The articles will be saved in the configured output directory structure
specified in gic.config.yaml.

Examples:
  # Generate articles with GitHub token
  github-issue-cms generate --token YOUR_GITHUB_TOKEN

  # Generate with info logging
  github-issue-cms -v generate --token YOUR_GITHUB_TOKEN

  # Generate with debug logging
  github-issue-cms -vv generate --token YOUR_GITHUB_TOKEN

  # Generate articles with OGP images
  github-issue-cms generate --token YOUR_GITHUB_TOKEN --with-ogimage`,

		RunE: func(cmd *cobra.Command, args []string) error {
			return runGenerate(cmd, githubToken, withOGImage)
		},
	}

	// Define flags.
	cmd.Flags().StringVarP(&githubToken, "token", "t", "", "GitHub API Token (required)")
	cmd.Flags().BoolVar(&withOGImage, "with-ogimage", false, "Generate OGP images alongside articles")
	_ = cmd.MarkFlagRequired("token")

	return cmd
}

func runGenerate(cmd *cobra.Command, githubToken string, withOGImage bool) error {
	// Load configuration.
	conf, err := config.Get()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}
	if conf.GitHub.Username == "" || conf.GitHub.Repository == "" {
		return fmt.Errorf("please set username and repository in gic.config.yaml; run 'github-issue-cms init' to create a config file")
	}

	url := conf.GitHub.RepositoryURL()
	slog.Info("Target Repository: " + url)

	// Create the article generator.
	generator, err := core.NewArticleGeneratorWithLogger(conf, githubToken, slog.Default())
	if err != nil {
		return fmt.Errorf("failed to create generator: %w", err)
	}

	// Set up OGP image generation hook if requested.
	if withOGImage {
		renderer, err := ogimage.NewRenderer("")
		if err != nil {
			return fmt.Errorf("failed to create OGP renderer: %w", err)
		}
		generator.SetOnArticleSaved(func(article *core.Article) error {
			return generateOGPForArticle(cmd, conf, renderer, article)
		})
		slog.Info("OGP image generation enabled (--with-ogimage)")
	}

	// Generate articles.
	slog.Info("Generating articles...")
	count, err := generator.Generate(cmd.Context(), conf.GitHub.Username, conf.GitHub.Repository)
	if err != nil {
		return fmt.Errorf("failed to generate articles: %w", err)
	}

	slog.Info("Complete: " + strconv.Itoa(count) + " articles generated")
	return nil
}

// generateOGPForArticle renders an OGP image for the given article and saves
// it as "ogp.jpeg" in the article's output directory.
func generateOGPForArticle(cmd *cobra.Command, conf config.Config, renderer *ogimage.Renderer, article *core.Article) error {
	data := articleToOGPData(article)

	jpeg, err := renderer.Render(cmd.Context(), data)
	if err != nil {
		return fmt.Errorf("render OGP: %w", err)
	}

	outputPath, err := resolveOGPArticlePath(conf, article)
	if err != nil {
		return fmt.Errorf("resolve OGP path: %w", err)
	}

	if err := os.MkdirAll(filepath.Dir(outputPath), 0o755); err != nil {
		return fmt.Errorf("create output directory: %w", err)
	}

	if err := os.WriteFile(outputPath, jpeg, 0o644); err != nil {
		return fmt.Errorf("write OGP image: %w", err)
	}

	slog.Debug("OGP image generated: " + outputPath)
	return nil
}

// resolveOGPArticlePath returns the path where the OGP image should be saved
// for an article. It places "ogp.jpeg" in the same directory as the article.
func resolveOGPArticlePath(conf config.Config, article *core.Article) (string, error) {
	datetime, err := article.ParseDateTime()
	if err != nil {
		datetime = time.Now()
	}

	articleDir := conf.Output.Articles.Directory
	if articleDir == "" {
		return "", fmt.Errorf("output articles directory is not configured")
	}
	articleDir = config.CompileTimeTemplate(datetime, articleDir)

	return filepath.Join(articleDir, "ogp.jpeg"), nil
}
