package convert

import (
	"reflect"
	"testing"
)

func TestArticle_markdownLinks(t *testing.T) {
	create := func(content string) *Article {
		return &Article{Content: content}
	}

	tests := []struct {
		name    string
		article *Article
		want    []string
	}{
		{
			name:    "no links",
			article: create(""),
			want:    []string{},
		},
		{
			name:    "markdown link",
			article: create("[text](url)"),
			want:    []string{"[text](url)"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.article.markdownLinks(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Article.markdownLinks() = %#v, want %#v", got, tt.want)
			}
		})
	}
}

func TestArticle_images(t *testing.T) {
	create := func(content string) *Article {
		return &Article{Content: content}
	}

	tests := []struct {
		name    string
		article *Article
		want    []ArticleImage
	}{
		{
			name:    "no images",
			article: create(""),
			want:    []ArticleImage{},
		},
		{
			name:    "markdown image",
			article: create("![alt](src)"),
			want: []ArticleImage{
				{
					Source:   "src",
					Alt:      "alt",
					Original: "![alt](src)",
				},
			},
		},
		{
			name:    "HTML image",
			article: create("<img src=\"src\" alt=\"alt\">"),
			want: []ArticleImage{
				{
					Source:   "src",
					Alt:      "alt",
					Original: "<img src=\"src\" alt=\"alt\">",
				},
			},
		},
		{
			name:    "HTML image with alt first",
			article: create("<img alt=\"alt\" src=\"src\">"),
			want: []ArticleImage{
				{
					Source:   "src",
					Alt:      "alt",
					Original: "<img alt=\"alt\" src=\"src\">",
				},
			},
		},
		{
			name: "multiple images",
			article: create(`
				![alt1](src1)
				<img src="src2" alt="alt2">
				<img alt="alt3" src="src3">
			`),
			want: []ArticleImage{
				{
					Source:   "src1",
					Alt:      "alt1",
					Original: "![alt1](src1)",
				},
				{
					Source:   "src2",
					Alt:      "alt2",
					Original: "<img src=\"src2\" alt=\"alt2\">",
				},
				{
					Source:   "src3",
					Alt:      "alt3",
					Original: "<img alt=\"alt3\" src=\"src3\">",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.article.images(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Article.images() = %#v, want %#v", got, tt.want)
			}
		})
	}
}
