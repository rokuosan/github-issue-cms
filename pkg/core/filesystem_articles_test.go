package core

import (
	"context"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/rokuosan/github-issue-cms/pkg/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeImageRepository struct {
	contentType string
	body        string
}

func (r *fakeImageRepository) Fetch(ctx context.Context, image *Image) (*ImageAsset, error) {
	return &ImageAsset{
		Body:        io.NopCloser(strings.NewReader(r.body)),
		ContentType: r.contentType,
	}, nil
}

func TestFileSystemArticleRepository_Save_RewritesImageURLs(t *testing.T) {
	tempDir := t.TempDir()

	conf := *config.NewConfig()
	conf.Output.Articles.Directory = filepath.Join(tempDir, "content", "%Y-%m-%d")
	conf.Output.Images.Directory = filepath.Join(tempDir, "static", "images", "%Y-%m-%d")
	conf.Output.Images.BaseURL = Ptr("/images/%Y-%m-%d")
	conf.Output.Articles.Filename = "index.md"
	conf.Output.Images.Filename = "[:id].png"

	imageURL := "https://example.com/image.jpeg"
	imageRepo := &fakeImageRepository{contentType: "image/jpeg", body: "jpeg"}
	repo := &FileSystemArticleRepository{
		imageRepo: imageRepo,
		renderer:  NewHugoArticleRenderer(),
		logger:    slog.Default(),
	}

	article := &Article{
		Author:   "Author",
		Title:    "Title",
		Content:  "![image](" + imageURL + ")",
		Date:     "2021-01-01T00:00:00Z",
		Category: "Category",
		Tags:     []string{"tag"},
		Images: []*Image{
			NewImage(imageURL, "2021-01-01_000000", 0),
		},
	}

	err := repo.Save(context.Background(), article, conf)
	require.NoError(t, err)

	outputPath := filepath.Join(tempDir, "content", "2021-01-01", "index.md")
	data, err := os.ReadFile(outputPath)
	require.NoError(t, err)
	expectedFilename := "0.png"
	assert.Contains(t, string(data), "![image](/images/2021-01-01/"+expectedFilename+")")

	imagePath := filepath.Join(tempDir, "static", "images", "2021-01-01", expectedFilename)
	imageData, err := os.ReadFile(imagePath)
	require.NoError(t, err)
	assertEqualCmp(t, "jpeg", string(imageData))
}

func TestFileSystemArticleRepository_Save_UsesFrontMatterDateForOutputPaths(t *testing.T) {
	tempDir := t.TempDir()

	conf := *config.NewConfig()
	conf.Output.Articles.Directory = filepath.Join(tempDir, "content", "%Y-%m-%d_%H%M%S")
	conf.Output.Articles.Filename = "index.md"
	conf.Output.Images.Directory = filepath.Join(tempDir, "content", "%Y-%m-%d_%H%M%S", "images")
	conf.Output.Images.BaseURL = Ptr("/images/%Y-%m-%d_%H%M%S")
	conf.Output.Images.Filename = "%H-[:id].png"

	imageURL := "https://example.com/image.png"
	repo := &FileSystemArticleRepository{
		imageRepo: &fakeImageRepository{contentType: "image/png", body: "png"},
		renderer:  NewHugoArticleRenderer(),
		logger:    slog.Default(),
	}

	article := &Article{
		Author:      "Author",
		Title:       "Title",
		Content:     "![image](" + imageURL + ")",
		Date:        "2021-01-01T00:00:00Z",
		FrontMatter: NewFrontMatter(map[string]any{"date": "2021-02-03T04:05:06Z"}),
		Images: []*Image{
			NewImage(imageURL, "2021-01-01_000000", 0),
		},
	}

	err := repo.Save(context.Background(), article, conf)
	require.NoError(t, err)

	outputPath := filepath.Join(tempDir, "content", "2021-02-03_040506", "index.md")
	data, err := os.ReadFile(outputPath)
	require.NoError(t, err)
	assert.Contains(t, string(data), `date: "2021-02-03T04:05:06Z"`)
	assert.Contains(t, string(data), "![image](/images/2021-02-03_040506/04-0.png)")

	imagePath := filepath.Join(tempDir, "content", "2021-02-03_040506", "images", "04-0.png")
	imageData, err := os.ReadFile(imagePath)
	require.NoError(t, err)
	assertEqualCmp(t, "png", string(imageData))
}

func TestResolveImageFilename_UsesResolvedArticleDatetime(t *testing.T) {
	conf := *config.NewConfig()
	conf.Output.Images.Filename = "%H-[:id].png"

	datetime := time.Date(2021, 2, 3, 4, 5, 6, 0, time.UTC)
	got := resolveImageFilename(conf, NewImage("https://example.com/image.png", "2021-01-01_000000", 7), datetime)
	assertEqualCmp(t, "04-7.png", got)
}
