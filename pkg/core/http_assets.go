package core

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"mime"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const defaultHTTPTimeout = 30
const maxRedirects = 10

// HTTPImageRepository downloads images over HTTP.
type HTTPImageRepository struct {
	token  string
	logger *slog.Logger
	client *http.Client
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
		client: &http.Client{Timeout: defaultHTTPTimeout * time.Second},
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
func (r *HTTPImageRepository) downloadImage(ctx context.Context, imageURL string) (io.ReadCloser, string, error) {
	// Only send the token over HTTPS to prevent leaking credentials.
	if r.token != "" && isHTTPS(imageURL) {
		if body, contentType, err := r.sendRequest(ctx, imageURL, true); err == nil {
			return body, contentType, nil
		} else {
			r.logger.Warn("authenticated image download failed; retrying without token", "url", imageURL, "error", err)

			body, contentType, fallbackErr := r.sendRequest(ctx, imageURL, false)
			if fallbackErr == nil {
				return body, contentType, nil
			}
			return nil, "", errors.Join(
				fmt.Errorf("authenticated request failed: %w", err),
				fmt.Errorf("unauthenticated fallback failed: %w", fallbackErr),
			)
		}
	}

	// No token was configured or the URL is not HTTPS — only an unauthenticated request is possible.
	return r.sendRequest(ctx, imageURL, false)
}

// sendRequest sends an HTTP request.
func (r *HTTPImageRepository) sendRequest(ctx context.Context, url string, includeToken bool) (io.ReadCloser, string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, "", err
	}

	client := r.client
	if includeToken && r.token != "" {
		req.Header.Set("Authorization", "token "+r.token)
		// Prevent the token from leaking to a redirect destination.
		client = r.clientWithRedirectGuard()
	}

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

// clientWithRedirectGuard returns a copy of the HTTP client that strips the
// Authorization header when redirecting from HTTPS to HTTP to prevent token
// leakage. Redirects to other HTTPS URLs retain the token.
func (r *HTTPImageRepository) clientWithRedirectGuard() *http.Client {
	c := *r.client
	c.CheckRedirect = func(redirectReq *http.Request, via []*http.Request) error {
		if !isHTTPS(redirectReq.URL.String()) {
			redirectReq.Header.Del("Authorization")
		}
		if len(via) >= maxRedirects {
			return fmt.Errorf("stopped after %d redirects", maxRedirects)
		}
		return nil
	}
	return &c
}

func isHTTPS(rawURL string) bool {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return false
	}
	return strings.EqualFold(parsed.Scheme, "https")
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
