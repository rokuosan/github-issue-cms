package converter_test

import (
	"testing"
	"time"

	"github.com/google/go-github/v67/github"
	"github.com/rokuosan/github-issue-cms/pkg/config"
	"github.com/rokuosan/github-issue-cms/pkg/converter"
	converter_mock "github.com/rokuosan/github-issue-cms/pkg/converter/mock"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

// Example test using MockConverter
func TestMockConverter(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Create a mock converter
	mock := converter_mock.NewMockConverter(ctrl)

	// Set up expectations for GetIssues
	issue1 := &github.Issue{
		Number: github.Int(1),
		Title:  github.String("Test Issue 1"),
		State:  github.String("open"),
	}
	issue2 := &github.Issue{
		Number: github.Int(2),
		Title:  github.String("Test Issue 2"),
		State:  github.String("closed"),
	}
	mock.EXPECT().GetIssues().Return([]*github.Issue{issue1, issue2})

	// Set up expectations for IssueToArticle
	mock.EXPECT().IssueToArticle(issue1).Return(&converter.Article{
		Title:    issue1.GetTitle(),
		Draft:    issue1.GetState() == "open",
		Date:     time.Now().Format("2006-01-02T15:04:05Z"),
		Category: "test",
		Content:  "Test content",
	})

	// Test GetIssues
	issues := mock.GetIssues()
	assert.Equal(t, 2, len(issues))
	assert.Equal(t, "Test Issue 1", issues[0].GetTitle())

	// Test IssueToArticle
	article := mock.IssueToArticle(issues[0])
	assert.NotNil(t, article)
	assert.Equal(t, "Test Issue 1", article.Title)
	assert.True(t, article.Draft)
}

// Test that converterImpl implements Converter interface
func TestConverterInterface(t *testing.T) {
	var c = converter.NewConverter(converter.Config{Config: *config.NewConfig()}, "test-token")
	assert.NotNil(t, c)
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

func TestConverter_IssueToArticle(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mock := converter_mock.NewMockConverter(ctrl)

	tests := []struct {
		name     string
		issue    *github.Issue
		expected *converter.Article
	}{
		{
			name: "valid",
			issue: &github.Issue{
				Title:     stringPtr("Test"),
				Body:      stringPtr("This is a test"),
				CreatedAt: parseTime("2021-01-01T00:00:00Z"),
				Labels:    []*github.Label{},
			},
			expected: &converter.Article{
				Title:   "Test",
				Content: "This is a test\n",
				Date:    "2021-01-01T00:00:00Z",
				Key:     "2021-01-01_000000",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock.EXPECT().IssueToArticle(tt.issue).Return(tt.expected)
			result := mock.IssueToArticle(tt.issue)

			assert.Equal(t, tt.expected, result)
		})
	}
}
