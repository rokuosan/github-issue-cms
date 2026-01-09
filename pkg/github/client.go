package github

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/google/go-github/v67/github"
)

type API interface{}

type client struct {
	context context.Context

	httpClient   *http.Client
	githubClient *github.Client

	token string
}

type GetIssuesOptions struct {
	MaxWorkers     int
	RetryAttempts  int
	InitialBackoff time.Duration
}

type pageJob struct {
	page int
}

type pageResult struct {
	issues []*github.Issue
	err    error
	page   int
}

func NewClientWithToken(token string) (*client, error) {
	httpClient := new(http.Client)
	githubClient := github.NewClient(httpClient).WithAuthToken(token)

	return &client{
		context:      context.Background(),
		token:        token,
		httpClient:   httpClient,
		githubClient: githubClient,
	}, nil
}

func (c *client) GetClosedIssues(owner, repository string) ([]*github.Issue, error) {
	return c.GetAllClosedIssues(owner, repository, nil)
}

func (c *client) GetAllClosedIssues(owner, repo string, opts *GetIssuesOptions) ([]*github.Issue, error) {
	if opts == nil {
		opts = &GetIssuesOptions{
			MaxWorkers:     10,
			RetryAttempts:  3,
			InitialBackoff: time.Second,
		}
	}

	// デフォルト値の設定
	if opts.MaxWorkers <= 0 {
		opts.MaxWorkers = 10
	}
	if opts.RetryAttempts <= 0 {
		opts.RetryAttempts = 3
	}
	if opts.InitialBackoff <= 0 {
		opts.InitialBackoff = time.Second
	}

	// 最初のページを取得して総ページ数を計算
	listOpts := &github.IssueListByRepoOptions{
		State: "closed",
		ListOptions: github.ListOptions{
			Page:    1,
			PerPage: 200,
		},
	}

	firstPage, resp, err := c.githubClient.Issues.ListByRepo(c.context, owner, repo, listOpts)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch first page: %w", err)
	}

	if resp.LastPage == 0 {
		// 1ページのみの場合
		return firstPage, nil
	}

	totalPages := resp.LastPage
	allIssues := make([]*github.Issue, 0, len(firstPage)*totalPages)
	allIssues = append(allIssues, firstPage...)

	// レート制限を確認して並列度を決定
	rateLimit, _, err := c.githubClient.RateLimit.Get(c.context)
	if err != nil {
		// レート制限の取得に失敗した場合はデフォルトで進める
		fmt.Printf("Warning: failed to get rate limits: %v\n", err)
	}

	workers := opts.MaxWorkers
	if rateLimit != nil && rateLimit.Core.Remaining < workers*5 {
		// レート制限が少ない場合は並列度を下げる
		workers = rateLimit.Core.Remaining / 5
		if workers < 1 {
			workers = 1
		}
	}

	// チャネルの作成
	jobCh := make(chan pageJob, totalPages)
	resultCh := make(chan pageResult, totalPages)

	// Worker Poolの起動
	var wg sync.WaitGroup
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go c.worker(owner, repo, opts, jobCh, resultCh, &wg)
	}

	// ジョブの投入（2ページ目から）
	for page := 2; page <= totalPages; page++ {
		jobCh <- pageJob{page: page}
	}
	close(jobCh)

	// 結果の収集
	go func() {
		wg.Wait()
		close(resultCh)
	}()

	var fetchErrors []error
	for result := range resultCh {
		if result.err != nil {
			fetchErrors = append(fetchErrors, fmt.Errorf("failed to fetch page %d: %w", result.page, result.err))
			continue
		}
		allIssues = append(allIssues, result.issues...)
	}

	// 部分的な失敗は許容するが、エラーがあった場合は警告する
	if len(fetchErrors) > 0 {
		fmt.Printf("Warning: %d pages failed to fetch\n", len(fetchErrors))
		for _, err := range fetchErrors {
			fmt.Printf("  - %v\n", err)
		}
	}

	return allIssues, nil
}

func (c *client) worker(owner, repo string, opts *GetIssuesOptions, jobs <-chan pageJob, results chan<- pageResult, wg *sync.WaitGroup) {
	defer wg.Done()

	for job := range jobs {
		issues, err := c.fetchPageWithRetry(owner, repo, job.page, opts)
		results <- pageResult{
			issues: issues,
			err:    err,
			page:   job.page,
		}
	}
}

func (c *client) fetchPageWithRetry(owner, repo string, page int, opts *GetIssuesOptions) ([]*github.Issue, error) {
	listOpts := &github.IssueListByRepoOptions{
		State: "closed",
		ListOptions: github.ListOptions{
			Page:    page,
			PerPage: 200,
		},
	}

	var lastErr error
	backoff := opts.InitialBackoff

	for attempt := 0; attempt <= opts.RetryAttempts; attempt++ {
		if attempt > 0 {
			time.Sleep(backoff)
			backoff *= 2 // Exponential backoff
		}

		issues, _, err := c.githubClient.Issues.ListByRepo(c.context, owner, repo, listOpts)
		if err == nil {
			return issues, nil
		}

		lastErr = err
		// レート制限エラーの場合は追加の待機時間を設ける
		if _, ok := err.(*github.RateLimitError); ok {
			time.Sleep(time.Minute)
		}
	}

	return nil, fmt.Errorf("all retry attempts failed: %w", lastErr)
}
