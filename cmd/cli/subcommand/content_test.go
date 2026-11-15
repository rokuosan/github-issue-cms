package subcommand

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/rokuosan/github-issue-cms/pkg/core"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestContentCommandsValidateBeforeLoading(t *testing.T) {
	t.Setenv("GITHUB_TOKEN", "")
	for _, command := range []string{"serve", "export"} {
		for _, tc := range []struct {
			name string
			args []string
			want string
		}{
			{"missing token", nil, "GitHub token is required"},
			{"invalid page size", []string{"--token", "token", "--page-size", "0"}, "page size must"},
			{"large page size", []string{"--token", "token", "--page-size", "101"}, "page size must"},
			{"positional argument", []string{"extra"}, "unknown command"},
		} {
			t.Run(command+"/"+tc.name, func(t *testing.T) {
				deps := contentDependencies{load: func(context.Context, contentOptions) (*core.ContentSnapshot, error) {
					t.Fatal("invalid command must not load content")
					return nil, nil
				}}
				var cmd *cobra.Command
				if command == "serve" {
					cmd = newServeCommand(deps)
				} else {
					cmd = newExportCommand(deps)
				}
				cmd.SetOut(io.Discard)
				cmd.SetErr(io.Discard)
				cmd.SetArgs(tc.args)
				err := cmd.Execute()
				require.Error(t, err)
				assert.Contains(t, err.Error(), tc.want)
			})
		}
	}
}

func TestExportCommandLoadsOnceAndPassesOptions(t *testing.T) {
	t.Setenv("GITHUB_TOKEN", "environment-token")
	for _, explicitToken := range []bool{false, true} {
		t.Run(fmt.Sprint(explicitToken), func(t *testing.T) {
			loads := 0
			snapshot := &core.ContentSnapshot{}
			deps := contentDependencies{
				load: func(ctx context.Context, options contentOptions) (*core.ContentSnapshot, error) {
					loads++
					wantToken := "environment-token"
					if explicitToken {
						wantToken = "flag-token"
					}
					assert.Equal(t, wantToken, options.token)
					assert.Equal(t, 7, options.pageSize)
					assert.True(t, options.includeDrafts)
					return snapshot, nil
				},
				export: func(ctx context.Context, got *core.ContentSnapshot, output string) error {
					assert.Same(t, snapshot, got)
					assert.Equal(t, "static-content", output)
					return nil
				},
			}
			cmd := newExportCommand(deps)
			args := []string{"--page-size", "7", "--include-drafts", "--output", "static-content"}
			if explicitToken {
				args = append(args, "--token", "flag-token")
			}
			var output bytes.Buffer
			cmd.SetOut(&output)
			cmd.SetArgs(args)
			require.NoError(t, cmd.Execute())
			assert.Equal(t, 1, loads)
			assert.Contains(t, output.String(), "Content exported to static-content")
		})
	}
}

func TestContentCommandsPropagateLoadFailure(t *testing.T) {
	t.Setenv("GITHUB_TOKEN", "token")
	want := errors.New("GitHub unavailable")
	deps := contentDependencies{load: func(context.Context, contentOptions) (*core.ContentSnapshot, error) {
		return nil, want
	}}
	for _, cmd := range []*cobra.Command{newServeCommand(deps), newExportCommand(deps)} {
		cmd.SetOut(io.Discard)
		cmd.SetErr(io.Discard)
		cmd.SetArgs([]string{})
		assert.ErrorIs(t, cmd.Execute(), want)
	}
}

func TestServeCommandLoadsOnceBeforeServing(t *testing.T) {
	t.Setenv("GITHUB_TOKEN", "token")
	loads := 0
	deps := contentDependencies{
		load: func(context.Context, contentOptions) (*core.ContentSnapshot, error) {
			loads++
			return core.NewContentSnapshot(nil, 20)
		},
		serve: func(ctx context.Context, output io.Writer, address string, handler http.Handler) error {
			assert.Equal(t, 1, loads)
			assert.Equal(t, ":9000", address)
			assert.NotNil(t, handler)
			return nil
		},
	}
	cmd := newServeCommand(deps)
	cmd.SetArgs([]string{"--addr", ":9000"})
	require.NoError(t, cmd.Execute())
	assert.Equal(t, 1, loads)
}

func TestExportCommandPropagatesWriteFailure(t *testing.T) {
	t.Setenv("GITHUB_TOKEN", "token")
	want := errors.New("disk full")
	cmd := newExportCommand(contentDependencies{
		load: func(context.Context, contentOptions) (*core.ContentSnapshot, error) {
			return &core.ContentSnapshot{}, nil
		},
		export: func(context.Context, *core.ContentSnapshot, string) error { return want },
	})
	cmd.SetOut(io.Discard)
	cmd.SetErr(io.Discard)
	cmd.SetArgs([]string{})
	assert.ErrorIs(t, cmd.Execute(), want)
}

type serverAddressWriter struct {
	address chan string
}

func (w serverAddressWriter) Write(data []byte) (int, error) {
	w.address <- strings.TrimSpace(strings.TrimPrefix(string(data), "Content server: "))
	return len(data), nil
}

func TestRunContentServerServesAndShutsDown(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	addresses := make(chan string, 1)
	done := make(chan error, 1)
	go func() {
		done <- runContentServer(ctx, serverAddressWriter{addresses}, "127.0.0.1:0", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_, _ = io.WriteString(w, "snapshot")
		}))
	}()
	var address string
	select {
	case address = <-addresses:
	case <-time.After(5 * time.Second):
		t.Fatal("server did not start")
	}
	client := &http.Client{Timeout: 2 * time.Second}
	resp, err := client.Get(address)
	require.NoError(t, err)
	body, err := io.ReadAll(resp.Body)
	require.NoError(t, resp.Body.Close())
	require.NoError(t, err)
	assert.Equal(t, "snapshot", string(body))
	cancel()
	select {
	case err := <-done:
		require.NoError(t, err)
	case <-time.After(6 * time.Second):
		t.Fatal("server did not shut down")
	}
	_, err = client.Get(address)
	assert.Error(t, err)
}

func TestRunContentServerFailsToListen(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	defer listener.Close()
	err = runContentServer(context.Background(), io.Discard, listener.Addr().String(), http.NotFoundHandler())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "listen for content server")
}

func TestRunContentServerCancelledBeforeStartup(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	assert.ErrorIs(t, runContentServer(ctx, io.Discard, "127.0.0.1:0", http.NotFoundHandler()), context.Canceled)
}

func TestLoadContentSnapshotHonorsCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err := loadContentSnapshot(ctx, contentOptions{})
	assert.ErrorIs(t, err, context.Canceled)
}

func TestValidateContentOrigin(t *testing.T) {
	for _, origin := range []string{"", "*", "https://example.com", "http://localhost:3000", "http://[::1]:3000"} {
		assert.NoError(t, validateContentOrigin(origin), origin)
	}
	for _, origin := range []string{
		"null", "example.com", "ftp://example.com", "https://example.com/", "https://example.com/path",
		"https://example.com?q=1", "https://example.com?", "https://example.com#fragment", "https://example.com#",
		"https://user:secret@example.com", "http://:3000", "https://example.com:bad", "https://example.com\r\nX: injected",
	} {
		assert.Error(t, validateContentOrigin(origin), origin)
	}
}

func TestServeCommandRejectsInvalidCORSBeforeLoading(t *testing.T) {
	t.Setenv("GITHUB_TOKEN", "token")
	cmd := newServeCommand(contentDependencies{load: func(context.Context, contentOptions) (*core.ContentSnapshot, error) {
		t.Fatal("invalid CORS option must fail before fetching GitHub")
		return nil, nil
	}})
	cmd.SetArgs([]string{"--cors-origin", "https://example.com/path"})
	cmd.SetOut(io.Discard)
	cmd.SetErr(io.Discard)
	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "cors origin")
}

func TestContentCORS(t *testing.T) {
	for _, tc := range []struct {
		name, configured, origin, method, requestedMethod string
		wantOrigin, wantMethods, wantVary                 string
		wantStatus                                        int
	}{
		{name: "disabled", origin: "https://site.example", method: "GET", wantStatus: 200},
		{name: "matching origin", configured: "https://site.example", origin: "https://site.example", method: "GET", wantOrigin: "https://site.example", wantVary: "Origin", wantStatus: 200},
		{name: "matching HEAD", configured: "https://site.example", origin: "https://site.example", method: "HEAD", wantOrigin: "https://site.example", wantVary: "Origin", wantStatus: 200},
		{name: "different origin", configured: "https://site.example", origin: "https://other.example", method: "GET", wantVary: "Origin", wantStatus: 200},
		{name: "no origin", configured: "https://site.example", method: "GET", wantVary: "Origin", wantStatus: 200},
		{name: "wildcard", configured: "*", origin: "https://site.example", method: "GET", wantOrigin: "*", wantStatus: 200},
		{name: "wildcard without origin", configured: "*", method: "GET", wantStatus: 200},
		{name: "preflight", configured: "https://site.example", origin: "https://site.example", method: "OPTIONS", requestedMethod: "GET", wantOrigin: "https://site.example", wantMethods: "GET, HEAD", wantVary: "Origin", wantStatus: 204},
		{name: "wildcard HEAD preflight", configured: "*", origin: "https://site.example", method: "OPTIONS", requestedMethod: "HEAD", wantOrigin: "*", wantMethods: "GET, HEAD", wantStatus: 204},
		{name: "preflight mismatch", configured: "https://site.example", origin: "https://other.example", method: "OPTIONS", requestedMethod: "GET", wantVary: "Origin", wantStatus: 405},
		{name: "unsupported preflight method", configured: "*", origin: "https://site.example", method: "OPTIONS", requestedMethod: "POST", wantStatus: 405},
		{name: "plain options", configured: "*", origin: "https://site.example", method: "OPTIONS", wantStatus: 405},
		{name: "unsupported method", configured: "*", origin: "https://site.example", method: "POST", wantStatus: 405},
	} {
		t.Run(tc.name, func(t *testing.T) {
			handler := withContentCORS(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodGet && r.Method != http.MethodHead {
					w.Header().Set("Allow", "GET, HEAD")
					w.WriteHeader(http.StatusMethodNotAllowed)
				}
			}), tc.configured)
			req := httptest.NewRequest(tc.method, "/data/articles", nil)
			req.Header.Set("Origin", tc.origin)
			req.Header.Set("Access-Control-Request-Method", tc.requestedMethod)
			res := httptest.NewRecorder()
			handler.ServeHTTP(res, req)
			assert.Equal(t, tc.wantStatus, res.Code)
			assert.Equal(t, tc.wantOrigin, res.Header().Get("Access-Control-Allow-Origin"))
			assert.Equal(t, tc.wantMethods, res.Header().Get("Access-Control-Allow-Methods"))
			assert.Equal(t, tc.wantVary, res.Header().Get("Vary"))
			assert.Empty(t, res.Header().Get("Access-Control-Allow-Credentials"))
		})
	}
}
