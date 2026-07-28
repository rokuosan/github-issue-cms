package ogimage

import (
	"bytes"
	"context"
	_ "embed"
	"fmt"
	"html/template"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
	"github.com/go-rod/rod/lib/proto"
)

// DefaultTemplate is the embedded default OGP HTML template.
//
//go:embed template.html
var DefaultTemplate string

// Renderer is an OGP image renderer that uses a headless Chromium browser.
// It loads an HTML template, expands it with OGPData, renders it in a headless
// browser, and captures a JPEG screenshot at 1200×630.
type Renderer struct {
	tmpl       *template.Template
	browserBin string
	tmplDir    string // directory of custom template (empty for embedded)
}

// NewRenderer creates a Renderer using the embedded default template.
// If browserBin is empty, go-rod's launcher auto-downloads and caches Chromium.
// The GIC_CHROMIUM_BIN environment variable takes precedence over browserBin.
func NewRenderer(browserBin string) (*Renderer, error) {
	tmpl, err := template.New("ogp").Parse(DefaultTemplate)
	if err != nil {
		return nil, fmt.Errorf("ogimage: parse default template: %w", err)
	}
	return &Renderer{tmpl: tmpl, browserBin: browserBin}, nil
}

// NewRendererWithTemplate creates a Renderer from a custom template file.
// If browserBin is empty, go-rod's launcher auto-downloads and caches Chromium.
// The template file's directory is preserved so relative asset references
// (images, stylesheets) in the template resolve correctly.
func NewRendererWithTemplate(tmplPath string, browserBin string) (*Renderer, error) {
	tmpl, err := template.ParseFiles(tmplPath)
	if err != nil {
		return nil, fmt.Errorf("ogimage: parse template %s: %w", tmplPath, err)
	}
	absPath, err := filepath.Abs(tmplPath)
	if err != nil {
		absPath = tmplPath
	}
	return &Renderer{
		tmpl:       tmpl,
		browserBin: browserBin,
		tmplDir:    filepath.Dir(absPath),
	}, nil
}

// Render executes the template with the given data, opens a headless browser
// page, and captures a 1200×630 JPEG screenshot.
func (r *Renderer) Render(ctx context.Context, data OGPData) ([]byte, error) {
	html, err := r.executeTemplate(data)
	if err != nil {
		return nil, err
	}

	browser, err := r.launchBrowser(ctx)
	if err != nil {
		return nil, fmt.Errorf("ogimage: launch browser: %w", err)
	}
	defer closeBrowser(browser)

	pageURL := "about:blank"
	if r.tmplDir != "" {
		// Use a file:// URL based on the template directory so relative
		// asset references (e.g. <img src="logo.png">) resolve correctly.
		pageURL = "file://" + filepath.ToSlash(r.tmplDir) + "/"
	}

	page, err := browser.Page(proto.TargetCreateTarget{URL: pageURL})
	if err != nil {
		return nil, fmt.Errorf("ogimage: create page: %w", err)
	}
	defer closePage(page)

	if err := page.SetViewport(&proto.EmulationSetDeviceMetricsOverride{
		Width:             1200,
		Height:            630,
		DeviceScaleFactor: 1,
		Mobile:            false,
	}); err != nil {
		return nil, fmt.Errorf("ogimage: set viewport: %w", err)
	}

	if err := page.SetDocumentContent(html); err != nil {
		return nil, fmt.Errorf("ogimage: set content: %w", err)
	}

	if err := page.WaitStable(1 * time.Second); err != nil {
		return nil, fmt.Errorf("ogimage: wait stable: %w", err)
	}

	quality := 90
	screenshot, err := page.Screenshot(false, &proto.PageCaptureScreenshot{
		Format:  proto.PageCaptureScreenshotFormatJpeg,
		Quality: &quality,
		Clip: &proto.PageViewport{
			X:      0,
			Y:      0,
			Width:  1200,
			Height: 630,
			Scale:  1,
		},
	})
	if err != nil {
		return nil, fmt.Errorf("ogimage: screenshot: %w", err)
	}

	return screenshot, nil
}

func (r *Renderer) executeTemplate(data OGPData) (string, error) {
	var buf bytes.Buffer
	if err := r.tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("ogimage: execute template: %w", err)
	}
	return buf.String(), nil
}

func (r *Renderer) launchBrowser(ctx context.Context) (*rod.Browser, error) {
	bin := r.resolveChromiumBin()

	var l *launcher.Launcher
	if bin != "" {
		l = launcher.New().Bin(bin)
	} else {
		l = launcher.New()
	}

	l = l.Headless(true).NoSandbox(true).Leakless(false).Context(ctx)
	url, err := l.Launch()
	if err != nil {
		return nil, fmt.Errorf("launch chromium: %w (hint: set GIC_CHROMIUM_BIN to an explicit chromium binary path)", err)
	}

	browser := rod.New().ControlURL(url).Context(ctx)
	if err := browser.Connect(); err != nil {
		l.Kill()
		return nil, fmt.Errorf("connect to browser: %w", err)
	}

	return browser, nil
}

// resolveChromiumBin returns the Chromium binary path to use.
// Priority: GIC_CHROMIUM_BIN env var > Renderer.browserBin > empty (auto).
func (r *Renderer) resolveChromiumBin() string {
	if bin := strings.TrimSpace(os.Getenv("GIC_CHROMIUM_BIN")); bin != "" {
		return bin
	}
	return r.browserBin
}

// closeBrowser closes the browser and logs any error (panics from MustClose
// are avoided by using the error-returning Close API).
func closeBrowser(browser *rod.Browser) {
	if err := browser.Close(); err != nil {
		// Browser close failure is non-fatal during cleanup.
		_ = err
	}
}

// closePage closes the page and logs any error.
func closePage(page *rod.Page) {
	if err := page.Close(); err != nil {
		// Page close failure is non-fatal during cleanup.
		_ = err
	}
}
