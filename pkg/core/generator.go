package core

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/google/go-github/v67/github"
	"github.com/rokuosan/github-issue-cms/pkg/config"
)

type IssueStore interface {
	ListIssues(ctx context.Context, username, repository string) ([]*github.Issue, error)
}

type ArticleStore interface {
	Save(ctx context.Context, article *Article, conf config.Config) error
}

// ArticleGenerator generates Hugo articles from GitHub issues.
type ArticleGenerator struct {
	issueRepo   IssueStore
	articleRepo ArticleStore
	service     *ArticleService
	config      config.Config
	logger      *slog.Logger
}

// NewArticleGenerator creates a new ArticleGenerator.
func NewArticleGenerator(conf config.Config, token string) (*ArticleGenerator, error) {
	return NewArticleGeneratorWithLogger(conf, token, nil)
}

// NewArticleGeneratorWithLogger creates a new ArticleGenerator with an injected logger.
func NewArticleGeneratorWithLogger(conf config.Config, token string, logger *slog.Logger) (*ArticleGenerator, error) {
	// Initialize repositories.
	issueRepo, err := NewGitHubIssueRepositoryWithLogger(token, logger)
	if err != nil {
		return nil, err
	}

	imageRepo := NewHTTPImageRepositoryWithLogger(token, logger)
	articleRepo := NewFileSystemArticleRepositoryWithLogger(imageRepo, logger)

	// Initialize services.
	articleService := NewArticleService(conf)

	return &ArticleGenerator{
		issueRepo:   issueRepo,
		articleRepo: articleRepo,
		service:     articleService,
		config:      conf,
		logger:      defaultLogger(logger),
	}, nil
}

// GetIssues retrieves all issues from the specified repository.
func (g *ArticleGenerator) GetIssues(ctx context.Context, username, repository string) ([]*github.Issue, error) {
	return g.issueRepo.ListIssues(ctx, username, repository)
}

// ConvertIssueToArticle converts an issue into an Article entity.
func (g *ArticleGenerator) ConvertIssueToArticle(issue *github.Issue) *Article {
	return g.service.ConvertIssueToArticle(issue)
}

// SaveArticle stores an Article in the filesystem.
func (g *ArticleGenerator) SaveArticle(ctx context.Context, article *Article) error {
	return g.articleRepo.Save(ctx, article, g.config)
}

// Generate fetches issues, converts them to articles, and saves them.
func (g *ArticleGenerator) Generate(ctx context.Context, username, repository string) (int, error) {
	// Fetch issues.
	issues, err := g.GetIssues(ctx, username, repository)
	if err != nil {
		return 0, err
	}

	g.logger.Info("Found issues", "count", len(issues))

	// Convert and save articles.
	successCount := 0
	var saveErr error
	for _, issue := range issues {
		if err := ctx.Err(); err != nil {
			return successCount, err
		}
		article := g.ConvertIssueToArticle(issue)
		if article == nil {
			continue
		}

		if err := g.SaveArticle(ctx, article); err != nil {
			g.logger.Error("Failed to save article", "issue", issue.GetNumber(), "error", err)
			saveErr = errors.Join(saveErr, fmt.Errorf("issue #%d: %w", issue.GetNumber(), err))
			continue
		}
		successCount++
	}

	if saveErr != nil {
		return successCount, fmt.Errorf("failed to save one or more articles: %w", saveErr)
	}

	return successCount, nil
}
