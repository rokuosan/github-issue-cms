package core

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"mime"
	"net/http"
	"strings"
)

// HTTPImageRepository downloads images over HTTP.
type HTTPImageRepository struct {
	token  string
	logger *slog.Logger
}

// NewHTTPImageRepository creates a new HTTPImageRepository.
func NewHTTPImageRepository(token string) AssetFetcher {
	return NewHTTPImageRepositoryWithLogger(token, nil)
}

// NewHTTPImageRepositoryWithLogger creates a new HTTPImageRepository with an injected logger.
func NewHTTPImageRepositoryWithLogger(token string, logger *slog.Logger) AssetFetcher {
	return &HTTPImageRepository{
		token:  token,
		logger: defaultLogger(logger),
	}
}

// Fetch retrieves an image stream over HTTP.
func (r *HTTPImageRepository) Fetch(ctx context.Context, image *Image) (*ImageAsset, error) {
	body, contentType, err := r.downloadImage(ctx, image.URL)
	if err != nil {
		return nil, fmt.Errorf("failed to download image from %s: %w", image.URL, err)
	}

	return &ImageAsset{
		Body:        body,
		ContentType: contentType,
	}, nil
}

// downloadImage downloads an image over HTTP.
func (r *HTTPImageRepository) downloadImage(ctx context.Context, url string) (io.ReadCloser, string, error) {
	// Try an authenticated request first.
	if r.token != "" {
		if body, contentType, err := r.sendRequest(ctx, url, true); err == nil {
			return body, contentType, nil
		}
	}

	// Fall back to an unauthenticated request.
	return r.sendRequest(ctx, url, false)
}

// sendRequest sends an HTTP request.
func (r *HTTPImageRepository) sendRequest(ctx context.Context, url string, includeToken bool) (io.ReadCloser, string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, "", err
	}

	if includeToken && r.token != "" {
		req.Header.Set("Authorization", "token "+r.token)
	}

	client := new(http.Client)
	resp, err := client.Do(req)
	if err != nil {
		return nil, "", err
	}

	// Validate the response.
	contentType := normalizeContentType(resp.Header.Get("Content-Type"))
	if resp.StatusCode != http.StatusOK || !isSupportedImageContentType(contentType) {
		resp.Body.Close()
		return nil, "", fmt.Errorf("bad response: status=%d, content-type=%s", resp.StatusCode, contentType)
	}

	return resp.Body, contentType, nil
}

func normalizeContentType(value string) string {
	mediaType, _, err := mime.ParseMediaType(value)
	if err != nil {
		return strings.TrimSpace(value)
	}
	return mediaType
}

func isSupportedImageContentType(contentType string) bool {
	switch contentType {
	case "image/png", "image/jpeg", "image/jpg", "image/gif", "image/webp", "image/svg+xml", "image/bmp":
		return true
	default:
		return false
	}
}
