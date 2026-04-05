package converter_v2

import (
	"context"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// Export helper utilities for IssueArticle media handling.

type exportAssetLocation struct {
	dir    string
	prefix string
	mkdirs bool
}

func (a *IssueArticle) exportedContent(w io.Writer, options ExportOptions) (string, error) {
	content := a.Content()
	replacements := map[string]string{}

	location, err := a.resolveAssetDir(w, options)
	if err != nil {
		return "", err
	}

	for _, image := range a.Images() {
		src := image.Destination()
		if !isGitHubUserContentMediaURL(src) || replacements[src] != "" {
			continue
		}
		localPath, err := a.downloadAndSaveMedia(src, location)
		if err != nil {
			return "", err
		}
		replacements[src] = localPath
	}

	for source, replacement := range replacements {
		content = strings.ReplaceAll(content, source, replacement)
	}
	return content, nil
}

func (a *IssueArticle) resolveAssetDir(w io.Writer, options ExportOptions) (exportAssetLocation, error) {
	outputFile := ""
	if file, ok := w.(*os.File); ok {
		info, err := file.Stat()
		if err != nil {
			return exportAssetLocation{}, err
		}
		if !info.Mode().IsRegular() {
			return exportAssetLocation{}, fmt.Errorf("writer is not a regular file")
		}
		outputFile = file.Name()
	}

	if options.AssetDirectory != "" {
		assetDir := filepath.Clean(options.AssetDirectory)
		info, err := os.Stat(assetDir)
		if err != nil {
			return exportAssetLocation{}, err
		}
		if !info.IsDir() {
			return exportAssetLocation{}, fmt.Errorf("%s is not a directory", assetDir)
		}
		return exportAssetLocation{
			dir:    assetDir,
			prefix: toRelativePath(outputFile, assetDir),
		}, nil
	}

	if outputFile == "" {
		tmpDir, err := os.MkdirTemp("", "gic-assets-*")
		if err != nil {
			return exportAssetLocation{}, err
		}
		return exportAssetLocation{
			dir:    tmpDir,
			prefix: filepath.ToSlash(tmpDir),
		}, nil
	}

	assetDir := filepath.Join(filepath.Dir(outputFile), "assets")
	return exportAssetLocation{
		dir:    assetDir,
		prefix: "assets",
		mkdirs: true,
	}, nil
}

func toRelativePath(outputFile, targetDir string) string {
	if outputFile == "" {
		return filepath.ToSlash(targetDir)
	}
	outputDir := filepath.Dir(outputFile)
	rel, err := filepath.Rel(outputDir, targetDir)
	if err != nil {
		return filepath.Base(targetDir)
	}
	if rel == "." {
		return "."
	}
	return filepath.ToSlash(rel)
}

func (a *IssueArticle) downloadAndSaveMedia(mediaURL string, location exportAssetLocation) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, mediaURL, nil)
	if err != nil {
		return "", err
	}
	resp, err := a.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to download media %s: status=%d", mediaURL, resp.StatusCode)
	}

	filename := a.mediaFileName(mediaURL, resp.Header.Get("Content-Type"))
	if location.dir == "" {
		return "", fmt.Errorf("asset directory is not set")
	}
	if location.mkdirs {
		if err := os.MkdirAll(location.dir, 0o755); err != nil {
			return "", err
		}
	}

	downloadPath := filepath.Join(location.dir, filename)
	file, err := os.Create(downloadPath)
	if err != nil {
		return "", err
	}
	defer file.Close()
	if _, err := io.Copy(file, resp.Body); err != nil {
		return "", err
	}
	return toOutputMediaPath(location.prefix, filename), nil
}

func (a *IssueArticle) mediaFileName(mediaURL, contentType string) string {
	ext := mediaExtension(mediaURL)
	if ext == "" {
		ext = mediaExtensionFromContentType(contentType)
	}
	if ext == "" {
		ext = ".bin"
	}
	return fmt.Sprintf("%s%s", hashMediaURL(mediaURL), ext)
}

func toOutputMediaPath(prefix, filename string) string {
	if prefix == "." {
		return filename
	}
	return filepath.ToSlash(filepath.Join(prefix, filename))
}

func isGitHubUserContentMediaURL(raw string) bool {
	return isGitHubUserContentURL(raw) && mediaExtension(raw) != ""
}

func isGitHubUserContentURL(raw string) bool {
	parsed, err := url.Parse(raw)
	if err != nil {
		return false
	}
	return parsed.Scheme == "https" &&
		strings.HasSuffix(strings.ToLower(parsed.Hostname()), ".githubusercontent.com")
}

func mediaExtension(raw string) string {
	parsed, err := url.Parse(raw)
	if err != nil {
		return ""
	}
	switch strings.ToLower(filepath.Ext(parsed.Path)) {
	case ".png", ".jpg", ".jpeg", ".gif", ".webp", ".bmp", ".svg", ".mp4", ".webm", ".mov", ".mkv", ".m4v", ".ogv", ".avi":
		return strings.ToLower(filepath.Ext(parsed.Path))
	default:
		return ""
	}
}

func mediaExtensionFromContentType(contentType string) string {
	switch strings.ToLower(strings.TrimSpace(strings.SplitN(contentType, ";", 2)[0])) {
	case "image/png":
		return ".png"
	case "image/jpeg", "image/jpg":
		return ".jpg"
	case "image/gif":
		return ".gif"
	case "image/webp":
		return ".webp"
	case "video/mp4":
		return ".mp4"
	case "video/webm":
		return ".webm"
	case "video/quicktime":
		return ".mov"
	case "video/x-msvideo":
		return ".avi"
	default:
		return ""
	}
}

func hashMediaURL(raw string) string {
	sum := sha1.Sum([]byte(raw))
	return hex.EncodeToString(sum[:])[:12]
}
