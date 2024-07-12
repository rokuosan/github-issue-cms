package converter

import "testing"

func Test_Article_Transform(t *testing.T) {
	article := &Article{
		Author:           "John Doe",
		Title:            "Hello, World!",
		Content:          "This is a test content.",
		Date:             "2021-01-01T00:00:00Z",
		Category:         "Test",
		Tags:             []string{"test", "sample"},
		Draft:            false,
		ExtraFrontMatter: "extra: true\nauthors:\n    - TEST\n",
		Key:              "hello-world",
	}

	got, err := article.Transform()
	if err != nil {
		t.Fatalf("Article.Transform() error = %v", err)
	}

	want := `---
author: John Doe
title: Hello, World!
date: "2021-01-01T00:00:00Z"
categories: Test
tags:
    - test
    - sample
draft: false
authors:
    - TEST
extra: true
---

This is a test content.
`

	if got != want {
		t.Errorf("Article.Transform() = %v, want %v", got, want)
	}
}
