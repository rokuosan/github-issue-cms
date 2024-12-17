package convert

import (
	"github.com/google/go-github/v67/github"
	"github.com/rokuosan/github-issue-cms/internal/api"
	"reflect"
	"testing"
	"time"
)

func TestIssue_ConvertToArticle(t *testing.T) {
	tests := []struct {
		name    string
		i       *issue
		want    *Article
		wantErr bool
	}{
		{
			name: "front matter が存在しない場合",
			i: &issue{GitHubIssue: &api.GitHubIssue{Issue: &github.Issue{
				Title:     github.String("This is a sample issue title"),
				Body:      github.String("This is a sample issue body"),
				CreatedAt: &github.Timestamp{Time: time.Date(2024, 12, 18, 1, 2, 3, 0, time.UTC)},
				User:      &github.User{Login: github.String("user")},
				Milestone: &github.Milestone{Title: github.String("category")},
				Labels: []*github.Label{
					{Name: github.String("tag1")},
					{Name: github.String("tag2")},
					{Name: github.String("tag3")},
				},
				State: github.String("open"),
			}}},
			want: &Article{
				Title:      "This is a sample issue title",
				Content:    "This is a sample issue body\n",
				Author:     "user",
				Date:       "2024-12-18T01:02:03Z",
				Categories: []string{"category"},
				Tags:       []string{"tag1", "tag2", "tag3"},
				Draft:      true,
			},
		},
		{
			name: "front matter にある内容が優先される",
			i: &issue{GitHubIssue: &api.GitHubIssue{Issue: &github.Issue{
				Title:     github.String("This is a sample issue title"),
				Body:      github.String("```\nauthor: author\n```\nThis is a sample issue body"),
				CreatedAt: &github.Timestamp{Time: time.Date(2024, 12, 18, 1, 2, 3, 0, time.UTC)},
				User:      &github.User{Login: github.String("user")},
				Milestone: &github.Milestone{Title: github.String("category")},
				Labels: []*github.Label{
					{Name: github.String("tag1")},
					{Name: github.String("tag2")},
					{Name: github.String("tag3")},
				},
				State: github.String("open"),
			}}},
			want: &Article{
				Title:      "This is a sample issue title",
				Content:    "This is a sample issue body\n",
				Author:     "author",
				Date:       "2024-12-18T01:02:03Z",
				Categories: []string{"category"},
				Tags:       []string{"tag1", "tag2", "tag3"},
				Draft:      true,
			},
		},
		{
			name: "front matter に任意のフィールドが存在する場合",
			i: &issue{GitHubIssue: &api.GitHubIssue{Issue: &github.Issue{
				Title:     github.String("This is a sample issue title"),
				Body:      github.String("```\nauthor: author\nsample: example\n```\nThis is a sample issue body"),
				CreatedAt: &github.Timestamp{Time: time.Date(2024, 12, 18, 1, 2, 3, 0, time.UTC)},
				User:      &github.User{Login: github.String("user")},
				Milestone: &github.Milestone{Title: github.String("category")},
				Labels: []*github.Label{
					{Name: github.String("tag1")},
					{Name: github.String("tag2")},
					{Name: github.String("tag3")},
				},
				State: github.String("open"),
			}}},
			want: &Article{
				Title:            "This is a sample issue title",
				Content:          "This is a sample issue body\n",
				Author:           "author",
				Date:             "2024-12-18T01:02:03Z",
				Categories:       []string{"category"},
				Tags:             []string{"tag1", "tag2", "tag3"},
				Draft:            true,
				ExtraFrontMatter: "sample: example\n",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			i := tt.i
			got, err := i.ConvertToArticle()
			if (err != nil) != tt.wantErr {
				t.Errorf("issue.ConvertToArticle() error = %#v, wantErr %#v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("issue.ConvertToArticle() = %#v, want %#v", got, tt.want)
			}
		})
	}
}

func TestIssue_body(t *testing.T) {
	tests := []struct {
		name string
		i    *issue
		want string
	}{
		{
			name: "末尾に改行が挿入される",
			i: &issue{GitHubIssue: &api.GitHubIssue{Issue: &github.Issue{
				Body: github.String("This is a sample issue body"),
			}}},
			want: "This is a sample issue body\n",
		},
		{
			name: "空の文字列であっても末尾に改行が挿入される",
			i: &issue{GitHubIssue: &api.GitHubIssue{Issue: &github.Issue{
				Body: github.String(""),
			}}},
			want: "\n",
		},
		{
			name: "すでに末尾に改行がある場合は、新たに改行されることはない",
			i: &issue{GitHubIssue: &api.GitHubIssue{Issue: &github.Issue{
				Body: github.String("This is a sample issue body\n"),
			}}},
			want: "This is a sample issue body\n",
		},
		{
			name: "改行文字のトリムと、行頭復帰文字の削除が行われる",
			i: &issue{GitHubIssue: &api.GitHubIssue{Issue: &github.Issue{
				Body: github.String("\r\n\r\n\r\n\n\n\nThis is a\r\n sample issue body\n\r\n"),
			}}},
			want: "This is a\n sample issue body\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			i := tt.i
			if got := i.body(); got != tt.want {
				t.Errorf("issue.body() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIssue_content(t *testing.T) {
	tests := []struct {
		name string
		i    *issue
		want string
	}{
		{
			name: "front matter が存在しない場合はそのままの内容が返る",
			i: &issue{GitHubIssue: &api.GitHubIssue{Issue: &github.Issue{
				Body: github.String("This is a sample issue body"),
			}}},
			want: "This is a sample issue body\n",
		},
		{
			name: "front matter が存在する場合は front matter を削除した内容が返る",
			i: &issue{GitHubIssue: &api.GitHubIssue{Issue: &github.Issue{
				Body: github.String("```\nkey: value\n```\nThis is a sample issue body"),
			}}},
			want: "This is a sample issue body\n",
		},
		{
			name: "front matter が存在するが、中身が空の場合は front matter だけが削除される",
			i: &issue{GitHubIssue: &api.GitHubIssue{Issue: &github.Issue{
				Body: github.String("```\n\n```\nThis is a sample issue body"),
			}}},
			want: "This is a sample issue body\n",
		},
		{
			name: "front matter が存在するが、YAML として不正な場合は front matter だけが削除される",
			i: &issue{GitHubIssue: &api.GitHubIssue{Issue: &github.Issue{
				Body: github.String("```\nkey=value\n```\nThis is a sample issue body"),
			}}},
			want: "This is a sample issue body\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			i := tt.i
			if got := i.content(); got != tt.want {
				t.Errorf("issue.content() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestIssue_frontMatter(t *testing.T) {
	tests := []struct {
		name    string
		i       *issue
		want    string
		wantErr bool
	}{
		{
			name: "front matter が存在しない場合はエラーが返る",
			i: &issue{GitHubIssue: &api.GitHubIssue{Issue: &github.Issue{
				Body: github.String("This is a sample issue body"),
			}}},
			want: "",
		},
		{
			name: "front matter が存在する場合は front matter が返る",
			i: &issue{GitHubIssue: &api.GitHubIssue{Issue: &github.Issue{
				Body: github.String("```\nkey: value\n```\nThis is a sample issue body"),
			}}},
			want: "key: value",
		},
		{
			name: "front matter が存在するが、中身が空の場合は空文字列が返る",
			i: &issue{GitHubIssue: &api.GitHubIssue{Issue: &github.Issue{
				Body: github.String("```\n\n```\nThis is a sample issue body"),
			}}},
			want: "",
		},
		{
			name: "front matter が存在するが、YAML として不正な場合はエラーが返る",
			i: &issue{GitHubIssue: &api.GitHubIssue{Issue: &github.Issue{
				Body: github.String("```\nkey=value\n```\nThis is a sample issue body"),
			}}},
			want:    "",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			i := tt.i
			got, err := i.frontMatter()
			if (err != nil) != tt.wantErr {
				t.Errorf("issue.frontMatter() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("issue.frontMatter() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIssue_tags(t *testing.T) {
	tests := []struct {
		name string
		i    *issue
		want []string
	}{
		{
			name: "empty",
			i:    &issue{GitHubIssue: &api.GitHubIssue{Issue: &github.Issue{}}},
			want: []string{},
		},
		{
			name: "single",
			i: &issue{
				GitHubIssue: &api.GitHubIssue{
					Issue: &github.Issue{
						Labels: []*github.Label{
							{Name: github.String("tag1")},
						},
					},
				},
			},
			want: []string{"tag1"},
		},
		{
			name: "multiple",
			i: &issue{
				GitHubIssue: &api.GitHubIssue{
					Issue: &github.Issue{
						Labels: []*github.Label{
							{Name: github.String("tag1")},
							{Name: github.String("tag2")},
							{Name: github.String("tag3")},
						},
					},
				},
			},
			want: []string{"tag1", "tag2", "tag3"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			i := tt.i
			if got := i.tags(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("issue.tags() = %v, want %v", got, tt.want)
			}
		})
	}
}
