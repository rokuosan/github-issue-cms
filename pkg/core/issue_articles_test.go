package core

import (
	"strings"
	"testing"
	"time"

	"github.com/google/go-github/v86/github"
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
	conf.Output.Images.BaseURL = Ptr("/images/%Y-%m-%d")
	conf.Output.Images.Filename = "[:id].png"
	conf.GitHub.Labels = []string{"published"}

	service := NewArticleService(conf)

	tests := []struct {
		name  string
		issue *github.Issue
		want  map[string]interface{}
	}{
		{
			name: "基本的なIssue変換",
			issue: &github.Issue{
				Title:     Ptr("Test Issue"),
				Body:      Ptr("Test body content"),
				CreatedAt: parseTime("2021-01-01T12:00:00Z"),
				User:      &github.User{Login: Ptr("testuser")},
				State:     Ptr("closed"),
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
				Title:     Ptr("Draft Issue"),
				Body:      Ptr("Draft content"),
				CreatedAt: parseTime("2021-06-15T10:30:00Z"),
				User:      &github.User{Login: Ptr("author")},
				State:     Ptr("open"),
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
				Title:     Ptr("Tagged Issue"),
				Body:      Ptr("Content"),
				CreatedAt: parseTime("2021-03-10T00:00:00Z"),
				User:      &github.User{Login: Ptr("user")},
				State:     Ptr("closed"),
				Labels: []*github.Label{
					{Name: Ptr("bug")},
					{Name: Ptr("enhancement")},
				},
			},
			want: map[string]interface{}{
				"tags": []string{"bug", "enhancement"},
			},
		},
		{
			name: "フィルタ用ラベルはタグから除外",
			issue: &github.Issue{
				Title:     Ptr("Filtered Label Issue"),
				Body:      Ptr("Content"),
				CreatedAt: parseTime("2021-03-10T00:00:00Z"),
				User:      &github.User{Login: Ptr("user")},
				State:     Ptr("closed"),
				Labels: []*github.Label{
					{Name: Ptr("published")},
					{Name: Ptr("go")},
				},
			},
			want: map[string]interface{}{
				"tags": []string{"go"},
			},
		},
		{
			name: "マイルストーン付きIssue",
			issue: &github.Issue{
				Title:     Ptr("Milestone Issue"),
				Body:      Ptr("Content"),
				CreatedAt: parseTime("2021-05-01T00:00:00Z"),
				User:      &github.User{Login: Ptr("user")},
				State:     Ptr("closed"),
				Labels:    []*github.Label{},
				Milestone: &github.Milestone{Title: Ptr("v1.0")},
			},
			want: map[string]interface{}{
				"category": "v1.0",
			},
		},
		{
			name: "フロントマター付きIssue",
			issue: &github.Issue{
				Title:     Ptr("FrontMatter Issue"),
				Body:      Ptr("```\nauthor: Custom Author\ncustom: value\n```\n\nContent here"),
				CreatedAt: parseTime("2021-02-01T00:00:00Z"),
				User:      &github.User{Login: Ptr("user")},
				State:     Ptr("closed"),
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
				Title:     Ptr("YAML FrontMatter Issue"),
				Body:      Ptr("---\nauthor: YAML Author\ncustom: value\n---\n\nContent here"),
				CreatedAt: parseTime("2021-02-01T00:00:00Z"),
				User:      &github.User{Login: Ptr("user")},
				State:     Ptr("closed"),
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
				Title:     Ptr("TOML FrontMatter Issue"),
				Body:      Ptr("+++\nauthor = \"TOML Author\"\ncustom = \"value\"\n+++\n\nContent here"),
				CreatedAt: parseTime("2021-02-01T00:00:00Z"),
				User:      &github.User{Login: Ptr("user")},
				State:     Ptr("closed"),
				Labels:    []*github.Label{},
			},
			want: map[string]interface{}{
				"content":     "Content here\n",
				"frontMatter": map[string]any{"author": "TOML Author", "custom": "value"},
			},
		},
		{
			name: "設定済みプレフィックスに一致するURLだけを収集",
			issue: &github.Issue{
				Title: Ptr("GitHub Asset Issue"),
				Body: Ptr(
					"![image](https://github.com/user-attachments/assets/11111111-1111-1111-1111-111111111111)\n\n" +
						`<img src="https://private-user-images.githubusercontent.com/22222222/33333333-4444-5555-6666-777777777777.png?jwt=token" alt="test">` + "\n\n" +
						"[legacy](https://user-images.githubusercontent.com/1234567/abcdef01-2345-6789-abcd-ef0123456789.png)\n\n" +
						"![ignored](https://example.com/image.png)",
				),
				CreatedAt: parseTime("2021-04-01T15:00:00Z"),
				User:      &github.User{Login: Ptr("user")},
				State:     Ptr("closed"),
				Labels:    []*github.Label{},
			},
			want: map[string]interface{}{
				"imageCount": 3,
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

func TestExtractGitHubHostedImages(t *testing.T) {
	targetURLs := config.NewOutputImagesConfig().TargetURLs()

	tests := []struct {
		name     string
		content  string
		wantURLs []string
	}{
		{
			name: "deduplicates repeated URLs and keeps detection order",
			content: strings.Join([]string{
				"![a](https://github.com/user-attachments/assets/11111111-1111-1111-1111-111111111111)",
				`<img src="https://github.com/user-attachments/assets/11111111-1111-1111-1111-111111111111">`,
				"https://private-user-images.githubusercontent.com/22222222/33333333-4444-5555-6666-777777777777.png?jwt=token",
				"https://user-images.githubusercontent.com/1234567/abcdef01-2345-6789-abcd-ef0123456789.png",
			}, "\n"),
			wantURLs: []string{
				"https://github.com/user-attachments/assets/11111111-1111-1111-1111-111111111111",
				"https://private-user-images.githubusercontent.com/22222222/33333333-4444-5555-6666-777777777777.png?jwt=token",
				"https://user-images.githubusercontent.com/1234567/abcdef01-2345-6789-abcd-ef0123456789.png",
			},
		},
		{
			name: "trims trailing backticks and punctuation",
			content: strings.Join([]string{
				"`https://github.com/user-attachments/assets/11111111-1111-1111-1111-111111111111`",
				"`https://private-user-images.githubusercontent.com/22222222/33333333-4444-5555-6666-777777777777.png?jwt=token`,",
			}, "\n"),
			wantURLs: []string{
				"https://github.com/user-attachments/assets/11111111-1111-1111-1111-111111111111",
				"https://private-user-images.githubusercontent.com/22222222/33333333-4444-5555-6666-777777777777.png?jwt=token",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractTargetImages(tt.content, "2021-01-01_000000", targetURLs)

			require.Len(t, got, len(tt.wantURLs))
			for i, wantURL := range tt.wantURLs {
				assertEqualCmp(t, wantURL, got[i].URL)
				assertEqualCmp(t, i, got[i].ID)
			}
		})
	}
}

func TestExtractTargetImages_UsesConfiguredPrefixes(t *testing.T) {
	got := extractTargetImages(strings.Join([]string{
		"https://cdn.example.com/assets/alpha.png",
		"https://media.example.net/ignored.png",
		"https://cdn.example.com/assets/beta.png",
	}, "\n"), "2021-01-01_000000", []string{
		"https://cdn.example.com/assets/",
	})

	require.Len(t, got, 2)
	assertEqualCmp(t, "https://cdn.example.com/assets/alpha.png", got[0].URL)
	assertEqualCmp(t, "https://cdn.example.com/assets/beta.png", got[1].URL)
}

func TestExtractTargetImages_UsesWildcardHostPatterns(t *testing.T) {
	got := extractTargetImages(strings.Join([]string{
		"https://user-images.githubusercontent.com/123/a.png",
		"https://private-user-images.githubusercontent.com/456/b.png?jwt=token",
		"https://example.com/c.png",
	}, "\n"), "2021-01-01_000000", []string{
		"https://*.githubusercontent.com",
	})

	require.Len(t, got, 2)
	assertEqualCmp(t, "https://user-images.githubusercontent.com/123/a.png", got[0].URL)
	assertEqualCmp(t, "https://private-user-images.githubusercontent.com/456/b.png?jwt=token", got[1].URL)
}

func TestArticleService_ConvertIssueToArticle_PullRequest(t *testing.T) {
	service := NewArticleService(*config.NewConfig())

	pr := &github.Issue{
		Title:            Ptr("PR Title"),
		Body:             Ptr("PR body"),
		CreatedAt:        parseTime("2021-01-01T00:00:00Z"),
		User:             &github.User{Login: Ptr("user")},
		State:            Ptr("open"),
		Labels:           []*github.Label{},
		PullRequestLinks: &github.PullRequestLinks{},
	}

	got := service.ConvertIssueToArticle(pr)
	assert.Nil(t, got)
}

func TestArticleService_ConvertIssueToArticle_CRRemoval(t *testing.T) {
	service := NewArticleService(*config.NewConfig())

	issue := &github.Issue{
		Title:     Ptr("Test"),
		Body:      Ptr("Line1\r\nLine2\r\nLine3"),
		CreatedAt: parseTime("2021-01-01T00:00:00Z"),
		User:      &github.User{Login: Ptr("user")},
		State:     Ptr("closed"),
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

func parseTime(s string) *github.Timestamp {
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		panic(err)
	}
	return &github.Timestamp{Time: t}
}
