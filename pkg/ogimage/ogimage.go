// Package ogimage provides OGP (Open Graph Protocol) image generation.
//
// It renders an HTML template with article metadata (title, author, date, etc.)
// using a headless Chromium browser (via go-rod) and outputs a 1200×630 JPEG image.
package ogimage

// OGPData holds the values injected into the OGP HTML template.
// It is intentionally separate from core.Article so this package
// has no dependency on the core domain types.
type OGPData struct {
	Title    string
	Author   string
	Date     string
	Category string
	Tags     []string
}
