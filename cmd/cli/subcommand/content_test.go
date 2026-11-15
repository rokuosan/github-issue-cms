package subcommand

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
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
