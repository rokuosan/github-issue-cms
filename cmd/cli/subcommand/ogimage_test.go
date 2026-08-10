package subcommand

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/rokuosan/github-issue-cms/pkg/config"
	"github.com/rokuosan/github-issue-cms/pkg/core"
	"github.com/rokuosan/github-issue-cms/pkg/ogimage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// testConfig creates a minimal Config for testing.
func testConfig(imageDir string) config.Config {
	return config.Config{
		Output: &config.OutputConfig{
			Articles: &config.OutputArticlesConfig{
				Directory: "content/posts",
				Filename:  "%Y-%m-%d_%H%M%S.md",
			},
			Images: &config.OutputImagesConfig{
				Directory: imageDir,
				Filename:  "[:id].png",
			},
		},
	}
}

func TestNewOGCImageCommand(t *testing.T) {
	cmd := NewOGCImageCommand()

	assert.NotNil(t, cmd)
	assert.Equal(t, "ogimage", cmd.Use)
	assert.Contains(t, cmd.Short, "OGP image")

	// Verify the file flag.
	fileFlag := cmd.Flags().Lookup("file")
	assert.NotNil(t, fileFlag)
	assert.Equal(t, "f", fileFlag.Shorthand)

	// Verify the template flag.
	tmplFlag := cmd.Flags().Lookup("template")
	assert.NotNil(t, tmplFlag)
	assert.Equal(t, "t", tmplFlag.Shorthand)
}

func TestOGCImageCommand_MissingFile(t *testing.T) {
	cmd := NewOGCImageCommand()
	cmd.SetArgs([]string{}) // No file provided.

	err := cmd.Execute()
	assert.Error(t, err, "Should error when file is missing")
}

func TestOGCImageCommand_Help(t *testing.T) {
	cmd := NewOGCImageCommand()
	cmd.SetArgs([]string{"--help"})

	err := cmd.Execute()
	assert.NoError(t, err)
}

func TestArticleToOGPData(t *testing.T) {
	article := &core.Article{
		Title:    "Test Title",
		Author:   "alice",
		Date:     "2024-01-15T10:30:00Z",
		Category: "tech",
		Tags:     []string{"published", "go", "testing"},
	}

	conf := *config.NewConfig()
	conf.GitHub.Labels = []string{"published"}
	data := articleToOGPData(article, conf)

	assert.Equal(t, "Test Title", data.Title)
	assert.Equal(t, "alice", data.Author)
	assert.Equal(t, "2024-01-15", data.Date) // Formatted
	assert.Equal(t, "tech", data.Category)
	assert.Equal(t, []string{"go", "testing"}, data.Tags)
	assert.Equal(t, []string{"published", "go", "testing"}, article.Tags)
}

func TestArticleToOGPData_EmptyFields(t *testing.T) {
	article := &core.Article{
		Title: "Only Title",
	}

	data := articleToOGPData(article, *config.NewConfig())

	assert.Equal(t, "Only Title", data.Title)
	assert.Equal(t, "", data.Author)
	assert.Equal(t, "", data.Date)
	assert.Equal(t, "", data.Category)
	assert.Nil(t, data.Tags)
}

func TestFormatDateForOGP(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"2024-01-15T10:30:00Z", "2024-01-15"},
		{"2024-01-15T10:30:00+09:00", "2024-01-15"},
		{"2024-01-15", "2024-01-15"},
		{"invalid-date", "invalid-date"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := formatDateForOGP(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestResolveOGPOutputPath(t *testing.T) {
	dir := t.TempDir()

	// Simulate a config with image output directory containing time patterns.
	conf := testConfig(filepath.Join(dir, "images", "%Y-%m-%d_%H%M%S"))

	article := &core.Article{
		Date: "2024-01-15T10:30:00Z",
	}

	path, err := resolveOGPOutputPath(conf, article)
	require.NoError(t, err)
	assert.Contains(t, path, "ogp.jpeg")
	assert.Contains(t, path, "2024-01-15")
}

func TestResolveOGPOutputPath_NoImageDir(t *testing.T) {
	conf := testConfig(t.TempDir())
	conf.Output.Images.Directory = ""

	article := &core.Article{
		Date: "2024-01-15T10:30:00Z",
	}

	_, err := resolveOGPOutputPath(conf, article)
	assert.Error(t, err)
}

func TestOGCImage_Integration(t *testing.T) {
	// Skip unless explicitly requested.
	if os.Getenv("GIC_INTEGRATION_TEST") == "" {
		t.Skip("Skipping integration test: set GIC_INTEGRATION_TEST=1 to run")
	}

	dir := t.TempDir()

	// Create a test markdown file.
	mdContent := `---
author: testuser
title: Integration Test
date: 2024-01-15T10:30:00Z
categories: tech
tags:
  - go
  - testing
---

This is the article body.`

	mdPath := filepath.Join(dir, "article.md")
	err := os.WriteFile(mdPath, []byte(mdContent), 0o644)
	require.NoError(t, err)

	// Parse the markdown.
	article, err := core.ParseArticleFromMarkdown(mdPath)
	require.NoError(t, err)
	assert.Equal(t, "testuser", article.Author)

	// Convert to OGPData.
	data := articleToOGPData(article, *config.NewConfig())
	assert.Equal(t, "Integration Test", data.Title)

	// Render using the ogimage package (requires Chromium).
	renderer, err := ogimage.NewRenderer("")
	require.NoError(t, err)

	jpeg, err := renderer.Render(t.Context(), data)
	require.NoError(t, err)
	require.NotEmpty(t, jpeg)

	// Verify it's a JPEG.
	assert.Equal(t, byte(0xFF), jpeg[0])
	assert.Equal(t, byte(0xD8), jpeg[1])
}
