package core

import (
	"context"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"io"
	"log/slog"
	"mime"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/rokuosan/github-issue-cms/pkg/config"
)

type AssetFetcher interface {
	Fetch(ctx context.Context, image *Image) (*ImageAsset, error)
}

// FileSystemArticleRepository stores articles in the filesystem.
type FileSystemArticleRepository struct {
	imageRepo AssetFetcher
	renderer  ArticleRenderer
	logger    *slog.Logger
}

// NewFileSystemArticleRepository creates a new FileSystemArticleRepository.
func NewFileSystemArticleRepository(imageRepo AssetFetcher) ArticleStore {
	return NewFileSystemArticleRepositoryWithLogger(imageRepo, nil)
}

// NewFileSystemArticleRepositoryWithLogger creates a new FileSystemArticleRepository with an injected logger.
func NewFileSystemArticleRepositoryWithLogger(imageRepo AssetFetcher, logger *slog.Logger) ArticleStore {
	return &FileSystemArticleRepository{
		imageRepo: imageRepo,
		renderer:  NewHugoArticleRenderer(),
		logger:    defaultLogger(logger),
	}
}

// Save stores an article in the filesystem.
func (r *FileSystemArticleRepository) Save(ctx context.Context, article *Article, conf config.Config) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	datetime, err := article.ParseDateTime()
	if err != nil {
		return fmt.Errorf("failed to parse datetime: %w", err)
	}

	articleDir, err := resolveArticleDirectory(conf, datetime)
	if err != nil {
		return err
	}
	if err := createDirectoryIfNotExist(articleDir); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", articleDir, err)
	}

	rendered := article.Clone()

	articlePath, err := resolveArticlePath(conf, datetime, articleDir)
	if err != nil {
		return err
	}

	imageDir, imageURLBase, err := resolveImageOutput(conf, datetime)
	if err != nil {
		return err
	}
	for _, image := range rendered.Images {
		if err := ctx.Err(); err != nil {
			return err
		}
		filename, err := r.saveImage(ctx, image, imageDir)
		if err != nil {
			r.logger.Error("Failed to download image", "url", image.URL, "error", err)
			continue
		}
		rendered.Content = strings.ReplaceAll(rendered.Content, image.URL, joinURLPath(imageURLBase, filename))
	}

	text, err := r.renderer.Render(rendered)
	if err != nil {
		return fmt.Errorf("failed to render article: %w", err)
	}
	if err := createFileAndWrite(articlePath, text); err != nil {
		return fmt.Errorf("failed to write file %s: %w", articlePath, err)
	}

	return nil
}

func resolveArticleDirectory(conf config.Config, datetime time.Time) (string, error) {
	dest := conf.Output.Articles.Directory
	if dest == "" {
		return "", fmt.Errorf("output articles directory is not set")
	}
	return filepath.Clean(config.CompileTimeTemplate(datetime, dest)), nil
}

func resolveArticlePath(conf config.Config, datetime time.Time, directory string) (string, error) {
	filename := conf.Output.Articles.Filename
	if filename == "" {
		return "", fmt.Errorf("output articles filename is not set")
	}
	filename = config.CompileTimeTemplate(datetime, filename)
	return filepath.Join(directory, filename), nil
}

func resolveImageOutput(conf config.Config, datetime time.Time) (string, string, error) {
	imageDir := conf.Output.Images.Directory
	if imageDir == "" {
		return "", "", fmt.Errorf("output images directory is not set")
	}
	imageURLBase := config.CompileTimeTemplate(datetime, conf.Output.Images.URL())
	return config.CompileTimeTemplate(datetime, imageDir), imageURLBase, nil
}

// createDirectoryIfNotExist creates the directory if it does not exist.
func createDirectoryIfNotExist(path string) error {
	return os.MkdirAll(path, 0o755)
}

// createFileAndWrite creates a file and writes content to it.
func createFileAndWrite(path string, content string) error {
	return os.WriteFile(path, []byte(content), 0o644)
}

func (r *FileSystemArticleRepository) saveImage(ctx context.Context, image *Image, imageDir string) (string, error) {
	if err := createDirectoryIfNotExist(imageDir); err != nil {
		return "", fmt.Errorf("failed to create directory %s: %w", imageDir, err)
	}

	asset, err := r.imageRepo.Fetch(ctx, image)
	if err != nil {
		return "", err
	}
	defer asset.Body.Close()

	filename := resolveImageBasename(image) + extensionFromContentType(asset.ContentType)
	if filepath.Ext(filename) == "" {
		filename += ".img"
	}

	fullPath := filepath.Join(imageDir, filename)
	file, err := os.Create(fullPath)
	if err != nil {
		return "", fmt.Errorf("failed to create file %s: %w", fullPath, err)
	}
	defer file.Close()

	if _, err := io.Copy(file, asset.Body); err != nil {
		return "", fmt.Errorf("failed to write image to %s: %w", fullPath, err)
	}

	return filename, nil
}

func resolveImageBasename(image *Image) string {
	sum := sha1.Sum([]byte(image.URL))
	return hex.EncodeToString(sum[:])[:12]
}

func joinURLPath(base, filename string) string {
	if base == "" {
		return filename
	}
	return path.Join(base, filename)
}

func extensionFromContentType(contentType string) string {
	mediaType, _, err := mime.ParseMediaType(contentType)
	if err != nil {
		mediaType = contentType
	}

	switch mediaType {
	case "image/png":
		return ".png"
	case "image/jpeg", "image/jpg":
		return ".jpg"
	case "image/gif":
		return ".gif"
	case "image/webp":
		return ".webp"
	case "image/svg+xml":
		return ".svg"
	case "image/bmp":
		return ".bmp"
	default:
		return ""
	}
}
