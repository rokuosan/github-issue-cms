package internal

import (
	"bufio"
	"context"
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
	if _, ok := err.(*github.RateLimitError); ok {
		Logger.Error("hit rate limit")
	}
	if _, ok := err.(*github.AbuseRateLimitError); ok {
		Logger.Error("hit secondary rate limit")
	}
	if err != nil {
		panic(err)
	}

	// Filter issues
	issues := []*github.Issue{}
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

	// Get front matter
	frontMatter := func() []string {
		re := regexp.MustCompile("(?s)^\\s*```[\\n|\\r|\\n\\r]([^`]*)[\\n|\\r|\\n\\r]```")
		match := re.FindStringSubmatch(issue.GetBody())
		if len(match) > 0 {
			return match
		}

		return nil
	}()

	// Get content
	content := issue.GetBody()
	content = strings.Replace(content, "\r", "", -1)

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
	re := regexp.MustCompile(`!\[image*\]\((.*)\)`)
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

	// Write file
	w := bufio.NewWriter(file)
	w.WriteString("---\n")
	w.WriteString("title: " + article.Title + "\n")
	w.WriteString("author: " + article.Author + "\n")
	w.WriteString("date: " + article.Date + "\n")
	w.WriteString("categories:\n  - " + article.Category + "\n")
	w.WriteString("tags:\n")
	for _, tag := range article.Tags {
		w.WriteString("  - " + tag + "\n")
	}
	w.WriteString("draft: " + fmt.Sprintf("%t", article.Draft) + "\n")
	w.WriteString(article.ExtraFrontMatter)
	w.WriteString("---\n")
	w.WriteString(article.Content)
	w.Flush()
}
