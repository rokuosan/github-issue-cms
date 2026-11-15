package core

import (
	"context"
	"errors"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/google/go-github/v86/github"
	"github.com/rokuosan/github-issue-cms/pkg/config"
	"github.com/stretchr/testify/require"
)

func TestContentSnapshotQuery(t *testing.T) {
	s, err := NewContentSnapshot([]ContentArticle{
		{ID: 1, Title: "Go", Content: "Cloudflare Workers", Date: "2026-01-01", Tags: []string{"Go", "Web"}, Category: "tech"},
		{ID: 2, Title: "Other", Content: "CLOUDFLARE", Date: "2026-01-02", Tags: []string{"Web"}, Category: "tech"},
		{ID: 3, Title: "Go", Content: "Cloudflare", Date: "2026-01-03", Tags: []string{"Go", "Web"}, Category: "travel"},
	}, 1)
	require.NoError(t, err)
	page, err := s.Query(ContentQuery{Q: "cloudflare", Tags: []string{"Go", "Web"}, Category: "tech"})
	require.NoError(t, err)
	require.Equal(t, Pagination{Page: 1, PerPage: 1, TotalItems: 1, TotalPages: 1}, page.Pagination)
	require.Equal(t, 1, page.Items[0].ID)
	require.Equal(t, "/data/article/1.json", page.Items[0].URL)
	page, err = s.Query(ContentQuery{Page: 2, Sort: "date", Order: "asc"})
	require.NoError(t, err)
	require.Equal(t, 2, page.Items[0].ID)
	page, err = s.Query(ContentQuery{Page: int(^uint(0) >> 1)})
	require.NoError(t, err)
	require.Empty(t, page.Items)
	require.NotNil(t, page.Items)
	for _, q := range []ContentQuery{{Page: -1}, {PerPage: 101}, {PerPage: -1}, {Sort: "bad"}, {Order: "bad"}} {
		_, err := s.Query(q)
		require.Error(t, err)
	}
}

func TestContentSnapshotDatesAndCopies(t *testing.T) {
	input := []ContentArticle{{ID: 1, Date: "2026-01-01T09:00:00+09:00", Tags: []string{"a"}}, {ID: 2, Date: "2026-01-01T00:00:00Z"}}
	s, err := NewContentSnapshot(input, 20)
	require.NoError(t, err)
	require.Equal(t, 2, s.Articles[0].ID)
	require.Equal(t, "2026-01-01T00:00:00Z", s.Articles[1].Date)
	input[0].Tags[0] = "changed"
	a, ok := s.Find(1)
	require.True(t, ok)
	require.Equal(t, []string{"a"}, a.Tags)
	a.Tags[0] = "also changed"
	a, _ = s.Find(1)
	require.Equal(t, []string{"a"}, a.Tags)
	_, ok = s.Find(999)
	require.False(t, ok)
	_, err = NewContentSnapshot([]ContentArticle{{ID: 1, Date: "invalid"}}, 20)
	require.Error(t, err)
	_, err = NewContentSnapshot(input, 0)
	require.Error(t, err)
	_, err = NewContentSnapshot(append(input, input[0]), 20)
	require.Error(t, err)
}

type snapshotIssueStore struct {
	issues []*github.Issue
	calls  int
	query  IssueListQuery
}

func (s *snapshotIssueStore) ListIssues(_ context.Context, q IssueListQuery) ([]*github.Issue, error) {
	s.calls++
	s.query = q
	return s.issues, nil
}

type snapshotAssets struct {
	calls int
	err   error
}

func (s *snapshotAssets) Fetch(_ context.Context, _ *Image) (*ImageAsset, error) {
	s.calls++
	if s.err != nil {
		return nil, s.err
	}
	return &ImageAsset{Body: io.NopCloser(strings.NewReader("image")), ContentType: "image/png"}, nil
}

func TestLoadContentSnapshot(t *testing.T) {
	conf := *config.NewConfig()
	conf.GitHub.Username = "owner"
	conf.GitHub.Repository = "repo"
	conf.GitHub.Labels = []string{"publish"}
	created := github.Timestamp{Time: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)}
	store := &snapshotIssueStore{issues: []*github.Issue{
		{Number: github.Ptr(1), Title: github.Ptr("Original"), State: github.Ptr("open"), CreatedAt: &created, Body: github.Ptr("---\ntitle: Override\ndraft: false\ntags: [publish, Go]\ncustom: value\n---\n![image](https://github.com/user-attachments/assets/test)")},
		{Number: github.Ptr(2), State: github.Ptr("open"), CreatedAt: &created, Body: github.Ptr("Draft")},
		{Number: github.Ptr(3), State: github.Ptr("closed"), CreatedAt: &created, PullRequestLinks: &github.PullRequestLinks{}},
	}}
	assets := &snapshotAssets{}
	s, err := LoadContentSnapshot(context.Background(), store, assets, conf, 20, false)
	require.NoError(t, err)
	require.Equal(t, 1, store.calls)
	require.Equal(t, conf.GitHub.Labels, store.query.Labels)
	require.Len(t, s.Articles, 1)
	require.Equal(t, "Override", s.Articles[0].Title)
	require.Equal(t, []string{"Go"}, s.Articles[0].Tags)
	require.Equal(t, map[string]any{"custom": "value"}, s.Articles[0].Metadata)
	require.Contains(t, s.Articles[0].Content, "/images/1/0.png")
	require.Equal(t, []byte("image"), s.Assets["/images/1/0.png"].Data)
	require.Equal(t, 1, assets.calls)
	_, err = s.Query(ContentQuery{})
	require.NoError(t, err)
	require.Equal(t, 1, store.calls)
	s, err = LoadContentSnapshot(context.Background(), store, assets, conf, 20, true)
	require.NoError(t, err)
	require.Len(t, s.Articles, 2)
	assets.err = errors.New("offline")
	_, err = LoadContentSnapshot(context.Background(), store, assets, conf, 20, false)
	require.ErrorContains(t, err, "offline")
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	before := store.calls
	_, err = LoadContentSnapshot(ctx, store, assets, conf, 20, false)
	require.ErrorIs(t, err, context.Canceled)
	require.Equal(t, before, store.calls)
}

func TestContentSnapshotRejectsNonJSONMetadata(t *testing.T) {
	conf := config.Config{GitHub: &config.GitHubConfig{Username: "owner", Repository: "repo"}}
	created := github.Timestamp{Time: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)}
	store := &snapshotIssueStore{issues: []*github.Issue{{
		Number: github.Ptr(42), State: github.Ptr("closed"), CreatedAt: &created,
		Body: github.Ptr("---\ncustom:\n  1: one\n---\nBody"),
	}}}
	snapshot, err := LoadContentSnapshot(context.Background(), store, nil, conf, 20, false)
	require.ErrorContains(t, err, "article 42 cannot be encoded as JSON")
	require.Nil(t, snapshot)
}

func TestLoadContentSnapshotWithMinimalConfig(t *testing.T) {
	conf := config.Config{GitHub: &config.GitHubConfig{Username: "owner", Repository: "repo"}}
	created := github.Timestamp{Time: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)}
	store := &snapshotIssueStore{issues: []*github.Issue{{
		Number: github.Ptr(1), State: github.Ptr("closed"), CreatedAt: &created, Body: github.Ptr("Body"),
	}}}
	snapshot, err := LoadContentSnapshot(context.Background(), store, nil, conf, 20, false)
	require.NoError(t, err)
	require.Len(t, snapshot.Articles, 1)
	require.Nil(t, conf.Output)
}
