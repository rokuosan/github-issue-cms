package converter_v2

import (
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/google/go-github/v67/github"
	"gopkg.in/yaml.v3"
)

func body(t *testing.T, frontMatter, content string) string {
	t.Helper()
	var sb strings.Builder
	if frontMatter != "" {
		sb.WriteString("```\n")
		sb.WriteString(frontMatter)
		sb.WriteString("\n```\n")
	}
	sb.WriteString(content)
	return sb.String()
}

func TestIssueArticle_Title(t *testing.T) {
	tests := []struct {
		name        string
		issueTitle  string
		frontmatter string
		expected    string
	}{
		{
			name:        "Title from front matter",
			issueTitle:  "Issue Title",
			frontmatter: "title: Custom Title",
			expected:    "Custom Title",
		},
		{
			name:        "Title from issue when no front matter",
			issueTitle:  "Issue Title",
			frontmatter: "",
			expected:    "Issue Title",
		},
		{
			name:        "Title from issue when front matter has no title",
			issueTitle:  "Issue Title",
			frontmatter: "title: ",
			expected:    "Issue Title",
		},
		{
			name:        "Empty title when both front matter and issue title are empty",
			issueTitle:  "",
			frontmatter: "",
			expected:    "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			issue := &github.Issue{
				Title: github.String(tt.issueTitle),
				Body:  github.String(body(t, tt.frontmatter, "This is the body.")),
			}
			article, err := NewIssueArticle(Markdown, issue)
			if err != nil {
				t.Fatalf("Failed to create IssueArticle: %v", err)
			}

			if article.Title() != tt.expected {
				t.Errorf("Expected title %q, got %q", tt.expected, article.Title())
			}
		})
	}
}

func TestIssueArticle_Author(t *testing.T) {
	tests := []struct {
		name        string
		issueUser   string
		frontmatter string
		expected    string
	}{
		{
			name:        "Author from front matter",
			issueUser:   "issue_user",
			frontmatter: "author: Custom Author",
			expected:    "Custom Author",
		},
		{
			name:        "Author from issue when no front matter",
			issueUser:   "issue_user",
			frontmatter: "",
			expected:    "issue_user",
		},
		{
			name:        "Author from issue when front matter has no author",
			issueUser:   "issue_user",
			frontmatter: "author: ",
			expected:    "issue_user",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			issue := &github.Issue{
				User: &github.User{
					Login: github.String(tt.issueUser),
				},
				Body: github.String(body(t, tt.frontmatter, "This is the body.")),
			}
			article, err := NewIssueArticle(Markdown, issue)
			if err != nil {
				t.Fatalf("Failed to create IssueArticle: %v", err)
			}

			if article.Author() != tt.expected {
				t.Errorf("Expected author %q, got %q", tt.expected, article.Author())
			}
		})
	}
}

func TestIssueArticle_Category(t *testing.T) {
	tests := []struct {
		name        string
		milestone   string
		frontmatter string
		expected    string
	}{
		{
			name:        "Category from front matter",
			milestone:   "Milestone Category",
			frontmatter: "category: Custom Category",
			expected:    "Custom Category",
		},
		{
			name:        "Category from milestone when no front matter",
			milestone:   "Milestone Category",
			frontmatter: "",
			expected:    "Milestone Category",
		},
		{
			name:        "Category from milestone when front matter has no category",
			milestone:   "Milestone Category",
			frontmatter: "category: ",
			expected:    "Milestone Category",
		},
		{
			name:        "Category is selected from categories list in front matter",
			milestone:   "Milestone Category",
			frontmatter: "categories:\n  - First Category\n  - Second Category",
			expected:    "First Category",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			issue := &github.Issue{
				Milestone: &github.Milestone{
					Title: github.String(tt.milestone),
				},
				Body: github.String(body(t, tt.frontmatter, "This is the body.")),
			}
			article, err := NewIssueArticle(Markdown, issue)
			if err != nil {
				t.Fatalf("Failed to create IssueArticle: %v", err)
			}

			if article.Category() != tt.expected {
				t.Errorf("Expected category %q, got %q", tt.expected, article.Category())
			}
		})
	}
}

func TestIssueArticle_Content(t *testing.T) {
	fm := `title: Sample Title
author: Sample Author
category: Sample Category
	`
	content := "\n# This is the content\n\nHere is some more detailed content."

	issue := &github.Issue{
		Body: github.String(body(t, fm, content)),
	}
	article, err := NewIssueArticle(Markdown, issue)
	if err != nil {
		t.Fatalf("Failed to create IssueArticle: %v", err)
	}

	if article.Content() != content {
		t.Errorf("Expected content %q, got %q", content, article.Content())
	}
}

func TestIssueArticle_Date(t *testing.T) {
	tests := []struct {
		name        string
		createdAt   time.Time
		frontmatter string
		expected    time.Time
	}{
		{
			name:        "Date Only",
			createdAt:   time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
			frontmatter: "",
			expected:    time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
		},
		{
			name:        "Date Only (prefer front matter value)",
			createdAt:   time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
			frontmatter: "date: 2025-12-25",
			expected:    time.Date(2025, 12, 25, 0, 0, 0, 0, time.UTC),
		},
		{
			name:        "RFC3339",
			createdAt:   time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
			frontmatter: "date: 2025-12-25T15:30:00Z",
			expected:    time.Date(2025, 12, 25, 15, 30, 0, 0, time.UTC),
		},
		{
			name:        "Date Only String",
			createdAt:   time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
			frontmatter: `date: "2025-11-11"`,
			expected:    time.Date(2025, 11, 11, 0, 0, 0, 0, time.UTC),
		},
		{
			name:        "RFC3339 String",
			createdAt:   time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
			frontmatter: `date: "2025-10-10T10:10:10Z"`,
			expected:    time.Date(2025, 10, 10, 10, 10, 10, 0, time.UTC),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			issue := &github.Issue{
				CreatedAt: &github.Timestamp{Time: tt.createdAt},
				Body:      github.String(body(t, tt.frontmatter, "This is the body.")),
			}
			article, err := NewIssueArticle(Markdown, issue)
			if err != nil {
				t.Fatalf("Failed to create IssueArticle: %v", err)
			}

			if !article.Date().Equal(tt.expected) {
				t.Errorf("Expected date %v, got %v", tt.expected, article.Date())
			}
		})
	}
}

func TestIssueArticle_Images(t *testing.T) {
	frontmatter := `title: Sample Title
author: Sample Author
category: Sample Category
	`
	content := `
# This is the content

Here is an image:

![Alt Text](https://example.com/image1.png)

Another image:

![Another Image](https://example.com/image2.jpg)
`

	issue := &github.Issue{
		Body: github.String(body(t, frontmatter, content)),
	}
	article, err := NewIssueArticle(Markdown, issue)
	if err != nil {
		t.Fatalf("Failed to create IssueArticle: %v", err)
	}

	images := article.Images()
	if len(images) != 2 {
		t.Fatalf("Expected 2 images, got %d", len(images))
	}

	expectedSrcs := []string{
		"https://example.com/image1.png",
		"https://example.com/image2.jpg",
	}

	for i, img := range images {
		if string(img.Destination()) != expectedSrcs[i] {
			t.Errorf("Expected image %d src %q, got %q", i, expectedSrcs[i], string(img.Destination()))
		}
	}
}

func TestIssueArticle_IsDraft(t *testing.T) {
	tests := []struct {
		name        string
		isClosed    bool
		frontmatter string
		expected    bool
	}{
		{
			name:        "Draft from front matter true",
			isClosed:    false,
			frontmatter: "draft: true",
			expected:    true,
		},
		{
			name:        "Draft from front matter false",
			isClosed:    false,
			frontmatter: "draft: false",
			expected:    false,
		},
		{
			name:        "Draft from issue when front matter absent",
			isClosed:    false,
			frontmatter: "",
			expected:    true,
		},
		{
			name:        "Not draft from issue when closed",
			isClosed:    true,
			frontmatter: "",
			expected:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			state := "open"
			if !tt.isClosed {
				state = "closed"
			}
			issue := &github.Issue{
				State: github.String(state),
				Body:  github.String(body(t, tt.frontmatter, "This is the body.")),
			}
			article, err := NewIssueArticle(Markdown, issue)
			if err != nil {
				t.Fatalf("Failed to create IssueArticle: %v", err)
			}

			if article.IsDraft() != tt.expected {
				t.Errorf("Expected IsDraft %v, got %v", tt.expected, article.IsDraft())
			}
		})
	}
}

func TestIssueArticle_Tags(t *testing.T) {
	tests := []struct {
		name        string
		labels      []string
		frontmatter string
		expected    []string
	}{
		{
			name:        "Tags from labels only",
			labels:      []string{"bug", "enhancement"},
			frontmatter: "",
			expected:    []string{"bug", "enhancement"},
		},
		{
			name:        "Tags from front matter only",
			labels:      []string{},
			frontmatter: "tags:\n  - feature\n  - documentation",
			expected:    []string{"feature", "documentation"},
		},
		{
			name:        "Tags from both labels and front matter",
			labels:      []string{"bug"},
			frontmatter: "tags:\n  - urgent\n  - bug",
			expected:    []string{"bug", "urgent"},
		},
		{
			name:        "Duplicate tags from labels and front matter",
			labels:      []string{"bug", "feature"},
			frontmatter: "tags:\n  - feature\n  - improvement",
			expected:    []string{"bug", "feature", "improvement"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var ghLabels []*github.Label
			for _, label := range tt.labels {
				ghLabels = append(ghLabels, &github.Label{Name: github.String(label)})
			}
			issue := &github.Issue{
				Labels: ghLabels,
				Body:   github.String(body(t, tt.frontmatter, "This is the body.")),
			}
			article, err := NewIssueArticle(Markdown, issue)
			if err != nil {
				t.Fatalf("Failed to create IssueArticle: %v", err)
			}

			gotTags := article.Tags()
			if len(gotTags) != len(tt.expected) {
				t.Fatalf("Expected %d tags, got %d", len(tt.expected), len(gotTags))
			}
			tagSet := make(map[string]struct{})
			for _, tag := range gotTags {
				tagSet[tag] = struct{}{}
			}
			for _, expectedTag := range tt.expected {
				if _, found := tagSet[expectedTag]; !found {
					t.Errorf("Expected tag %q not found in result", expectedTag)
				}
			}
		})
	}
}

func TestIssueArticle_FrontMatter(t *testing.T) {
	t.Run("preserves markdown front matter and merges issue attributes", func(t *testing.T) {
		issue := &github.Issue{
			Title: github.String("Issue Title"),
			User:  &github.User{Login: github.String("issue-user")},
			CreatedAt: &github.Timestamp{
				Time: time.Date(2025, 1, 11, 0, 0, 0, 0, time.UTC),
			},
			Labels: []*github.Label{
				{Name: github.String("shared")},
				{Name: github.String("bug")},
			},
			Milestone: &github.Milestone{
				Title: github.String("Milestone Category"),
			},
			Body: github.String(body(t, `title: Front Title
author: Front Author
categories:
  - Front
  - Secondary
date: 2025-01-10
draft: true
tags:
  - frontend
  - shared
foo: bar`, "Body")),
		}
		article, err := NewIssueArticle(Markdown, issue)
		if err != nil {
			t.Fatalf("Failed to create IssueArticle: %v", err)
		}

		fm, err := article.FrontMatter()
		if err != nil {
			t.Fatalf("Failed to get front matter: %v", err)
		}
		var got map[string]any
		if err := yaml.Unmarshal([]byte(fm), &got); err != nil {
			t.Fatalf("Failed to unmarshal front matter: %v", err)
		}

		if got["title"] != "Front Title" {
			t.Errorf("Expected title %q, got %v", "Front Title", got["title"])
		}
		if got["author"] != "Front Author" {
			t.Errorf("Expected author %q, got %v", "Front Author", got["author"])
		}
		categories, ok := got["categories"].([]any)
		if !ok || len(categories) != 2 || categories[0] != "Front" || categories[1] != "Secondary" {
			t.Errorf("Expected categories [Front Secondary], got %v", got["categories"])
		}
		if draft, ok := got["draft"].(bool); !ok || draft != true {
			t.Errorf("Expected draft %v, got %v", true, got["draft"])
		}
		if got["foo"] != "bar" {
			t.Errorf("Expected custom front matter value, got %v", got["foo"])
		}
		date := dateFromFrontMatterValue(t, got["date"])
		if !date.Equal(time.Date(2025, 1, 10, 0, 0, 0, 0, time.UTC)) {
			t.Errorf("Expected date 2025-01-10T00:00:00Z, got %v", date)
		}
		actualTags := flattenTags(t, got["tags"])
		expectedTags := []string{"bug", "frontend", "shared"}
		if !compareStringSlices(actualTags, expectedTags) {
			t.Errorf("Expected tags %v, got %v", expectedTags, actualTags)
		}
	})

	t.Run("fills missing fields from issue and milestone", func(t *testing.T) {
		createdAt := time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC)
		closedAt := time.Date(2025, 1, 2, 13, 30, 0, 0, time.UTC)
		issue := &github.Issue{
			Title: github.String("Fallback Title"),
			User:  &github.User{Login: github.String("issue-user")},
			State: github.String("closed"),
			Labels: []*github.Label{
				{Name: github.String("bug")},
				{Name: github.String("feature")},
			},
			Milestone: &github.Milestone{
				Title: github.String("Milestone"),
			},
			CreatedAt: &github.Timestamp{Time: createdAt},
			ClosedAt:  &github.Timestamp{Time: closedAt},
			Body:      github.String("Body"),
		}
		article, err := NewIssueArticle(Markdown, issue)
		if err != nil {
			t.Fatalf("Failed to create IssueArticle: %v", err)
		}

		fm, err := article.FrontMatter()
		if err != nil {
			t.Fatalf("Failed to get front matter: %v", err)
		}
		var got map[string]any
		if err := yaml.Unmarshal([]byte(fm), &got); err != nil {
			t.Fatalf("Failed to unmarshal front matter: %v", err)
		}

		if got["title"] != "Fallback Title" {
			t.Errorf("Expected title %q, got %v", "Fallback Title", got["title"])
		}
		if got["author"] != "issue-user" {
			t.Errorf("Expected author %q, got %v", "issue-user", got["author"])
		}
		if got["category"] != "Milestone" {
			t.Errorf("Expected category %q, got %v", "Milestone", got["category"])
		}
		if draft, ok := got["draft"].(bool); !ok || draft != true {
			t.Errorf("Expected draft %v, got %v", true, got["draft"])
		}
		date := dateFromFrontMatterValue(t, got["date"])
		if !date.Equal(closedAt) {
			t.Errorf("Expected date %v, got %v", closedAt, date)
		}

		actualTags := flattenTags(t, got["tags"])
		expectedTags := []string{"bug", "feature"}
		if !compareStringSlices(actualTags, expectedTags) {
			t.Errorf("Expected tags %v, got %v", expectedTags, actualTags)
		}
	})
}

func flattenTags(t *testing.T, value any) []string {
	t.Helper()
	raw, ok := value.([]any)
	if !ok {
		t.Fatalf("Expected tags to be []any, got %T", value)
	}
	tags := make([]string, 0, len(raw))
	for _, value := range raw {
		tag, ok := value.(string)
		if !ok {
			t.Fatalf("Expected tag string, got %T", value)
		}
		tags = append(tags, tag)
	}
	sort.Strings(tags)
	return tags
}

func compareStringSlices(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func dateFromFrontMatterValue(t *testing.T, value any) time.Time {
	t.Helper()
	switch date := value.(type) {
	case time.Time:
		return date
	case string:
		parsed, err := time.Parse(time.RFC3339, date)
		if err == nil {
			return parsed
		}
		parsed, err = time.Parse(time.DateOnly, date)
		if err == nil {
			return parsed
		}
		t.Fatalf("Failed to parse date %q", date)
	default:
		t.Fatalf("Expected date string or time.Time, got %T", value)
	}
	return time.Time{}
}
