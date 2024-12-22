package convert

import (
	"github.com/google/go-github/v67/github"
	"github.com/rokuosan/github-issue-cms/internal/api"
	"reflect"
	"testing"
	"time"
)

func TestIssueToArticleConverter_Convert_Convert(t *testing.T) {
	c := &IssueToArticleConverter{GitHub: &mockGitHub{}}
	expects := []*Article{
		{
			Title:      "Title 1",
			Content:    "Body 1\n",
			Author:     "user",
			Date:       "2024-12-18T01:02:03Z",
			Categories: []string{"category"},
			Tags:       []string{"tag1"},
			Draft:      true,
		},
		{
			Title:            "Title 2",
			Content:          "Body 2\n",
			Author:           "user",
			Date:             "2024-12-18T01:02:03Z",
			Categories:       []string{"category"},
			Tags:             []string{"tag1"},
			Draft:            false,
			ExtraFrontMatter: "key: value\n",
		},
	}

	articles, err := c.Convert()
	if err != nil {
		t.Error(err)
	}

	for i, article := range articles {
		if !reflect.DeepEqual(article, expects[i]) {
			t.Errorf("got: %#v, want: %#v", article, expects[i])
		}
	}
}

type mockGitHub struct{}

func (m *mockGitHub) GetIssues() []*api.GitHubIssue {
	return []*api.GitHubIssue{
		{Issue: &github.Issue{
			Title: github.String("Title 1"),
			Body:  github.String("Body 1"),
			CreatedAt: &github.Timestamp{
				Time: time.Date(2024, 12, 18, 1, 2, 3, 0, time.UTC),
			},
			User:      &github.User{Login: github.String("user")},
			Milestone: &github.Milestone{Title: github.String("category")},
			Labels: []*github.Label{
				{Name: github.String("tag1")},
			},
			State: github.String("open"),
		}},
		{Issue: &github.Issue{
			Title: github.String("Title 2"),
			Body:  github.String("```\r\nkey: value\r\n```\r\nBody 2"),
			CreatedAt: &github.Timestamp{
				Time: time.Date(2024, 12, 18, 1, 2, 3, 0, time.UTC),
			},
			User:      &github.User{Login: github.String("user")},
			Milestone: &github.Milestone{Title: github.String("category")},
			Labels: []*github.Label{
				{Name: github.String("tag1")},
			},
			State: github.String("closed"),
		}},
	}
}
