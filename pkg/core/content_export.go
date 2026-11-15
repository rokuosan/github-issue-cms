package core

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const contentExportMarker = ".gic-content-export"
const contentExportVersion = "github-issue-cms content export v1\n"

// ContentMetadata describes the generated JSON resources.
type ContentMetadata struct {
	SchemaVersion       int           `json:"schemaVersion"`
	GeneratedAt         time.Time     `json:"generatedAt"`
	TotalArticles       int           `json:"totalArticles"`
	PageSize            int           `json:"pageSize"`
	TotalPages          int           `json:"totalPages"`
	Pages               []string      `json:"pages"`
	ArticlePathTemplate string        `json:"articlePathTemplate"`
	Search              ContentSearch `json:"search"`
}

// ContentSearch identifies the static search index and its searchable fields.
type ContentSearch struct {
	Index  string   `json:"index"`
	Fields []string `json:"fields"`
}

// SearchEntry contains the text needed for client-side metadata search.
type SearchEntry struct {
	ID      int      `json:"id"`
	Title   string   `json:"title"`
	Tags    []string `json:"tags"`
	Summary string   `json:"summary"`
	URL     string   `json:"url"`
}

// Metadata describes this snapshot's fixed-size pages.
func (s *ContentSnapshot) Metadata() ContentMetadata {
	totalPages := len(s.Articles) / s.PageSize
	if len(s.Articles)%s.PageSize != 0 {
		totalPages++
	}
	pages := make([]string, 0, totalPages)
	for page := 1; page <= totalPages; page++ {
		pages = append(pages, fmt.Sprintf("/data/articles/page/%d.json", page))
	}
	return ContentMetadata{
		SchemaVersion: 1, GeneratedAt: s.GeneratedAt,
		TotalArticles: len(s.Articles), PageSize: s.PageSize, TotalPages: totalPages,
		Pages: pages, ArticlePathTemplate: "/data/article/{id}.json",
		Search: ContentSearch{Index: "/data/search/index.json", Fields: []string{"title", "tags", "summary"}},
	}
}

// SearchIndex returns search records in the snapshot's default order.
func (s *ContentSnapshot) SearchIndex() []SearchEntry {
	entries := make([]SearchEntry, 0, len(s.Articles))
	for _, article := range s.Articles {
		tags := append([]string{}, article.Tags...)
		entries = append(entries, SearchEntry{
			ID: article.ID, Title: article.Title, Tags: tags,
			Summary: contentSummary(article.Content), URL: fmt.Sprintf("/data/article/%d.json", article.ID),
		})
	}
	return entries
}

func contentSummary(content string) string {
	runes := []rune(strings.TrimSpace(content))
	if len(runes) > 200 {
		return string(runes[:200]) + "…"
	}
	return string(runes)
}

// ExportContentSnapshot writes a complete static deployment. An existing output
// must be empty or a previous export; generated output is replaced, not merged.
func ExportContentSnapshot(ctx context.Context, snapshot *ContentSnapshot, output string) error {
	if snapshot == nil || snapshot.PageSize < 1 || snapshot.PageSize > 100 {
		return fmt.Errorf("invalid content snapshot")
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	if strings.TrimSpace(output) == "" {
		return fmt.Errorf("output directory is required")
	}
	destination, err := filepath.Abs(output)
	if err != nil {
		return fmt.Errorf("resolve output directory: %w", err)
	}
	if err := validateContentDestination(destination); err != nil {
		return err
	}
	parent := filepath.Dir(destination)
	if err := os.MkdirAll(parent, 0o755); err != nil {
		return fmt.Errorf("create output parent: %w", err)
	}
	staging, err := os.MkdirTemp(parent, ".gic-content-stage-")
	if err != nil {
		return fmt.Errorf("create export staging directory: %w", err)
	}
	defer os.RemoveAll(staging)
	if err := writeContentExport(ctx, snapshot, staging); err != nil {
		return err
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	// Recheck before replacing an output that may have changed while writing.
	if err := validateContentDestination(destination); err != nil {
		return err
	}
	if err := os.Chmod(staging, 0o755); err != nil {
		return err
	}
	if _, err := os.Lstat(destination); os.IsNotExist(err) {
		return os.Rename(staging, destination)
	} else if err != nil {
		return err
	}
	backup, err := os.MkdirTemp(parent, ".gic-content-backup-")
	if err != nil {
		return err
	}
	if err := os.Remove(backup); err != nil {
		return err
	}
	if err := os.Rename(destination, backup); err != nil {
		return fmt.Errorf("move previous export: %w", err)
	}
	if err := os.Rename(staging, destination); err != nil {
		if restoreErr := os.Rename(backup, destination); restoreErr != nil {
			return fmt.Errorf("publish export: %w (restore failed: %v; previous export at %s)", err, restoreErr, backup)
		}
		return fmt.Errorf("publish export: %w", err)
	}
	if err := os.RemoveAll(backup); err != nil {
		return fmt.Errorf("remove previous export at %s: %w", backup, err)
	}
	return nil
}

func validateContentDestination(destination string) error {
	info, err := os.Lstat(destination)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return err
	}
	if !info.IsDir() {
		return fmt.Errorf("output must be a directory, not a file or symlink: %s", destination)
	}
	entries, err := os.ReadDir(destination)
	if err != nil {
		return err
	}
	if len(entries) == 0 {
		return nil
	}
	marker := filepath.Join(destination, contentExportMarker)
	markerInfo, err := os.Lstat(marker)
	if err != nil || !markerInfo.Mode().IsRegular() {
		return fmt.Errorf("refusing to replace non-export directory: %s", destination)
	}
	data, err := os.ReadFile(marker)
	if err != nil || string(data) != contentExportVersion {
		return fmt.Errorf("refusing to replace non-export directory: %s", destination)
	}
	return nil
}

func writeContentExport(ctx context.Context, snapshot *ContentSnapshot, directory string) error {
	write := func(name string, data []byte) error {
		if err := ctx.Err(); err != nil {
			return err
		}
		filename := filepath.Join(directory, filepath.FromSlash(name))
		if err := os.MkdirAll(filepath.Dir(filename), 0o755); err != nil {
			return err
		}
		return os.WriteFile(filename, data, 0o644)
	}
	writeJSON := func(name string, value any) error {
		data, err := json.Marshal(value)
		if err != nil {
			return fmt.Errorf("encode %s: %w", name, err)
		}
		return write(name, append(data, '\n'))
	}
	metadata := snapshot.Metadata()
	if err := writeJSON("data/metadata.json", metadata); err != nil {
		return err
	}
	for page := 1; page <= metadata.TotalPages; page++ {
		result, err := snapshot.Query(ContentQuery{Page: page})
		if err != nil {
			return err
		}
		if err := writeJSON(fmt.Sprintf("data/articles/page/%d.json", page), result); err != nil {
			return err
		}
	}
	for _, article := range snapshot.Articles {
		if article.ID < 1 {
			return fmt.Errorf("invalid article ID: %d", article.ID)
		}
		if err := writeJSON(fmt.Sprintf("data/article/%d.json", article.ID), article); err != nil {
			return err
		}
	}
	if err := writeJSON("data/search/index.json", snapshot.SearchIndex()); err != nil {
		return err
	}
	for name, asset := range snapshot.Assets {
		// Assets are generated paths, never arbitrary filesystem destinations.
		rel := strings.TrimPrefix(name, "/")
		if !strings.HasPrefix(name, "/images/") || strings.Contains(rel, "\\") || !filepath.IsLocal(filepath.FromSlash(rel)) || filepath.ToSlash(filepath.Clean(rel)) != rel {
			return fmt.Errorf("invalid asset path: %s", name)
		}
		if err := write(rel, asset.Data); err != nil {
			return err
		}
	}
	if err := write("_headers", []byte("/data/*\n  Content-Type: application/json; charset=utf-8\n  Cache-Control: no-cache\n/images/*\n  Cache-Control: no-cache\n")); err != nil {
		return err
	}
	return write(contentExportMarker, []byte(contentExportVersion))
}
