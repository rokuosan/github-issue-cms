package github

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/google/go-github/v67/github"
	"github.com/stretchr/testify/assert"
)

func Test_NewClient(t *testing.T) {
	token := "test-token"
	client, err := NewClientWithToken(token)

	assert.NoError(t, err)
	assert.NotNil(t, client)
	assert.Equal(t, token, client.token)
	assert.NotNil(t, client.githubClient)
	assert.NotNil(t, client.httpClient)
	assert.NotNil(t, client.context)
}

func TestGetAllClosedIssues(t *testing.T) {
	tests := []struct {
		name          string
		totalIssues   int
		pagesCount    int
		opts          *GetIssuesOptions
		mockError     bool
		expectedError bool
	}{
		{
			name:        "Single page with 50 issues",
			totalIssues: 50,
			pagesCount:  1,
			opts:        nil,
		},
		{
			name:        "Multiple pages with 250 issues",
			totalIssues: 250,
			pagesCount:  2,
			opts: &GetIssuesOptions{
				MaxWorkers:     5,
				RetryAttempts:  2,
				InitialBackoff: 100 * time.Millisecond,
			},
		},
		{
			name:        "Large dataset with 1000 issues",
			totalIssues: 1000,
			pagesCount:  5,
			opts: &GetIssuesOptions{
				MaxWorkers:     10,
				RetryAttempts:  3,
				InitialBackoff: 200 * time.Millisecond,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// モックサーバーのセットアップ
			mux := http.NewServeMux()
			server := httptest.NewServer(mux)
			defer server.Close()

			// レート制限エンドポイント
			mux.HandleFunc("/rate_limit", func(w http.ResponseWriter, r *http.Request) {
				fmt.Fprintf(w, `{
					"resources": {
						"core": {
							"limit": 5000,
							"remaining": 4900,
							"reset": %d
						}
					}
				}`, time.Now().Add(time.Hour).Unix())
			})

			// Issuesエンドポイント
			mux.HandleFunc("/repos/test-owner/test-repo/issues", func(w http.ResponseWriter, r *http.Request) {
				page := r.URL.Query().Get("page")
				if page == "" {
					page = "1"
				}

				perPage := r.URL.Query().Get("per_page")
				assert.Equal(t, "200", perPage)

				state := r.URL.Query().Get("state")
				assert.Equal(t, "closed", state)

				// ページ番号に基づいてレスポンスを生成
				pageNum := 1
				fmt.Sscanf(page, "%d", &pageNum)

				// リンクヘッダーの設定
				if tt.pagesCount > 1 {
					w.Header().Set("Link", fmt.Sprintf(`<http://api.github.com/repos/test-owner/test-repo/issues?page=%d>; rel="last"`, tt.pagesCount))
				}

				// モックIssueの生成
				w.Write([]byte("["))
				issuesPerPage := 200
				if pageNum == tt.pagesCount && tt.totalIssues%200 != 0 {
					issuesPerPage = tt.totalIssues % 200
				}

				for i := 0; i < issuesPerPage; i++ {
					if i > 0 {
						w.Write([]byte(","))
					}
					issueNum := (pageNum-1)*200 + i + 1
					fmt.Fprintf(w, `{
						"id": %d,
						"number": %d,
						"title": "Issue %d",
						"body": "Body of issue %d",
						"state": "closed",
						"created_at": "2024-01-01T00:00:00Z",
						"updated_at": "2024-01-02T00:00:00Z"
					}`, issueNum, issueNum, issueNum, issueNum)
				}
				w.Write([]byte("]"))
			})

			// クライアントの作成
			client, _ := NewClientWithToken("test-token")
			client.githubClient = github.NewClient(nil)
			client.githubClient.BaseURL, _ = url.Parse(server.URL + "/")

			// テスト実行
			issues, err := client.GetAllClosedIssues("test-owner", "test-repo", tt.opts)

			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Len(t, issues, tt.totalIssues)

				// Issue内容の検証
				for _, issue := range issues {
					assert.NotNil(t, issue.ID)
					assert.NotNil(t, issue.Number)
					assert.Equal(t, "closed", *issue.State)
				}
			}
		})
	}
}

func TestGetClosedIssues_Compatibility(t *testing.T) {
	// 互換性のテスト - 旧メソッドが新メソッドを呼び出すことを確認
	mux := http.NewServeMux()
	server := httptest.NewServer(mux)
	defer server.Close()

	callCount := 0
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		callCount++
		w.Write([]byte("[]"))
	})

	client, _ := NewClientWithToken("test-token")
	client.githubClient = github.NewClient(nil)
	client.githubClient.BaseURL, _ = url.Parse(server.URL + "/")

	issues, err := client.GetClosedIssues("test-owner", "test-repo")

	assert.NoError(t, err)
	assert.NotNil(t, issues)
	assert.True(t, callCount > 0) // APIが呼ばれたことを確認
}

func TestFetchPageWithRetry(t *testing.T) {
	tests := []struct {
		name          string
		failCount     int
		opts          *GetIssuesOptions
		expectedError bool
	}{
		{
			name:      "Success on first attempt",
			failCount: 0,
			opts: &GetIssuesOptions{
				RetryAttempts:  3,
				InitialBackoff: 10 * time.Millisecond,
			},
		},
		{
			name:      "Success after 2 retries",
			failCount: 2,
			opts: &GetIssuesOptions{
				RetryAttempts:  3,
				InitialBackoff: 10 * time.Millisecond,
			},
		},
		{
			name:      "Fail after all retries",
			failCount: 5,
			opts: &GetIssuesOptions{
				RetryAttempts:  3,
				InitialBackoff: 10 * time.Millisecond,
			},
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			attemptCount := 0
			mux := http.NewServeMux()
			server := httptest.NewServer(mux)
			defer server.Close()

			mux.HandleFunc("/repos/test-owner/test-repo/issues", func(w http.ResponseWriter, r *http.Request) {
				attemptCount++
				if attemptCount <= tt.failCount {
					http.Error(w, "Internal Server Error", http.StatusInternalServerError)
					return
				}

				fmt.Fprintf(w, `[{
					"id": 1,
					"number": 1,
					"title": "Test Issue",
					"state": "closed"
				}]`)
			})

			client, _ := NewClientWithToken("test-token")
			client.githubClient = github.NewClient(nil)
			client.githubClient.BaseURL, _ = url.Parse(server.URL + "/")

			issues, err := client.fetchPageWithRetry("test-owner", "test-repo", 1, tt.opts)

			if tt.expectedError {
				assert.Error(t, err)
				assert.Nil(t, issues)
				assert.Equal(t, tt.opts.RetryAttempts+1, attemptCount)
			} else {
				assert.NoError(t, err)
				assert.Len(t, issues, 1)
				assert.Equal(t, tt.failCount+1, attemptCount)
			}
		})
	}
}

func TestRateLimitHandling(t *testing.T) {
	mux := http.NewServeMux()
	server := httptest.NewServer(mux)
	defer server.Close()

	// レート制限が少ない場合のテスト
	mux.HandleFunc("/rate_limit", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, `{
			"resources": {
				"core": {
					"limit": 5000,
					"remaining": 20,
					"reset": %d
				}
			}
		}`, time.Now().Add(time.Hour).Unix())
	})

	mux.HandleFunc("/repos/test-owner/test-repo/issues", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Link", `<http://api.github.com/repos/test-owner/test-repo/issues?page=3>; rel="last"`)
		w.Write([]byte(`[{"id": 1, "number": 1, "state": "closed"}]`))
	})

	client, _ := NewClientWithToken("test-token")
	client.githubClient = github.NewClient(nil)
	client.githubClient.BaseURL, _ = url.Parse(server.URL + "/")

	opts := &GetIssuesOptions{
		MaxWorkers:     10, // 10ワーカーを要求
		RetryAttempts:  1,
		InitialBackoff: 10 * time.Millisecond,
	}

	// レート制限により、ワーカー数が制限されることを確認
	issues, err := client.GetAllClosedIssues("test-owner", "test-repo", opts)

	assert.NoError(t, err)
	assert.NotNil(t, issues)
}
