package core

import (
	"testing"
	"time"

	"github.com/google/go-github/v85/github"
	"github.com/rokuosan/github-issue-cms/pkg/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewArticleService(t *testing.T) {
	conf := *config.NewConfig()
	service := NewArticleService(conf)

	assert.NotNil(t, service)
	assertEqualCmp(t, conf, service.config)
}

func TestArticleService_ConvertIssueToArticle(t *testing.T) {
	conf := *config.NewConfig()
	conf.Output.Images.URL = "/images/%Y-%m-%d"
	conf.Output.Images.Filename = "[:id].png"

	service := NewArticleService(conf)

	tests := []struct {
		name  string
		issue *github.Issue
		want  map[string]interface{}
	}{
		{
			name: "基本的なIssue変換",
			issue: &github.Issue{
				Title:     stringPtr("Test Issue"),
				Body:      stringPtr("Test body content"),
				CreatedAt: parseTime("2021-01-01T12:00:00Z"),
				User:      &github.User{Login: stringPtr("testuser")},
				State:     stringPtr("closed"),
				Labels:    []*github.Label{},
			},
			want: map[string]interface{}{
				"title":   "Test Issue",
				"author":  "testuser",
				"content": "Test body content\n",
				"date":    "2021-01-01T12:00:00Z",
				"draft":   false,
				"key":     "2021-01-01_120000",
			},
		},
		{
			name: "open状態のIssue（下書き）",
			issue: &github.Issue{
				Title:     stringPtr("Draft Issue"),
				Body:      stringPtr("Draft content"),
				CreatedAt: parseTime("2021-06-15T10:30:00Z"),
				User:      &github.User{Login: stringPtr("author")},
				State:     stringPtr("open"),
				Labels:    []*github.Label{},
			},
			want: map[string]interface{}{
				"title":  "Draft Issue",
				"author": "author",
				"draft":  true,
			},
		},
		{
			name: "タグ付きIssue",
			issue: &github.Issue{
				Title:     stringPtr("Tagged Issue"),
				Body:      stringPtr("Content"),
				CreatedAt: parseTime("2021-03-10T00:00:00Z"),
				User:      &github.User{Login: stringPtr("user")},
				State:     stringPtr("closed"),
				Labels: []*github.Label{
					{Name: stringPtr("bug")},
					{Name: stringPtr("enhancement")},
				},
			},
			want: map[string]interface{}{
				"tags": []string{"bug", "enhancement"},
			},
		},
		{
			name: "マイルストーン付きIssue",
			issue: &github.Issue{
				Title:     stringPtr("Milestone Issue"),
				Body:      stringPtr("Content"),
				CreatedAt: parseTime("2021-05-01T00:00:00Z"),
				User:      &github.User{Login: stringPtr("user")},
				State:     stringPtr("closed"),
				Labels:    []*github.Label{},
				Milestone: &github.Milestone{Title: stringPtr("v1.0")},
			},
			want: map[string]interface{}{
				"category": "v1.0",
			},
		},
		{
			name: "フロントマター付きIssue",
			issue: &github.Issue{
				Title:     stringPtr("FrontMatter Issue"),
				Body:      stringPtr("```\nauthor: Custom Author\ncustom: value\n```\n\nContent here"),
				CreatedAt: parseTime("2021-02-01T00:00:00Z"),
				User:      &github.User{Login: stringPtr("user")},
				State:     stringPtr("closed"),
				Labels:    []*github.Label{},
			},
			want: map[string]interface{}{
				"content":     "Content here\n",
				"frontMatter": map[string]any{"author": "Custom Author", "custom": "value"},
			},
		},
		{
			name: "YAML front matter issue",
			issue: &github.Issue{
				Title:     stringPtr("YAML FrontMatter Issue"),
				Body:      stringPtr("---\nauthor: YAML Author\ncustom: value\n---\n\nContent here"),
				CreatedAt: parseTime("2021-02-01T00:00:00Z"),
				User:      &github.User{Login: stringPtr("user")},
				State:     stringPtr("closed"),
				Labels:    []*github.Label{},
			},
			want: map[string]interface{}{
				"content":     "Content here\n",
				"frontMatter": map[string]any{"author": "YAML Author", "custom": "value"},
			},
		},
		{
			name: "TOML front matter issue",
			issue: &github.Issue{
				Title:     stringPtr("TOML FrontMatter Issue"),
				Body:      stringPtr("+++\nauthor = \"TOML Author\"\ncustom = \"value\"\n+++\n\nContent here"),
				CreatedAt: parseTime("2021-02-01T00:00:00Z"),
				User:      &github.User{Login: stringPtr("user")},
				State:     stringPtr("closed"),
				Labels:    []*github.Label{},
			},
			want: map[string]interface{}{
				"content":     "Content here\n",
				"frontMatter": map[string]any{"author": "TOML Author", "custom": "value"},
			},
		},
		{
			name: "Markdown画像付きIssue",
			issue: &github.Issue{
				Title:     stringPtr("Image Issue"),
				Body:      stringPtr("![image](https://example.com/image1.png)\n\n![image](https://example.com/image2.png)"),
				CreatedAt: parseTime("2021-04-01T15:00:00Z"),
				User:      &github.User{Login: stringPtr("user")},
				State:     stringPtr("closed"),
				Labels:    []*github.Label{},
			},
			want: map[string]interface{}{
				"imageCount": 2,
			},
		},
		{
			name: "HTML画像付きIssue",
			issue: &github.Issue{
				Title:     stringPtr("HTML Image Issue"),
				Body:      stringPtr(`<img width="100" alt="test" src="https://example.com/test.png">`),
				CreatedAt: parseTime("2021-07-01T00:00:00Z"),
				User:      &github.User{Login: stringPtr("user")},
				State:     stringPtr("closed"),
				Labels:    []*github.Label{},
			},
			want: map[string]interface{}{
				"imageCount": 1,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := service.ConvertIssueToArticle(tt.issue)

			assert.NotNil(t, got)

			if title, ok := tt.want["title"]; ok {
				assertEqualCmp(t, title, got.Title)
			}
			if author, ok := tt.want["author"]; ok {
				assertEqualCmp(t, author, got.Author)
			}
			if content, ok := tt.want["content"]; ok {
				assertEqualCmp(t, content, got.Content)
			}
			if date, ok := tt.want["date"]; ok {
				assertEqualCmp(t, date, got.Date)
			}
			if draft, ok := tt.want["draft"]; ok {
				assertEqualCmp(t, draft, got.Draft)
			}
			if key, ok := tt.want["key"]; ok {
				assertEqualCmp(t, key, got.Key)
			}
			if tags, ok := tt.want["tags"]; ok {
				assertEqualCmp(t, tags, got.Tags)
			}
			if category, ok := tt.want["category"]; ok {
				assertEqualCmp(t, category, got.Category)
			}
			if frontMatter, ok := tt.want["frontMatter"]; ok {
				assertEqualCmp(t, frontMatter, got.FrontMatter.Values())
			}
			if imageCount, ok := tt.want["imageCount"]; ok {
				assertEqualCmp(t, imageCount, len(got.Images))
			}
		})
	}
}

func TestArticleService_ConvertIssueToArticle_PullRequest(t *testing.T) {
	service := NewArticleService(*config.NewConfig())

	pr := &github.Issue{
		Title:            stringPtr("PR Title"),
		Body:             stringPtr("PR body"),
		CreatedAt:        parseTime("2021-01-01T00:00:00Z"),
		User:             &github.User{Login: stringPtr("user")},
		State:            stringPtr("open"),
		Labels:           []*github.Label{},
		PullRequestLinks: &github.PullRequestLinks{},
	}

	got := service.ConvertIssueToArticle(pr)
	assert.Nil(t, got)
}

func TestArticleService_ConvertIssueToArticle_CRRemoval(t *testing.T) {
	service := NewArticleService(*config.NewConfig())

	issue := &github.Issue{
		Title:     stringPtr("Test"),
		Body:      stringPtr("Line1\r\nLine2\r\nLine3"),
		CreatedAt: parseTime("2021-01-01T00:00:00Z"),
		User:      &github.User{Login: stringPtr("user")},
		State:     stringPtr("closed"),
		Labels:    []*github.Label{},
	}

	got := service.ConvertIssueToArticle(issue)
	assert.NotNil(t, got)
	assertEqualCmp(t, "Line1\nLine2\nLine3\n", got.Content)
}

func TestMetadataParser_Parse(t *testing.T) {
	parser := newMetadataParser()

	tests := []struct {
		name       string
		body       string
		wantRaw    string
		wantParsed map[string]any
		wantErr    string
	}{
		{
			name:       "yaml code fence",
			body:       "```yaml\nauthor: Test User\ncustom: value\n```\n\nBody",
			wantRaw:    "```yaml\nauthor: Test User\ncustom: value\n```",
			wantParsed: map[string]any{"author": "Test User", "custom": "value"},
		},
		{
			name:       "empty yaml code fence",
			body:       "```yaml\n```\n\nBody",
			wantRaw:    "```yaml\n```",
			wantParsed: map[string]any{},
		},
		{
			name:       "yaml front matter",
			body:       "---\nauthor: Test User\ncustom: value\n---\n\nBody",
			wantRaw:    "---\nauthor: Test User\ncustom: value\n---",
			wantParsed: map[string]any{"author": "Test User", "custom": "value"},
		},
		{
			name:       "toml front matter",
			body:       "+++\nauthor = \"Test User\"\ncustom = \"value\"\n+++\n\nBody",
			wantRaw:    "+++\nauthor = \"Test User\"\ncustom = \"value\"\n+++",
			wantParsed: map[string]any{"author": "Test User", "custom": "value"},
		},
		{
			name:    "invalid yaml code fence",
			body:    "```\nauthor: [broken\n```\n\nBody",
			wantErr: "failed to parse front matter",
		},
		{
			name:    "no metadata",
			body:    "Body only",
			wantErr: "front matter not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			block, err := parser.Parse(tt.body)
			if tt.wantErr != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantErr)
				return
			}

			require.NoError(t, err)
			assertEqualCmp(t, tt.wantRaw, block.Raw)
			assertEqualCmp(t, tt.wantParsed, block.Values.Values())
		})
	}
}

func stringPtr(s string) *string {
	return &s
}

func parseTime(s string) *github.Timestamp {
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		panic(err)
	}
	return &github.Timestamp{Time: t}
}
