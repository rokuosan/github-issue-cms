package subcommand

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/rokuosan/github-issue-cms/pkg/config"
	"github.com/rokuosan/github-issue-cms/pkg/core"
	"github.com/spf13/cobra"
)

type contentOptions struct {
	token         string
	pageSize      int
	includeDrafts bool
}

type contentDependencies struct {
	load   func(context.Context, contentOptions) (*core.ContentSnapshot, error)
	export func(context.Context, *core.ContentSnapshot, string) error
	serve  func(context.Context, io.Writer, string, http.Handler) error
}

func defaultContentDependencies() contentDependencies {
	return contentDependencies{
		load:   loadContentSnapshot,
		export: core.ExportContentSnapshot,
		serve:  runContentServer,
	}
}

// NewServeCommand creates the content API server subcommand.
func NewServeCommand() *cobra.Command {
	return newServeCommand(defaultContentDependencies())
}

func newServeCommand(deps contentDependencies) *cobra.Command {
	var options contentOptions
	var address string
	cmd := &cobra.Command{
		Use:   "serve",
		Short: "Serve a snapshot of GitHub issues as a JSON content API",
		Long: `Fetch GitHub issues and attached images once at startup, then serve
the resulting snapshot without making GitHub requests for API calls.
Restart the server to refresh content. Draft articles are excluded by default.

Examples:
  github-issue-cms serve --token YOUR_GITHUB_TOKEN
  github-issue-cms serve --addr :8080 --page-size 20`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := options.validate(); err != nil {
				return err
			}
			if _, _, err := net.SplitHostPort(address); err != nil {
				return fmt.Errorf("invalid server address: %w", err)
			}
			ctx, stop := signal.NotifyContext(cmd.Context(), os.Interrupt, syscall.SIGTERM)
			defer stop()
			snapshot, err := deps.load(ctx, options)
			if err != nil {
				return err
			}
			return deps.serve(ctx, cmd.OutOrStdout(), address, core.NewContentHandler(snapshot))
		},
	}
	addContentFlags(cmd, &options)
	cmd.Flags().StringVar(&address, "addr", "localhost:8080", "HTTP listen address (use :8080 for deployment)")
	return cmd
}

// NewExportCommand creates the static content export subcommand.
func NewExportCommand() *cobra.Command {
	return newExportCommand(defaultContentDependencies())
}

func newExportCommand(deps contentDependencies) *cobra.Command {
	var options contentOptions
	var output string
	cmd := &cobra.Command{
		Use:   "export",
		Short: "Export GitHub issues as static JSON content and images",
		Long: `Fetch GitHub issues and attached images, then write a static content
snapshot with paginated article summaries, article details, metadata, and a
search index. Draft articles are excluded by default. Regenerate and redeploy
the output to refresh content.

Example:
  github-issue-cms export --output dist --page-size 20`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := options.validate(); err != nil {
				return err
			}
			if strings.TrimSpace(output) == "" {
				return fmt.Errorf("output directory must not be empty")
			}
			snapshot, err := deps.load(cmd.Context(), options)
			if err != nil {
				return err
			}
			if err := deps.export(cmd.Context(), snapshot, output); err != nil {
				return fmt.Errorf("export content: %w", err)
			}
			_, err = fmt.Fprintf(cmd.OutOrStdout(), "Content exported to %s\n", output)
			return err
		},
	}
	addContentFlags(cmd, &options)
	cmd.Flags().StringVar(&output, "output", "dist", "Directory for the static content snapshot")
	return cmd
}

func addContentFlags(cmd *cobra.Command, options *contentOptions) {
	cmd.Flags().StringVarP(&options.token, "token", "t", "", "GitHub API token (defaults to GITHUB_TOKEN)")
	cmd.Flags().IntVar(&options.pageSize, "page-size", 20, "Default articles per page (1-100)")
	cmd.Flags().BoolVar(&options.includeDrafts, "include-drafts", false, "Include draft articles in the published snapshot")
}

func (options *contentOptions) validate() error {
	if options.pageSize < 1 || options.pageSize > 100 {
		return fmt.Errorf("page size must be between 1 and 100")
	}
	if options.token == "" {
		options.token = os.Getenv("GITHUB_TOKEN")
	}
	if strings.TrimSpace(options.token) == "" {
		return fmt.Errorf("GitHub token is required; use --token or set GITHUB_TOKEN")
	}
	return nil
}

func loadContentSnapshot(ctx context.Context, options contentOptions) (*core.ContentSnapshot, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	conf, err := config.Get()
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}
	if conf.GitHub == nil || invalidContentRepositoryName(conf.GitHub.Username) || invalidContentRepositoryName(conf.GitHub.Repository) {
		return nil, fmt.Errorf("please set username and repository in gic.config.yaml; run 'github-issue-cms init' to create a config file")
	}
	store, err := core.NewGitHubIssueRepositoryWithLogger(options.token, slog.Default())
	if err != nil {
		return nil, fmt.Errorf("create GitHub issue repository: %w", err)
	}
	assets := core.NewHTTPImageRepositoryWithLogger(options.token, slog.Default())
	snapshot, err := core.LoadContentSnapshot(ctx, store, assets, conf, options.pageSize, options.includeDrafts)
	if err != nil {
		return nil, fmt.Errorf("load content snapshot: %w", err)
	}
	return snapshot, nil
}

func invalidContentRepositoryName(value string) bool {
	return strings.TrimSpace(value) == "" || strings.ContainsAny(value, "<>/")
}

func runContentServer(ctx context.Context, output io.Writer, address string, handler http.Handler) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	listener, err := net.Listen("tcp", address)
	if err != nil {
		return fmt.Errorf("listen for content server: %w", err)
	}
	server := &http.Server{
		Handler:           handler,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
	}
	serverErrors := make(chan error, 1)
	go func() { serverErrors <- server.Serve(listener) }()
	if _, err := fmt.Fprintf(output, "Content server: http://%s\n", listener.Addr()); err != nil {
		_ = server.Close()
		return fmt.Errorf("write content server address: %w", err)
	}
	select {
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := server.Shutdown(shutdownCtx); err != nil {
			_ = server.Close()
			return fmt.Errorf("shutdown content server: %w", err)
		}
		return nil
	case err := <-serverErrors:
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		return fmt.Errorf("content server: %w", err)
	}
}
