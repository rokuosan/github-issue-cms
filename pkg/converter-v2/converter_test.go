package converter_v2

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/go-github/v67/github"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_NewConverter(t *testing.T) {
	tests := []struct {
		name    string
		options []Option
		assert  func(t *testing.T, c Converter)
	}{
		{
			name:    "with token",
			options: []Option{WithToken("test-token")},
			assert: func(t *testing.T, c Converter) {
				impl := c.(*converterImpl)
				assert.Equal(t, "test-token", impl.Token)
			},
		},
		{
			name:    "without token",
			options: []Option{},
			assert: func(t *testing.T, c Converter) {
				impl := c.(*converterImpl)
				assert.Equal(t, "", impl.Token)
			},
		},
		{
			name:    "with repository",
			options: []Option{WithRepository("owner", "repo")},
			assert: func(t *testing.T, c Converter) {
				impl := c.(*converterImpl)
				assert.Equal(t, "owner", impl.Username)
				assert.Equal(t, "repo", impl.Repository)
			},
		},
		{
			name:    "with http client",
			options: []Option{WithHTTPClient(http.DefaultClient)},
			assert: func(t *testing.T, c Converter) {
				impl := c.(*converterImpl)
				assert.NotNil(t, impl.http)
			},
		},
		{
			name:    "with github client",
			options: []Option{WithGitHubClient(github.NewClient(nil))},
			assert: func(t *testing.T, c Converter) {
				impl := c.(*converterImpl)
				assert.NotNil(t, impl.github)
			},
		},
		{
			name: "with all options",
			options: []Option{
				WithToken("test-token"),
				WithRepository("owner", "repo"),
				WithHTTPClient(http.DefaultClient),
			},
			assert: func(t *testing.T, c Converter) {
				impl := c.(*converterImpl)
				assert.Equal(t, "test-token", impl.Token)
				assert.Equal(t, "owner", impl.Username)
				assert.Equal(t, "repo", impl.Repository)
				assert.NotNil(t, impl.http)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewConverter(tt.options...)
			tt.assert(t, got)
		})
	}
}

func Test_CheckRequirements(t *testing.T) {
	tests := []struct {
		name    string
		options []Option
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid requirements",
			options: []Option{
				WithToken("test-token"),
				WithRepository("owner", "repo"),
			},
			wantErr: false,
		},
		{
			name:    "missing token",
			options: []Option{WithRepository("owner", "repo")},
			wantErr: true,
			errMsg:  "missing token",
		},
		{
			name:    "missing repository",
			options: []Option{WithToken("test-token")},
			wantErr: true,
			errMsg:  "missing repository information",
		},
		{
			name: "missing repository name",
			options: []Option{
				WithToken("test-token"),
				WithRepository("owner", ""),
			},
			wantErr: true,
			errMsg:  "missing repository information",
		},
		{
			name: "missing owner",
			options: []Option{
				WithToken("test-token"),
				WithRepository("", "repo"),
			},
			wantErr: true,
			errMsg:  "missing repository information",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := NewConverter(tt.options...).(*converterImpl)
			err := c.CheckRequirements()
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func Test_HTTPClient(t *testing.T) {
	customClient := &http.Client{}
	c := NewConverter(WithHTTPClient(customClient)).(*converterImpl)
	assert.Equal(t, customClient, c.HTTPClient())
}

func Test_GitHubClient(t *testing.T) {
	t.Run("with custom client", func(t *testing.T) {
		customClient := github.NewClient(nil)
		c := NewConverter(WithGitHubClient(customClient)).(*converterImpl)
		assert.Equal(t, customClient, c.GitHubClient())
	})

	t.Run("with default client", func(t *testing.T) {
		c := NewConverter().(*converterImpl)
		client := c.GitHubClient()
		assert.NotNil(t, client)
	})
}

func Test_WalkIssues(t *testing.T) {
	tests := []struct {
		name           string
		options        []Option
		getIssuesOpts  WalkIssuesOptions
		mockHandler    func(w http.ResponseWriter, r *http.Request)
		wantErr        bool
		wantIssueCount int
	}{
		{
			name: "success with single page",
			options: []Option{
				WithToken("test-token"),
				WithRepository("owner", "repo"),
			},
			getIssuesOpts: WalkIssuesOptions{},
			mockHandler: func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "/repos/owner/repo/issues", r.URL.Path)
				assert.Equal(t, "GET", r.Method)
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`[
					{
						"id": 1,
						"number": 1,
						"title": "Test Issue 1",
						"state": "open"
					},
					{
						"id": 2,
						"number": 2,
						"title": "Test Issue 2",
						"state": "closed"
					}
				]`))
			},
			wantErr:        false,
			wantIssueCount: 2,
		},
		{
			name: "success with multiple pages",
			options: []Option{
				WithToken("test-token"),
				WithRepository("owner", "repo"),
			},
			getIssuesOpts: WalkIssuesOptions{},
			mockHandler: func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "/repos/owner/repo/issues", r.URL.Path)
				page := r.URL.Query().Get("page")
				w.Header().Set("Content-Type", "application/json")

				if page == "" || page == "1" {
					w.Header().Set("Link", `<https://api.github.com/repos/owner/repo/issues?page=2>; rel="next"`)
					w.WriteHeader(http.StatusOK)
					w.Write([]byte(`[
						{
							"id": 1,
							"number": 1,
							"title": "Test Issue 1",
							"state": "open"
						}
					]`))
				} else if page == "2" {
					w.WriteHeader(http.StatusOK)
					w.Write([]byte(`[
						{
							"id": 2,
							"number": 2,
							"title": "Test Issue 2",
							"state": "closed"
						}
					]`))
				}
			},
			wantErr:        false,
			wantIssueCount: 2,
		},
		{
			name: "ignore pull requests",
			options: []Option{
				WithToken("test-token"),
				WithRepository("owner", "repo"),
			},
			getIssuesOpts: WalkIssuesOptions{IgnorePullRequests: true},
			mockHandler: func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`[
					{
						"id": 1,
						"number": 1,
						"title": "Issue",
						"state": "open"
					},
					{
						"id": 2,
						"number": 2,
						"title": "PR",
						"state": "open",
						"pull_request": {
							"url": "https://api.github.com/repos/owner/repo/pulls/2"
						}
					}
				]`))
			},
			wantErr:        false,
			wantIssueCount: 1,
		},
		{
			name: "include pull requests when not ignored",
			options: []Option{
				WithToken("test-token"),
				WithRepository("owner", "repo"),
			},
			getIssuesOpts: WalkIssuesOptions{IgnorePullRequests: false},
			mockHandler: func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`[
					{
						"id": 1,
						"number": 1,
						"title": "Issue",
						"state": "open"
					},
					{
						"id": 2,
						"number": 2,
						"title": "PR",
						"state": "open",
						"pull_request": {
							"url": "https://api.github.com/repos/owner/repo/pulls/2"
						}
					}
				]`))
			},
			wantErr:        false,
			wantIssueCount: 2,
		},
		{
			name: "missing token",
			options: []Option{
				WithRepository("owner", "repo"),
			},
			getIssuesOpts: WalkIssuesOptions{},
			mockHandler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			},
			wantErr: true,
		},
		{
			name: "missing repository",
			options: []Option{
				WithToken("test-token"),
			},
			getIssuesOpts: WalkIssuesOptions{},
			mockHandler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			},
			wantErr: true,
		},
		{
			name: "API error",
			options: []Option{
				WithToken("test-token"),
				WithRepository("owner", "repo"),
			},
			getIssuesOpts: WalkIssuesOptions{},
			mockHandler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusUnauthorized)
				w.Write([]byte(`{"message": "Bad credentials"}`))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock server
			mux := http.NewServeMux()
			mux.HandleFunc("/repos/owner/repo/issues", tt.mockHandler)
			server := httptest.NewServer(mux)
			defer server.Close()

			// Create GitHub client with mock server
			httpClient := &http.Client{}
			ghClient := github.NewClient(httpClient)
			ghClient.BaseURL = must(ghClient.BaseURL.Parse(server.URL + "/"))

			// Add GitHub client option
			options := append(tt.options, WithGitHubClient(ghClient))
			c := NewConverter(options...).(*converterImpl)

			// Execute
			var gotIssues []*github.Issue
			err := c.WalkIssues(context.Background(), tt.getIssuesOpts, func(issues []*github.Issue) error {
				gotIssues = append(gotIssues, issues...)
				return nil
			})

			// Assert
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Len(t, gotIssues, tt.wantIssueCount)
			}
		})
	}
}

func Test_WalkIssues_WithNilContext(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/repos/owner/repo/issues", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`[{"id": 1, "number": 1, "title": "Test Issue"}]`))
	})
	server := httptest.NewServer(mux)
	defer server.Close()

	httpClient := &http.Client{}
	ghClient := github.NewClient(httpClient)
	ghClient.BaseURL = must(ghClient.BaseURL.Parse(server.URL + "/"))

	c := NewConverter(
		WithToken("test-token"),
		WithRepository("owner", "repo"),
		WithGitHubClient(ghClient),
	).(*converterImpl)

	var ctx context.Context
	err := c.WalkIssues(ctx, WalkIssuesOptions{}, func(issues []*github.Issue) error {
		return nil
	})
	require.NoError(t, err)
}

func Test_WalkIssues_WithNilCallback(t *testing.T) {
	requestCount := 0

	mux := http.NewServeMux()
	mux.HandleFunc("/repos/owner/repo/issues", func(w http.ResponseWriter, r *http.Request) {
		requestCount++
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`[{"id": 1, "number": 1, "title": "Test Issue"}]`))
	})
	server := httptest.NewServer(mux)
	defer server.Close()

	httpClient := &http.Client{}
	ghClient := github.NewClient(httpClient)
	ghClient.BaseURL = must(ghClient.BaseURL.Parse(server.URL + "/"))

	c := NewConverter(
		WithToken("test-token"),
		WithRepository("owner", "repo"),
		WithGitHubClient(ghClient),
	).(*converterImpl)

	err := c.WalkIssues(context.Background(), WalkIssuesOptions{}, nil)
	require.NoError(t, err)
	assert.Equal(t, 0, requestCount)
}

func Test_WalkIssues_WithPerPage(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/repos/owner/repo/issues", func(w http.ResponseWriter, r *http.Request) {
		perPage := r.URL.Query().Get("per_page")
		assert.Equal(t, "10", perPage)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`[{"id": 1, "number": 1, "title": "Test Issue"}]`))
	})
	server := httptest.NewServer(mux)
	defer server.Close()

	httpClient := &http.Client{}
	ghClient := github.NewClient(httpClient)
	ghClient.BaseURL = must(ghClient.BaseURL.Parse(server.URL + "/"))

	c := NewConverter(
		WithToken("test-token"),
		WithRepository("owner", "repo"),
		WithGitHubClient(ghClient),
	).(*converterImpl)

	got := make([]*github.Issue, 0)
	err := c.WalkIssues(context.Background(), WalkIssuesOptions{PerPage: 10}, func(issues []*github.Issue) error {
		got = append(got, issues...)
		return nil
	})
	require.NoError(t, err)
	assert.Len(t, got, 1)
}

func Test_WalkIssues_WithMultiplePages(t *testing.T) {
	requestedPages := map[string]int{}

	mux := http.NewServeMux()
	mux.HandleFunc("/repos/owner/repo/issues", func(w http.ResponseWriter, r *http.Request) {
		page := r.URL.Query().Get("page")
		if page == "" {
			page = "1"
		}

		requestedPages[page]++

		w.Header().Set("Content-Type", "application/json")
		switch page {
		case "1":
			w.Header().Set("Link", `<https://api.github.com/repos/owner/repo/issues?page=2>; rel="next"`)
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`[{"id": 1, "number": 1, "title": "Issue 1"}]`))
		case "2":
			w.Header().Set("Link", `<https://api.github.com/repos/owner/repo/issues?page=3>; rel="next", <https://api.github.com/repos/owner/repo/issues?page=3>; rel="last"`)
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`[{"id": 2, "number": 2, "title": "Issue 2"}]`))
		case "3":
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`[{"id": 3, "number": 3, "title": "Issue 3"}]`))
		default:
			t.Fatalf("unexpected page: %s", page)
		}
	})
	server := httptest.NewServer(mux)
	defer server.Close()

	httpClient := &http.Client{}
	ghClient := github.NewClient(httpClient)
	ghClient.BaseURL = must(ghClient.BaseURL.Parse(server.URL + "/"))

	c := NewConverter(
		WithToken("test-token"),
		WithRepository("owner", "repo"),
		WithGitHubClient(ghClient),
	).(*converterImpl)

	gotNumbers := make([]int, 0)
	err := c.WalkIssues(context.Background(), WalkIssuesOptions{}, func(issues []*github.Issue) error {
		for _, issue := range issues {
			gotNumbers = append(gotNumbers, issue.GetNumber())
		}
		return nil
	})
	require.NoError(t, err)
	assert.Equal(t, []int{1, 2, 3}, gotNumbers)
	assert.Equal(t, 1, requestedPages["1"])
	assert.Equal(t, 1, requestedPages["2"])
	assert.Equal(t, 1, requestedPages["3"])
}

func Test_getIssuesByPage(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/repos/owner/repo/issues", func(w http.ResponseWriter, r *http.Request) {
		page := r.URL.Query().Get("page")
		perPage := r.URL.Query().Get("per_page")
		state := r.URL.Query().Get("state")

		assert.Equal(t, "1", page)
		assert.Equal(t, "200", perPage)
		assert.Equal(t, "all", state)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`[{"id": 1, "number": 1, "title": "Test Issue"}]`))
	})
	server := httptest.NewServer(mux)
	defer server.Close()

	httpClient := &http.Client{}
	ghClient := github.NewClient(httpClient)
	ghClient.BaseURL = must(ghClient.BaseURL.Parse(server.URL + "/"))

	c := NewConverter(
		WithToken("test-token"),
		WithRepository("owner", "repo"),
		WithGitHubClient(ghClient),
	).(*converterImpl)

	issues, _, err := c.getIssuesByPage(context.Background(), 1, 200)
	require.NoError(t, err)
	assert.Len(t, issues, 1)
}

// Helper function for tests
func must[T any](v T, err error) T {
	if err != nil {
		panic(err)
	}
	return v
}
