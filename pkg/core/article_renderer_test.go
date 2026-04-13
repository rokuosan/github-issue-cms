package core

import (
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
)

func TestHugoArticleRenderer_Render(t *testing.T) {
	renderer := NewHugoArticleRenderer()

	tests := []struct {
		name    string
		article *Article
		want    string
		wantErr bool
	}{
		{
			name: "basic render",
			article: &Article{
				Author:      "John Doe",
				Title:       "Test Title",
				Content:     "Test content",
				Date:        "2021-01-01T00:00:00Z",
				Category:    "Test Category",
				Tags:        []string{"tag1", "tag2"},
				Draft:       false,
				FrontMatter: EmptyFrontMatter(),
			},
			want: `---
author: John Doe
title: Test Title
date: "2021-01-01T00:00:00Z"
categories: Test Category
tags:
    - tag1
    - tag2
draft: false
{}
---

Test content
`,
		},
		{
			name: "additional front matter",
			article: &Article{
				Author:      "John Doe",
				Title:       "Test Title",
				Content:     "Test content",
				Date:        "2021-01-01T00:00:00Z",
				Category:    "Test",
				Tags:        []string{"test"},
				Draft:       false,
				FrontMatter: NewFrontMatter(map[string]any{"custom_field": "custom_value"}),
			},
			want: `---
author: John Doe
title: Test Title
date: "2021-01-01T00:00:00Z"
categories: Test
tags:
    - test
draft: false
custom_field: custom_value
---

Test content
`,
		},
		{
			name: "override known fields",
			article: &Article{
				Author:      "Original Author",
				Title:       "Original Title",
				Content:     "Test content",
				Date:        "2021-01-01",
				Category:    "Original",
				Tags:        []string{"original"},
				Draft:       true,
				FrontMatter: NewFrontMatter(map[string]any{"author": "Override Author", "title": "Override Title"}),
			},
			want: `---
author: Override Author
title: Override Title
date: "2021-01-01"
categories: Original
tags:
    - original
draft: true
{}
---

Test content
`,
		},
		{
			name: "invalid front matter value",
			article: &Article{
				Author:  "John Doe",
				Title:   "Test",
				Content: "Test",
				Date:    "2021-01-01",
				FrontMatter: NewFrontMatter(map[string]any{
					"broken": []any{map[string]any{"nested": make(chan int)}},
				}),
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := renderer.Render(tt.article)
			if (err != nil) != tt.wantErr {
				t.Errorf("Render() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				assertEqualCmp(t, tt.want, got)
			}
		})
	}
}

func TestHugoArticleRenderer_DoesNotMutateReceiver(t *testing.T) {
	renderer := NewHugoArticleRenderer()
	article := &Article{
		Author:      "Original Author",
		Title:       "Original Title",
		Content:     "Test content",
		Date:        "2021-01-01",
		Category:    "Original Category",
		Tags:        []string{"original"},
		Draft:       true,
		FrontMatter: NewFrontMatter(map[string]any{"author": "Override Author", "tags": []any{"override"}}),
	}

	_, err := renderer.Render(article)
	if err != nil {
		t.Fatalf("Render() error = %v", err)
	}

	want := &Article{
		Author:      "Original Author",
		Title:       "Original Title",
		Content:     "Test content",
		Date:        "2021-01-01",
		Category:    "Original Category",
		Tags:        []string{"original"},
		Draft:       true,
		FrontMatter: NewFrontMatter(map[string]any{"author": "Override Author", "tags": []any{"override"}}),
	}
	assertEqualCmp(t, want, article, cmp.Comparer(func(x, y FrontMatter) bool {
		return cmp.Equal(x.Values(), y.Values())
	}))
}

func TestArticle_ParseDateTime(t *testing.T) {
	tests := []struct {
		name    string
		date    string
		want    time.Time
		wantErr bool
	}{
		{
			name:    "ISO8601 format",
			date:    "2021-01-01T00:00:00Z",
			want:    time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC),
			wantErr: false,
		},
		{
			name:    "date only format",
			date:    "2021-01-01",
			want:    time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC),
			wantErr: false,
		},
		{
			name:    "invalid format",
			date:    "invalid-date",
			want:    time.Time{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			article := &Article{Date: tt.date}
			got, err := article.ParseDateTime()
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseDateTime() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				assertEqualCmp(t, tt.want, got)
			}
		})
	}
}
