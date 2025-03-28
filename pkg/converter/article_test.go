package converter

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/rokuosan/github-issue-cms/pkg/config"
)

func Test_Article_Transform(t *testing.T) {
	article := &Article{
		Author:           "John Doe",
		Title:            "Hello, World!",
		Content:          "This is a test content.",
		Date:             "2021-01-01T00:00:00Z",
		Category:         "Test",
		Tags:             []string{"test", "sample"},
		Draft:            false,
		ExtraFrontMatter: "extra: true\nauthors:\n    - TEST\nauthor: sample\n",
		Key:              "hello-world",
	}

	got, err := article.Transform()
	if err != nil {
		t.Fatalf("Article.Transform() error = %v", err)
	}

	want := `---
author: sample
title: Hello, World!
date: "2021-01-01T00:00:00Z"
categories: Test
tags:
    - test
    - sample
draft: false
authors:
    - TEST
extra: true
---

This is a test content.
`

	if got != want {
		t.Errorf("Article.Transform() = %v, want %v", got, want)
	}
}
func Test_createDirectoryIfNotExist(t *testing.T) {
	tempDir := t.TempDir()
	subDir := filepath.Join(tempDir, "subdir")

	err := createDirectoryIfNotExist(subDir)
	if err != nil {
		t.Fatalf("createDirectoryIfNotExist() error = %v", err)
	}

	// Check if the directory exists
	if _, err := os.Stat(subDir); os.IsNotExist(err) {
		t.Errorf("Directory %s was not created", subDir)
	}
}

func Test_createFileAndWrite(t *testing.T) {
	tempDir := t.TempDir()
	filePath := filepath.Join(tempDir, "testfile.txt")
	content := "Hello, World!"

	err := createFileAndWrite(filePath, content)
	if err != nil {
		t.Fatalf("createFileAndWrite() error = %v", err)
	}

	// Check if the file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		t.Errorf("File %s was not created", filePath)
	}

	// Verify the file content
	data, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("Failed to read file %s: %v", filePath, err)
	}
	if string(data) != content {
		t.Errorf("File content = %s, want %s", string(data), content)
	}
}
func Test_parseDateTime(t *testing.T) {
	tests := []struct {
		name    string
		date    string
		want    time.Time
		wantErr bool
	}{
		{
			name:    "ISO8601 format",
			date:    "2021-01-01T00:00:00Z",
			want:    time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC),
			wantErr: false,
		},
		{
			name:    "Date only format",
			date:    "2021-01-01",
			want:    time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC),
			wantErr: false,
		},
		{
			name:    "Invalid format",
			date:    "invalid-date",
			want:    time.Time{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			article := &Article{
				Date: tt.date,
			}
			got, err := article.parseDateTime()
			if (err != nil) != tt.wantErr {
				t.Errorf("parseDateTime() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && !got.Equal(tt.want) {
				t.Errorf("parseDateTime() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestArticle_Export(t *testing.T) {
	// Create a mock config for testing
	tempDir := t.TempDir()
	articlesDir := filepath.Join(tempDir, "articles")
	imagesDir := filepath.Join(tempDir, "images")

	mockConfig := config.Config{
		Hugo: &config.HugoConfig{
			Directory: &config.HugoDirectoryConfig{
				Articles: articlesDir,
				Images:   imagesDir,
			},
			Filename: &config.HugoFilenameConfig{
				Articles: "%Y-%m-%d_%H%M%S.md",
				Images:   "image-[:id].png",
			},
		},
	}

	article := &Article{
		Author:           "John Doe",
		Title:            "Test Article",
		Content:          "This is a test article content.",
		Date:             "2021-01-01T00:00:00Z",
		Category:         "Test",
		Tags:             []string{"test", "export"},
		Draft:            false,
		ExtraFrontMatter: "",
		Key:              "test-article",
	}

	article.Export(mockConfig)

	// Check if the article file was created
	expectedFilePath := filepath.Join(articlesDir, "2021-01-01_000000.md")
	if _, err := os.Stat(expectedFilePath); os.IsNotExist(err) {
		t.Errorf("Export() did not create the expected file: %s", expectedFilePath)
	}
}

func TestArticle_TransformWithInvalidExtraFrontMatter(t *testing.T) {
	article := &Article{
		Author:           "John Doe",
		Title:            "Invalid Extra Front Matter",
		Content:          "Test content",
		Date:             "2021-01-01",
		Category:         "Test",
		Tags:             []string{"test"},
		Draft:            false,
		ExtraFrontMatter: "invalid: yaml: [unclosed",
		Key:              "test",
	}

	_, err := article.Transform()
	if err == nil {
		t.Errorf("Transform() with invalid ExtraFrontMatter should return error")
	}
}

func TestImageDescriptor_Save(t *testing.T) {
	// Start a test HTTP server that returns PNG data.
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "image/png")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("PNGDATA"))
	}))
	defer ts.Close()

	// Create a temporary directory for saving the image.
	tempDir, err := os.MkdirTemp("", "imagetest")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	// Create an instance of ImageDescriptor using the test server URL.
	img := &ImageDescriptor{
		Url: ts.URL,
		Id:  1,
	}

	filename := "test.png"
	// Invoke the Save method.
	img.Save(tempDir, filename)

	// Verify the saved file exists and contains the expected data.
	savedPath := filepath.Join(tempDir, filename)
	data, err := os.ReadFile(savedPath)
	if err != nil {
		t.Fatalf("failed to read saved file: %v", err)
	}
	if string(data) != "PNGDATA" {
		t.Errorf("expected file content %q but got %q", "PNGDATA", string(data))
	}
}
