package converter

import (
	"bufio"
	"fmt"
	"github.com/spf13/viper"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// Article is the article for Hugo.
type Article struct {
	// Author is the author of the article.
	Author string `json:"author"`

	// Title is the title of the article.
	Title string `json:"title"`

	// Content is the content of the article.
	Content string `json:"content"`

	// Date is the date of the article.
	Date string `json:"date"`

	// Category is the category of the article.
	Category string `json:"category"`

	// Tags is the tags of the article.
	Tags []string `json:"tags"`

	// Draft is the draft of the article.
	// If it is true, the article will not be published.
	Draft bool `json:"draft"`

	// ExtraFrontMatter is the extra front matter of the article.
	// It must be a valid YAML string.
	ExtraFrontMatter string `json:"extra_front_matter"`

	Key string
}

func (article *Article) ExportToMarkdown(name string) {
	// Get export directory
	articlesDir := viper.GetString("hugo.directory.articles")
	if articlesDir == "" {
		articlesDir = "content/posts"
	}
	// Sanitize
	articlesDir = strings.ReplaceAll(articlesDir, ".", "")
	if strings.HasPrefix(articlesDir, "/") {
		re := regexp.MustCompile("^/+")
		match := re.FindString(articlesDir)
		articlesDir = strings.Replace(articlesDir, match, "", 1)
	}
	if strings.HasSuffix(articlesDir, "/") {
		re := regexp.MustCompile("/+$")
		match := re.FindString(articlesDir)
		articlesDir = strings.Replace(articlesDir, match, "", 1)
	}

	// Create directory
	path := filepath.Join(strings.Split(articlesDir, "/")...)
	if _, err := os.Stat(path); os.IsNotExist(err) {
		if err := os.MkdirAll(path, 0777); err != nil {
			panic(err)
		}
	}

	// Create file
	path = filepath.Join(path, name+".md")
	file, err := os.Create(path)
	if err != nil {
		panic(err)
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			panic(err)
		}
	}(file)

	// Build String
	var content []byte
	var tags []byte
	for _, tag := range article.Tags {
		tags = append(tags, "  - "...)
		tags = append(tags, tag...)
		tags = append(tags, '\n')
	}
	rawText := strings.TrimSpace(fmt.Sprintf(`---
title: %s
author: %s
date: %s
draft: %t
categories: 
  - %s
tags:
%s
%s
---

%s
`,
		article.Title,
		article.Author,
		article.Date,
		article.Draft,
		article.Category,
		tags,
		article.ExtraFrontMatter,
		article.Content))
	content = append(content, rawText...)

	// Write
	writer := bufio.NewWriter(file)
	_, _ = writer.Write(content)
	_ = writer.Flush()
}
