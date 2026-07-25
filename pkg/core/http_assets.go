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

// HTTPImageRepository downloads images over HTTP.
type HTTPImageRepository struct {
	token        string
	trustedHosts map[string]struct{}
	logger       *slog.Logger
	client       *http.Client
}

// NewHTTPImageRepository creates a new HTTPImageRepository without trusted hosts.
func NewHTTPImageRepository(token string) AssetFetcher {
	return NewHTTPImageRepositoryWithTrustedHostsAndLogger(token, nil, nil)
}

// NewHTTPImageRepositoryWithTrustedHosts creates a new HTTPImageRepository that only sends a token to trusted HTTPS hosts.
func NewHTTPImageRepositoryWithTrustedHosts(token string, trustedHosts []string) *HTTPImageRepository {
	return NewHTTPImageRepositoryWithTrustedHostsAndLogger(token, trustedHosts, nil)
}

// NewHTTPImageRepositoryWithLogger creates a new HTTPImageRepository with an injected logger and no trusted hosts.
func NewHTTPImageRepositoryWithLogger(token string, logger *slog.Logger) AssetFetcher {
	return NewHTTPImageRepositoryWithTrustedHostsAndLogger(token, nil, logger)
}

// NewHTTPImageRepositoryWithTrustedHostsAndLogger creates a new HTTPImageRepository with trusted hosts and an injected logger.
func NewHTTPImageRepositoryWithTrustedHostsAndLogger(token string, trustedHosts []string, logger *slog.Logger) *HTTPImageRepository {
	hosts := make(map[string]struct{}, len(trustedHosts))
	for _, host := range trustedHosts {
		hosts[strings.ToLower(host)] = struct{}{}
	}

	return &HTTPImageRepository{
		token:        token,
		trustedHosts: hosts,
		logger:       defaultLogger(logger),
		client:       &http.Client{Timeout: defaultHTTPTimeout * time.Second},
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
	// A token is only sent to explicitly trusted HTTPS hosts.
	if r.token != "" && r.isTrustedURL(url) {
		if body, contentType, err := r.sendRequest(ctx, url, true); err == nil {
			return body, contentType, nil
		} else {
			r.logger.Warn("authenticated image download failed; retrying without token", "url", url, "error", err)

			body, contentType, fallbackErr := r.sendRequest(ctx, url, false)
			if fallbackErr == nil {
				return body, contentType, nil
			}
			return nil, "", errors.Join(
				fmt.Errorf("authenticated request failed: %w", err),
				fmt.Errorf("unauthenticated fallback failed: %w", fallbackErr),
			)
		}
	}

	// No token was configured, so only an unauthenticated request is possible.
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

	client := *r.client
	if includeToken && r.token != "" {
		previousCheckRedirect := client.CheckRedirect
		client.CheckRedirect = func(redirectReq *http.Request, via []*http.Request) error {
			redirectReq.Header.Del("Authorization")
			if r.isTrustedURL(redirectReq.URL.String()) {
				redirectReq.Header.Set("Authorization", "token "+r.token)
			}
			if previousCheckRedirect != nil {
				return previousCheckRedirect(redirectReq, via)
			}
			return nil
		}
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

func (r *HTTPImageRepository) isTrustedURL(rawURL string) bool {
	parsedURL, err := url.Parse(rawURL)
	if err != nil || !strings.EqualFold(parsedURL.Scheme, "https") {
		return false
	}
	_, trusted := r.trustedHosts[strings.ToLower(parsedURL.Host)]
	return trusted
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
