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
	tests := []struct {
		name              string
		contentType       string
		body              string
		content           string
		images            []*Image
		wantContains      []string
		wantNotContains   []string
		wantSavedFilename string
		wantSavedFileBody string
	}{
		{
			name:              "rewrites markdown image URLs",
			contentType:       "image/jpeg",
			body:              "jpeg",
			content:           "![image](https://example.com/image.jpeg)",
			images:            []*Image{NewImage("https://example.com/image.jpeg", "2021-01-01_000000", 0)},
			wantContains:      []string{"![image](/images/2021-01-01/0.png)"},
			wantSavedFilename: "0.png",
			wantSavedFileBody: "jpeg",
		},
		{
			name:        "rewrites HTML image URLs",
			contentType: "image/png",
			body:        "png",
			content:     `<img class="rounded" src="https://example.com/test.png" alt="test image" loading="lazy">`,
			images:      []*Image{NewImage("https://example.com/test.png", "2021-01-01_000000", 0)},
			wantContains: []string{
				`<img class="rounded" src="/images/2021-01-01/0.png" alt="test image" loading="lazy">`,
			},
			wantSavedFilename: "0.png",
			wantSavedFileBody: "png",
		},
		{
			name:        "rewrites multiple GitHub-hosted image URLs in one pass",
			contentType: "image/png",
			body:        "png",
			content: "![first](https://github.com/user-attachments/assets/11111111-1111-1111-1111-111111111111)\n\n" +
				`<img src="https://github.com/user-attachments/assets/22222222-2222-2222-2222-222222222222" alt="second">`,
			images: []*Image{
				NewImage("https://github.com/user-attachments/assets/11111111-1111-1111-1111-111111111111", "2021-01-01_000000", 0),
				NewImage("https://github.com/user-attachments/assets/22222222-2222-2222-2222-222222222222", "2021-01-01_000000", 1),
			},
			wantContains: []string{
				"![first](/images/2021-01-01/0.png)",
				`<img src="/images/2021-01-01/1.png" alt="second">`,
			},
			wantNotContains: []string{
				"https://github.com/user-attachments/assets/11111111-1111-1111-1111-111111111111",
				"https://github.com/user-attachments/assets/22222222-2222-2222-2222-222222222222",
			},
			wantSavedFilename: "0.png",
			wantSavedFileBody: "png",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tempDir := t.TempDir()

			conf := *config.NewConfig()
			conf.Output.Articles.Directory = filepath.Join(tempDir, "content", "%Y-%m-%d")
			conf.Output.Images.Directory = filepath.Join(tempDir, "static", "images", "%Y-%m-%d")
			conf.Output.Images.BaseURL = Ptr("/images/%Y-%m-%d")
			conf.Output.Articles.Filename = "index.md"
			conf.Output.Images.Filename = "[:id].png"

			repo := &FileSystemArticleRepository{
				imageRepo: &fakeImageRepository{contentType: tt.contentType, body: tt.body},
				renderer:  NewHugoArticleRenderer(),
				logger:    slog.Default(),
			}

			article := &Article{
				Author:   "Author",
				Title:    "Title",
				Content:  tt.content,
				Date:     "2021-01-01T00:00:00Z",
				Category: "Category",
				Tags:     []string{"tag"},
				Images:   tt.images,
			}

			err := repo.Save(context.Background(), article, conf)
			require.NoError(t, err)

			outputPath := filepath.Join(tempDir, "content", "2021-01-01", "index.md")
			data, err := os.ReadFile(outputPath)
			require.NoError(t, err)
			for _, want := range tt.wantContains {
				assert.Contains(t, string(data), want)
			}
			for _, unwanted := range tt.wantNotContains {
				assert.NotContains(t, string(data), unwanted)
			}

			imagePath := filepath.Join(tempDir, "static", "images", "2021-01-01", tt.wantSavedFilename)
			imageData, err := os.ReadFile(imagePath)
			require.NoError(t, err)
			assertEqualCmp(t, tt.wantSavedFileBody, string(imageData))
		})
	}
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

func TestFileSystemArticleRepository_Save_AcceptsFrontMatterDateWithOffset(t *testing.T) {
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
		FrontMatter: NewFrontMatter(map[string]any{"date": "2021-02-03T04:05:06+09:00"}),
		Images: []*Image{
			NewImage(imageURL, "2021-01-01_000000", 0),
		},
	}

	err := repo.Save(context.Background(), article, conf)
	require.NoError(t, err)

	outputPath := filepath.Join(tempDir, "content", "2021-02-03_040506", "index.md")
	data, err := os.ReadFile(outputPath)
	require.NoError(t, err)
	assert.Contains(t, string(data), `date: "2021-02-03T04:05:06+09:00"`)
	assert.Contains(t, string(data), "![image](/images/2021-02-03_040506/04-0.png)")
}

func TestResolveImageFilename_UsesResolvedArticleDatetime(t *testing.T) {
	conf := *config.NewConfig()
	conf.Output.Images.Filename = "%H-[:id].png"

	datetime := time.Date(2021, 2, 3, 4, 5, 6, 0, time.UTC)
	got := resolveImageFilename(conf, NewImage("https://example.com/image.png", "2021-01-01_000000", 7), datetime)
	assertEqualCmp(t, "04-7.png", got)
}
