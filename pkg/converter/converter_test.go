package converter

import (
	"regexp"
	"testing"
	"time"

	"github.com/google/go-github/v67/github"
	"github.com/rokuosan/github-issue-cms/pkg/config"
	"github.com/stretchr/testify/assert"
)

func TestConverterInterface(t *testing.T) {
	var c = NewConverter(Config{Config: *config.NewConfig()}, "test-token")
	assert.NotNil(t, c)
}

func TestConverter_IssueToArticle(t *testing.T) {
	tests := []struct {
		name         string
		issue        *github.Issue
		expected     *Article
		shouldBeNil  bool
		config       config.Config
	}{
		{
			name: "valid",
			issue: &github.Issue{
				Title:     stringPtr("Test"),
				Body:      stringPtr("This is a test"),
				CreatedAt: parseTime("2021-01-01T00:00:00Z"),
				Labels:    []*github.Label{},
				User: &github.User{
					Login: stringPtr("allowed_user1"),
				},
			},
			expected: &Article{
				Author: "allowed_user1",
				Title:   "Test",
				Content: "This is a test\n",
				Date:    "2021-01-01T00:00:00Z",
				Key:     "2021-01-01_000000",
			},
			config: config.Config{
				GitHub: &config.GitHubConfig{
					AllowedAuthors: []string{"allowed_user1", "allowed_user2"},
				},
				Hugo: config.NewHugoConfig(),
			},
			shouldBeNil: false,
		},
		{
			name: "allowed author",
			issue: &github.Issue{
				Title:     stringPtr("Test"),
				Body:      stringPtr("This is a test"),
				CreatedAt: parseTime("2021-01-01T00:00:00Z"),
				User: &github.User{
					Login: stringPtr("allowed_user1"),
				},
			},
			expected: &Article{
				Author: "allowed_user1",
				Title:   "Test",
				Content: "This is a test\n",
				Date:    "2021-01-01T00:00:00Z",
				Key:     "2021-01-01_000000",
			},
			config: config.Config{
				GitHub: &config.GitHubConfig{
					AllowedAuthors: []string{"allowed_user1", "allowed_user2"},
				},
				Hugo: config.NewHugoConfig(),
			},
			shouldBeNil: false,
		},
		{
			name: "not allowed author",
			issue: &github.Issue{
				Title:     stringPtr("Test"),
				Body:      stringPtr("This is a test"),
				CreatedAt: parseTime("2021-01-01T00:00:00Z"),
				User: &github.User{
					Login: stringPtr("not_allowed_user"),
				},
			},
			expected:    nil,
			config: config.Config{
				GitHub: &config.GitHubConfig{
					AllowedAuthors: []string{"allowed_user1", "allowed_user2"},
				},
				Hugo: config.NewHugoConfig(),
			},
			shouldBeNil: true,
		},
		{
			name: "no allowed authors set",
			issue: &github.Issue{
				Title:     stringPtr("Test"),
				Body:      stringPtr("This is a test"),
				CreatedAt: parseTime("2021-01-01T00:00:00Z"),
				User: &github.User{
					Login: stringPtr("any_user"),
				},
			},
			expected: &Article{
				Author: "any_user",
				Title:   "Test",
				Content: "This is a test\n",
				Date:    "2021-01-01T00:00:00Z",
				Key:     "2021-01-01_000000",
			},
			config: config.Config{
				GitHub: &config.GitHubConfig{
					AllowedAuthors: []string{},
				},
				Hugo: config.NewHugoConfig(),
			},
			shouldBeNil: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &converterImpl{config: Config{Config: tt.config}}
			result := c.IssueToArticle(tt.issue)

			if tt.shouldBeNil {
				assert.Nil(t, result)
				println("Expected nil result, got:", result)
			} else {
				assert.Equal(t, tt.expected, result)
				println("Expected result:", tt.expected, "got:", result)
			}
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
	return &github.Timestamp{
		Time: t,
	}
}

func TestConverter_removeCR(t *testing.T) {
	c := &converterImpl{}

	tests := []struct {
		name     string
		content  string
		expected string
	}{
		{
			name:     "valid 1",
			content:  "This is\r\r\r a test\r\r\r\r\r",
			expected: "This is a test",
		},
		{
			name:     "valid 2",
			content:  "This is a test\r\n\r\n",
			expected: "This is a test\n\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := c.removeCR(tt.content)

			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestConverter_insertTrailingNewline(t *testing.T) {
	c := &converterImpl{}

	tests := []struct {
		name     string
		content  string
		expected string
	}{
		{
			name:     "insert newline if not exists",
			content:  "This is a test",
			expected: "This is a test\n",
		},
		{
			name:     "do not insert newline if exists",
			content:  "This is a test\n",
			expected: "This is a test\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := c.insertTrailingNewline(tt.content)

			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestConverter_extractFrontMatter(t *testing.T) {
	c := &converterImpl{}

	tests := []struct {
		name     string
		body     string
		expected []string
		wantErr  bool
		errType  error
	}{
		{
			name: "valid front matter",
			body: "```\ntitle: Test\ndescription: This is a test\n```\nContent goes here",
			expected: []string{
				"```\ntitle: Test\ndescription: This is a test\n```",
				"title: Test\ndescription: This is a test",
			},
			wantErr: false,
		},
		{
			name:     "no front matter",
			body:     "Content without front matter",
			expected: nil,
			wantErr:  true,
			errType:  errFrontMatterNotFound(),
		},
		{
			name:     "invalid YAML in front matter",
			body:     "```\ntitle: Test\ndescription: This is a test\n  invalid: -\n```\nContent goes here",
			expected: nil,
			wantErr:  true,
		},
		{
			name: "front matter with extra spaces",
			body: "   ```\ntitle: Test\ndescription: This is a test\n```\nContent goes here",
			expected: []string{
				"   ```\ntitle: Test\ndescription: This is a test\n```",
				"title: Test\ndescription: This is a test",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := c.extractFrontMatter(tt.body)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errType != nil {
					assert.Equal(t, tt.errType.Error(), err.Error())
				}
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestConverter_createMarkdownImageExpression(t *testing.T) {
	c := &converterImpl{config: Config{Config: *config.NewConfig()}}

	tests := []struct {
		name     string
		images   string
		path     string
		alt      string
		id       int
		expected string
	}{
		{
			name:     "valid",
			images:   "[:id].png",
			path:     "images",
			alt:      "Test Image",
			id:       0,
			expected: "![Test Image](images/0.png)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c.config.Hugo.Filename.Images = tt.images
			result := c.createMarkdownImageExpression(tt.path, tt.alt, tt.id)

			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestConverter_replaceImageURL(t *testing.T) {
	c := &converterImpl{config: Config{Config: *config.NewConfig()}}

	tests := []struct {
		name     string
		re       *regexp.Regexp
		content  string
		baseURL  string
		time     string
		offset   int
		expected string
	}{
		{
			name: "valid",
			re:   regexMarkdownImage,
			content: `
				![image](https://example.com/image.png)
				![image](https://example.com/another.png)`,
			baseURL: "images",
			expected: `
				![https://example.com/image.png](images/0.png)
				![https://example.com/another.png](images/1.png)`,
		},
		{
			name:   "valid with offset",
			re:     regexMarkdownImage,
			offset: 10,
			content: `
				![image](https://example.com/image.png)
				![image](https://example.com/another.png)`,
			baseURL: "images",
			expected: `
				![https://example.com/image.png](images/10.png)
				![https://example.com/another.png](images/11.png)`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, _ := c.replaceImageURL(replaceImageURLInput{
				re:      tt.re,
				content: tt.content,
				baseURL: tt.baseURL,
				time:    tt.time,
				offset:  tt.offset,
			})

			assert.Equal(t, tt.expected, result)
		})
	}

}
