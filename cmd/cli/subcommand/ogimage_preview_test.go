package subcommand

import (
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewOGImagePreviewCommand(t *testing.T) {
	cmd := NewOGImagePreviewCommand()

	assert.Equal(t, "preview <template>", cmd.Use)
	assert.Contains(t, cmd.Short, "1200x630")
	assert.Equal(t, "localhost", cmd.Flag("host").DefValue)
	assert.Equal(t, "6140", cmd.Flag("port").DefValue)
}

func TestOGImagePreviewHandler_RendersTemplateAndAssets(t *testing.T) {
	dir := t.TempDir()
	templatePath := filepath.Join(dir, "ogp.html")
	require.NoError(t, os.WriteFile(templatePath, []byte(`<!doctype html><html><head></head><body><h1>{{ .Title }}</h1><img src="logo.svg"></body></html>`), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "logo.svg"), []byte("<svg></svg>"), 0o644))

	server := httptest.NewServer(newOGImagePreviewHandler(templatePath))
	t.Cleanup(server.Close)

	response, err := http.Get(server.URL + "/__preview")
	require.NoError(t, err)
	t.Cleanup(func() { _ = response.Body.Close() })
	assert.Equal(t, http.StatusOK, response.StatusCode)

	body := readResponseBody(t, response)
	assert.Contains(t, body, "OGP template preview")
	assert.Contains(t, body, `<base href="/__asset/">`)
	assert.NotContains(t, body, `\\n<base`)

	assetResponse, err := http.Get(server.URL + "/__asset/logo.svg")
	require.NoError(t, err)
	t.Cleanup(func() { _ = assetResponse.Body.Close() })
	assert.Equal(t, http.StatusOK, assetResponse.StatusCode)
	assert.Equal(t, "<svg></svg>", readResponseBody(t, assetResponse))
}

func TestOGImagePreviewHandler_ShowsTemplateErrors(t *testing.T) {
	templatePath := filepath.Join(t.TempDir(), "ogp.html")
	require.NoError(t, os.WriteFile(templatePath, []byte(`{{ if }}`), 0o644))

	request := httptest.NewRequest(http.MethodGet, "/__preview", nil)
	recorder := httptest.NewRecorder()
	newOGImagePreviewHandler(templatePath).ServeHTTP(recorder, request)

	assert.Equal(t, http.StatusUnprocessableEntity, recorder.Code)
	assert.Contains(t, recorder.Body.String(), "parse template")
	assert.Equal(t, "text/html; charset=utf-8", recorder.Header().Get("Content-Type"))
}

func TestOGImagePreviewHandler_Version(t *testing.T) {
	templatePath := filepath.Join(t.TempDir(), "ogp.html")
	require.NoError(t, os.WriteFile(templatePath, []byte("template"), 0o644))

	handler := newOGImagePreviewHandler(templatePath)
	first := httptest.NewRecorder()
	handler.ServeHTTP(first, httptest.NewRequest(http.MethodGet, "/__preview/version", nil))

	require.NoError(t, os.WriteFile(templatePath, []byte("updated template"), 0o644))
	second := httptest.NewRecorder()
	handler.ServeHTTP(second, httptest.NewRequest(http.MethodGet, "/__preview/version", nil))

	assert.NotEmpty(t, first.Body.String())
	assert.NotEqual(t, first.Body.String(), second.Body.String())
}

func TestInjectAssetBase_DoesNotOverrideExistingBase(t *testing.T) {
	document := `<html><head><base href="https://example.com/"></head></html>`

	assert.Equal(t, document, injectAssetBase(document))
}

func readResponseBody(t *testing.T, response *http.Response) string {
	t.Helper()
	body, err := io.ReadAll(response.Body)
	require.NoError(t, err)
	return string(body)
}
