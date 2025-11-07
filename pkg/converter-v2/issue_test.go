package converter_v2

import (
	"strings"
	"testing"
	"time"

	"github.com/google/go-github/v67/github"
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
