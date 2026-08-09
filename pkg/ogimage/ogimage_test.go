package ogimage

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOGPData(t *testing.T) {
	data := OGPData{
		Title:    "Test Title",
		Author:   "testuser",
		Date:     "2024-01-15",
		Category: "tech",
		Tags:     []string{"go", "testing"},
	}

	assert.Equal(t, "Test Title", data.Title)
	assert.Equal(t, "testuser", data.Author)
	assert.Equal(t, "2024-01-15", data.Date)
	assert.Equal(t, "tech", data.Category)
	assert.Equal(t, []string{"go", "testing"}, data.Tags)
}

func TestNewRenderer(t *testing.T) {
	r, err := NewRenderer("")
	require.NoError(t, err)
	require.NotNil(t, r)
	assert.NotNil(t, r.tmpl)
}

func TestNewRenderer_WithBin(t *testing.T) {
	r, err := NewRenderer("/usr/bin/chromium-browser")
	require.NoError(t, err)
	require.NotNil(t, r)
	assert.Equal(t, "/usr/bin/chromium-browser", r.browserBin)
}

func TestNewRendererWithTemplate(t *testing.T) {
	// Create a temporary custom template.
	dir := t.TempDir()
	tmplPath := filepath.Join(dir, "custom.html")
	err := os.WriteFile(tmplPath, []byte(`<html><body>{{ .Title }}</body></html>`), 0o644)
	require.NoError(t, err)

	r, err := NewRendererWithTemplate(tmplPath, "")
	require.NoError(t, err)
	require.NotNil(t, r)

	html, err := r.executeTemplate(OGPData{Title: "Hello"})
	require.NoError(t, err)
	assert.Contains(t, html, "Hello")
}

func TestNewRendererWithTemplate_NotFound(t *testing.T) {
	_, err := NewRendererWithTemplate("/nonexistent/template.html", "")
	assert.Error(t, err)
}

func TestExecuteTemplate(t *testing.T) {
	r, err := NewRenderer("")
	require.NoError(t, err)

	data := OGPData{
		Title:    "My Article",
		Author:   "alice",
		Date:     "2024-01-15",
		Category: "Tech",
		Tags:     []string{"go", "cli"},
	}

	html, err := r.executeTemplate(data)
	require.NoError(t, err)

	// The output should contain the injected values.
	assert.Contains(t, html, "My Article")
	assert.Contains(t, html, "alice")
	assert.Contains(t, html, "2024-01-15")
	assert.Contains(t, html, "Tech")
	assert.Contains(t, html, "go")
	assert.Contains(t, html, "cli")
}

func TestExecuteTemplate_EmptyData(t *testing.T) {
	r, err := NewRenderer("")
	require.NoError(t, err)

	html, err := r.executeTemplate(OGPData{})
	require.NoError(t, err)

	// Should render without errors even with empty data.
	assert.NotEmpty(t, html)
	assert.Contains(t, html, "<html")
}

func TestExecuteTemplate_SpecialChars(t *testing.T) {
	r, err := NewRenderer("")
	require.NoError(t, err)

	data := OGPData{
		Title:    "Title with special chars & unicode: 日本語",
		Author:   "user",
		Tags:     []string{"tag-1", "tag-2"},
	}

	html, err := r.executeTemplate(data)
	require.NoError(t, err)

	// Verify template renders correctly with special characters.
	assert.Contains(t, html, "Title with special chars")
	assert.Contains(t, html, "日本語")
	assert.Contains(t, html, "tag-1")
	assert.Contains(t, html, "tag-2")
}

func TestResolveChromiumBin(t *testing.T) {
	t.Run("env var takes precedence", func(t *testing.T) {
		t.Setenv("GIC_CHROMIUM_BIN", "/opt/chromium/chrome")
		r := &Renderer{browserBin: "/usr/bin/chrome"}
		assert.Equal(t, "/opt/chromium/chrome", r.resolveChromiumBin())
	})

	t.Run("falls back to browserBin", func(t *testing.T) {
		r := &Renderer{browserBin: "/usr/bin/chrome"}
		assert.Equal(t, "/usr/bin/chrome", r.resolveChromiumBin())
	})

	t.Run("empty when neither is set", func(t *testing.T) {
		r := &Renderer{}
		assert.Equal(t, "", r.resolveChromiumBin())
	})

	t.Run("env var with whitespace is trimmed", func(t *testing.T) {
		t.Setenv("GIC_CHROMIUM_BIN", "  /usr/bin/chromium  ")
		r := &Renderer{browserBin: "/usr/bin/chrome"}
		assert.Equal(t, "/usr/bin/chromium", r.resolveChromiumBin())
	})
}

func TestDefaultTemplate_Embedded(t *testing.T) {
	assert.NotEmpty(t, DefaultTemplate)
	assert.True(t, strings.Contains(DefaultTemplate, "<html"), "default template should be HTML")
	assert.True(t, strings.Contains(DefaultTemplate, "{{ .Title }}"), "template should have Title placeholder")
	assert.True(t, strings.Contains(DefaultTemplate, "{{ .Author }}"), "template should have Author placeholder")
}

// TestRenderIntegration performs a full render with a real browser.
// This is skipped by default because it requires Chromium.
func TestRenderIntegration(t *testing.T) {
	if os.Getenv("GIC_INTEGRATION_TEST") == "" {
		t.Skip("Skipping integration test: set GIC_INTEGRATION_TEST=1 to run")
	}

	r, err := NewRenderer("")
	require.NoError(t, err)
	require.NoError(t, r.Open(t.Context()))
	t.Cleanup(r.Close)
	browser := r.browser

	data := OGPData{
		Title:    "Integration Test Article",
		Author:   "testuser",
		Date:     "2024-01-15",
		Category: "Testing",
		Tags:     []string{"go", "integration"},
	}

	for range 2 {
		jpeg, err := r.Render(t.Context(), data)
		require.NoError(t, err)
		require.NotEmpty(t, jpeg)

		// JPEG files start with FF D8 FF.
		assert.Equal(t, byte(0xFF), jpeg[0])
		assert.Equal(t, byte(0xD8), jpeg[1])
		assert.Equal(t, byte(0xFF), jpeg[2])
		assert.Same(t, browser, r.browser)
	}
}
