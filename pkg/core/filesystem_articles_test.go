package core

import (
	"context"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"testing"

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
	conf.Hugo.Directory.Articles = filepath.Join(tempDir, "content", "%Y-%m-%d")
	conf.Hugo.Directory.Images = filepath.Join(tempDir, "static", "images", "%Y-%m-%d")
	conf.Hugo.Url.Images = "/images/%Y-%m-%d"
	conf.Hugo.Filename.Articles = "index.md"

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
	expectedFilename := resolveImageBasename(NewImage(imageURL, "2021-01-01_000000", 0)) + ".jpg"
	assert.Contains(t, string(data), "![image](/images/2021-01-01/"+expectedFilename+")")

	imagePath := filepath.Join(tempDir, "static", "images", "2021-01-01", expectedFilename)
	imageData, err := os.ReadFile(imagePath)
	require.NoError(t, err)
	assertEqualCmp(t, "jpeg", string(imageData))
}
