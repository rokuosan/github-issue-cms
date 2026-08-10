package subcommand

import (
	"bytes"
	"context"
	"fmt"
	"html"
	"html/template"
	"io"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/rokuosan/github-issue-cms/pkg/ogimage"
	"github.com/spf13/cobra"
)

// NewOGImagePreviewCommand creates the template preview subcommand.
func NewOGImagePreviewCommand() *cobra.Command {
	var (
		host string
		port int
	)

	cmd := &cobra.Command{
		Use:   "preview <template>",
		Short: "Preview an OGP HTML template in a fixed 1200x630 viewport",
		Long: `Start a local live preview server for an OGP HTML template.

The template is rendered with sample OGP data and displayed at the fixed
1200x630 OGP size. The preview reloads automatically when the template file
changes. Files referenced with relative URLs are served from the template's
directory.

Example:
  github-issue-cms ogimage preview ogp.html`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runOGImagePreview(cmd.Context(), cmd.OutOrStdout(), args[0], host, port)
		},
	}

	cmd.Flags().StringVar(&host, "host", "localhost", "Host to bind the preview server to")
	cmd.Flags().IntVarP(&port, "port", "p", 6140, "Port to bind the preview server to")

	return cmd
}

func runOGImagePreview(ctx context.Context, output io.Writer, templatePath, host string, port int) error {
	if port < 0 || port > 65535 {
		return fmt.Errorf("invalid port %d: must be between 0 and 65535", port)
	}

	absTemplatePath, err := filepath.Abs(templatePath)
	if err != nil {
		return fmt.Errorf("resolve template path: %w", err)
	}
	info, err := os.Stat(absTemplatePath)
	if err != nil {
		return fmt.Errorf("stat template %s: %w", templatePath, err)
	}
	if info.IsDir() {
		return fmt.Errorf("template path is a directory: %s", templatePath)
	}

	listener, err := net.Listen("tcp", net.JoinHostPort(host, strconv.Itoa(port)))
	if err != nil {
		return fmt.Errorf("listen for preview server: %w", err)
	}

	server := &http.Server{
		Handler:           newOGImagePreviewHandler(absTemplatePath),
		ReadHeaderTimeout: 5 * time.Second,
	}
	serverErrors := make(chan error, 1)
	go func() {
		serverErrors <- server.Serve(listener)
	}()

	actualPort := listener.Addr().(*net.TCPAddr).Port
	displayHost := host
	if displayHost == "" {
		displayHost = "localhost"
	}
	address := net.JoinHostPort(displayHost, strconv.Itoa(actualPort))
	if _, err := fmt.Fprintf(output, "OGP preview server: http://%s\n", address); err != nil {
		_ = server.Close()
		return fmt.Errorf("write preview server address: %w", err)
	}

	select {
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := server.Shutdown(shutdownCtx); err != nil {
			return fmt.Errorf("shutdown preview server: %w", err)
		}
		return ctx.Err()
	case err := <-serverErrors:
		if err == http.ErrServerClosed {
			return nil
		}
		return fmt.Errorf("preview server: %w", err)
	}
}

type ogImagePreviewHandler struct {
	templatePath string
	templateDir  string
}

func newOGImagePreviewHandler(templatePath string) http.Handler {
	return &ogImagePreviewHandler{
		templatePath: templatePath,
		templateDir:  filepath.Dir(templatePath),
	}
}

func (h *ogImagePreviewHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch {
	case r.URL.Path == "/":
		h.serveShell(w)
	case r.URL.Path == "/__preview":
		h.serveTemplate(w)
	case r.URL.Path == "/__preview/version":
		h.serveVersion(w)
	case strings.HasPrefix(r.URL.Path, "/__asset/"):
		h.serveAsset(w, r)
	default:
		http.NotFound(w, r)
	}
}

func (h *ogImagePreviewHandler) serveShell(w http.ResponseWriter) {
	name := html.EscapeString(filepath.Base(h.templatePath))
	shell := fmt.Sprintf(previewShell, name, name)
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = w.Write([]byte(shell))
}

func (h *ogImagePreviewHandler) serveTemplate(w http.ResponseWriter) {
	tmpl, err := template.ParseFiles(h.templatePath)
	if err != nil {
		h.serveTemplateError(w, fmt.Errorf("parse template: %w", err))
		return
	}

	var rendered bytes.Buffer
	if err := tmpl.Execute(&rendered, previewOGPData()); err != nil {
		h.serveTemplateError(w, fmt.Errorf("execute template: %w", err))
		return
	}

	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = w.Write([]byte(injectAssetBase(rendered.String())))
}

func (h *ogImagePreviewHandler) serveTemplateError(w http.ResponseWriter, err error) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusUnprocessableEntity)
	message := html.EscapeString(err.Error())
	_, _ = fmt.Fprintf(w, `<!doctype html><meta charset="utf-8"><style>body{margin:0;padding:32px;background:#1f1f1f;color:#ffb4ab;font:16px/1.5 ui-monospace,SFMono-Regular,Menlo,monospace;white-space:pre-wrap}</style>%s`, message)
}

func (h *ogImagePreviewHandler) serveVersion(w http.ResponseWriter) {
	info, err := os.Stat(h.templatePath)
	version := "missing"
	if err == nil {
		version = fmt.Sprintf("%d-%d", info.ModTime().UnixNano(), info.Size())
	}
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	_, _ = w.Write([]byte(version))
}

func (h *ogImagePreviewHandler) serveAsset(w http.ResponseWriter, r *http.Request) {
	assetServer := http.StripPrefix("/__asset/", http.FileServer(http.Dir(h.templateDir)))
	assetServer.ServeHTTP(w, r)
}

func previewOGPData() ogimage.OGPData {
	return ogimage.OGPData{
		Title:    "OGP template preview",
		Author:   "github-issue-cms",
		Date:     "2026-01-01",
		Category: "Example",
		Tags:     []string{"preview", "ogp"},
	}
}

func injectAssetBase(document string) string {
	lower := strings.ToLower(document)
	if strings.Contains(lower, "<base ") {
		return document
	}

	if headStart := strings.Index(lower, "<head"); headStart >= 0 {
		if headEnd := strings.Index(lower[headStart:], ">"); headEnd >= 0 {
			insertAt := headStart + headEnd + 1
			return document[:insertAt] + "\n<base href=\"/__asset/\">" + document[insertAt:]
		}
	}
	return `<base href="/__asset/">` + document
}

const previewShell = `<!doctype html>
<html lang="en">
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<title>OGP preview: %s</title>
<style>
  :root { color-scheme: dark; }
  * { box-sizing: border-box; }
  html, body { min-height: 100%%; }
  body {
    margin: 0;
    min-width: 320px;
    background: #202124;
    color: #e8eaed;
    font: 14px/1.4 -apple-system, BlinkMacSystemFont, "Segoe UI", sans-serif;
  }
  header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 16px;
    min-height: 52px;
    padding: 12px 20px;
    border-bottom: 1px solid #3c4043;
  }
  header strong { font-weight: 600; }
  header span { color: #9aa0a6; font-size: 12px; }
  main {
    display: grid;
    min-height: calc(100vh - 53px);
    place-items: center;
    padding: 16px;
  }
  .stage {
    position: relative;
    width: min(1200px, calc(100vw - 32px));
    aspect-ratio: 1200 / 630;
    overflow: hidden;
    background: #fff;
    box-shadow: 0 12px 40px rgb(0 0 0 / 35%%);
  }
  iframe {
    position: absolute;
    top: 0;
    left: 0;
    width: 1200px;
    height: 630px;
    border: 0;
    transform-origin: top left;
  }
</style>
</head>
<body>
<header><strong>OGP preview</strong><span>%s · 1200 × 630</span></header>
<main><div class="stage"><iframe id="preview" title="OGP template preview"></iframe></div></main>
<script>
const frame = document.getElementById('preview');
const stage = document.querySelector('.stage');
let version = '';

function resizePreview() {
  frame.style.transform = 'scale(' + (stage.clientWidth / 1200) + ')';
}

function reloadPreview(nextVersion) {
  version = nextVersion;
  frame.src = '/__preview?version=' + encodeURIComponent(nextVersion);
}

async function checkForChanges() {
  try {
    const response = await fetch('/__preview/version', { cache: 'no-store' });
    const nextVersion = await response.text();
    if (nextVersion !== version) reloadPreview(nextVersion);
  } catch (_) {
    // The server may be restarting while the file is being edited.
  }
}

window.addEventListener('resize', resizePreview);
resizePreview();
checkForChanges();
setInterval(checkForChanges, 500);
</script>
</body>
</html>`
