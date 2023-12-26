package internal

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/google/go-github/v56/github"
	"github.com/spf13/viper"
)

func GetIssues() []*github.Issue {
	if GitHubClient == nil {
		SetupGitHubClient()
		return GetIssues()
	}

	// Get Issues
	client := GitHubClient
	username := viper.GetString("github.username")
	repository := viper.GetString("github.repository")
	if username == "" || repository == "" {
		Logger.Error("Please set username and repository in gic.config.yaml")
		return nil
	}

	issuesAndPRs, _, err := client.Issues.ListByRepo(
		context.Background(),
		username,
		repository,
		&github.IssueListByRepoOptions{
			State: "all",
		},
	)

	// Check Rate Limits
	var rateLimitError *github.RateLimitError
	if errors.As(err, &rateLimitError) {
		Logger.Error("hit rate limit")
	}
	var abuseRateLimitError *github.AbuseRateLimitError
	if errors.As(err, &abuseRateLimitError) {
		Logger.Error("hit secondary rate limit")
	}
	if err != nil {
		panic(err)
	}

	// Filter issues
	var issues []*github.Issue
	for _, item := range issuesAndPRs {
		// Skip if it is a pull request.
		if item.IsPullRequest() {
			continue
		}
		issues = append(issues, item)
	}

	return issues
}

func IssueToArticle(issue *github.Issue) *Article {
	// Skip if it is a pull request.
	if issue.IsPullRequest() {
		return nil
	}

	// Get ID
	id := fmt.Sprintf("%d", *issue.ID)
	Logger.Debug("Processing issue ID: " + id)

	// Get content
	content := issue.GetBody()
	content = strings.Replace(content, "\r", "", -1)

	// Get front matter
	frontMatter := func() []string {
		re := regexp.MustCompile("(?s)^\\s*```[\\n|\\r]([^`]*)[\\n|\\r]```")
		match := re.FindStringSubmatch(content)
		if len(match) > 0 {
			return match
		}

		return nil
	}()

	// Remove front matter from content
	if frontMatter != nil {
		content = strings.Replace(content, frontMatter[0], "", 1)
	}

	// Remove empty lines at the beginning
	content = strings.TrimLeft(content, "\n")

	// Insert empty line at the end if not exists
	if !strings.HasSuffix(content, "\n") {
		content = content + "\n"
	}

	// Replace image URL to local path
	re := regexp.MustCompile(`!\[image*]\((.*)\)`)
	match := re.FindAllStringSubmatch(content, -1)
	for i, m := range match {
		url := m[1]
		before := m[0]
		replaced := "![" + url + "](/images/" + id + "/" + fmt.Sprintf("%d", i) + ".png)"

		// Skip if already replaced
		if strings.Contains(content, replaced) {
			continue
		}

		// Download image
		DownloadImage(url, id, i)

		// Replace url to local path
		content = strings.Replace(content, before, replaced, -1)
	}

	re = regexp.MustCompile(`<img width="\d+" alt="(\w+)" src="(\S+)">`)
	match = re.FindAllStringSubmatch(content, -1)
	for i, m := range match {
		alt := m[1]
		url := m[2]
		before := m[0]
		replaced := "![" + alt + "](/images/" + id + "/" + fmt.Sprintf("%d", i) + ".png)"

		fmt.Println("Replace: " + url)
		DownloadImage(url, id, i)

		content = strings.Replace(content, before, replaced, -1)
	}

	// Create article
	return &Article{
		Author:           issue.GetUser().GetLogin(),
		Title:            issue.GetTitle(),
		Date:             issue.GetCreatedAt().Format("2006-01-02T15:04:05Z"),
		Category:         issue.GetMilestone().GetTitle(),
		Draft:            issue.GetState() != "closed",
		Content:          content,
		ExtraFrontMatter: frontMatter[1],
		Tags: func() []string {
			tags := []string{}
			for _, label := range issue.Labels {
				tags = append(tags, label.GetName())
			}
			return tags
		}(),
	}
}

func ExportArticle(article *Article, id string) {
	// Create directory
	path := filepath.Join("content", "posts")
	if _, err := os.Stat(path); os.IsNotExist(err) {
		if err := os.MkdirAll(path, 0777); err != nil {
			panic(err)
		}
	}

	// Create file
	path = filepath.Join("content", "posts", id+".md")
	file, err := os.Create(path)
	if err != nil {
		panic(err)
	}
	defer file.Close()
	Logger.Info("Export article: " + path)

	// Build string
	content := make([]byte, 0, 1024)
	tags := make([]byte, 0, 128)
	for _, t := range article.Tags {
		tags = append(tags, "  - "...)
		tags = append(tags, t...)
		tags = append(tags, '\n')
	}
	raw := strings.TrimSpace(fmt.Sprintf(`---
title: %s
author: %s
date: %s
draft: %t
categories: 
  - %s
tags:
%s
%s
---

%s`, article.Title,
		article.Author,
		article.Date,
		article.Draft,
		article.Category,
		tags,
		article.ExtraFrontMatter,
		article.Content))
	content = append(content, raw...)

	// Write file
	w := bufio.NewWriter(file)
	_, _ = w.WriteString(string(content))
	_ = w.Flush()
}
