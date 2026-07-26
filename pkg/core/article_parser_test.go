package core

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseArticleFromMarkdown(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.md")

	content := `---
author: testuser
title: My Test Article
date: 2024-01-15T10:30:00Z
categories: tech
tags:
  - go
  - testing
draft: false
---

This is the article content.

It has multiple paragraphs.`

	err := os.WriteFile(path, []byte(content), 0o644)
	require.NoError(t, err)

	article, err := ParseArticleFromMarkdown(path)
	require.NoError(t, err)

	assert.Equal(t, "testuser", article.Author)
	assert.Equal(t, "My Test Article", article.Title)
	assert.Equal(t, "2024-01-15T10:30:00Z", article.Date)
	assert.Equal(t, "tech", article.Category)
	assert.Equal(t, []string{"go", "testing"}, article.Tags)
	assert.False(t, article.Draft)
	assert.Contains(t, article.Content, "This is the article content.")
	assert.Contains(t, article.Content, "It has multiple paragraphs.")
}

func TestParseArticleFromMarkdown_NoFrontmatter(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.md")

	content := "Just content, no frontmatter"
	err := os.WriteFile(path, []byte(content), 0o644)
	require.NoError(t, err)

	_, err = ParseArticleFromMarkdown(path)
	assert.Error(t, err)
}

func TestParseArticleFromMarkdown_NonexistentFile(t *testing.T) {
	_, err := ParseArticleFromMarkdown("/nonexistent/file.md")
	assert.Error(t, err)
}

func TestParseArticleContent(t *testing.T) {
	t.Run("valid article", func(t *testing.T) {
		content := `---
author: alice
title: Hello World
date: 2024-01-15
---
Article body here.`

		article, err := parseArticleContent(content)
		require.NoError(t, err)
		assert.Equal(t, "alice", article.Author)
		assert.Equal(t, "Hello World", article.Title)
		assert.Equal(t, "Article body here.", article.Content)
	})

	t.Run("missing closing delimiter", func(t *testing.T) {
		content := `---
author: alice
title: Hello
`
		_, err := parseArticleContent(content)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "missing closing frontmatter delimiter")
	})

	t.Run("empty fields", func(t *testing.T) {
		content := "---\n\n---\n"
		article, err := parseArticleContent(content)
		require.NoError(t, err)
		assert.Equal(t, "", article.Title)
	})
}
