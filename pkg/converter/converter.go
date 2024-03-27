package converter

import (
	"fmt"
	"github.com/google/go-github/v56/github"
	"log/slog"
	"regexp"
	"strconv"
	"strings"
)

type Converter struct {
	Client *Client
}

func (c *Converter) IssueToArticle(issue *github.Issue) *Article {
	if issue.IsPullRequest() {
		return nil
	}
	num := strconv.Itoa(issue.GetNumber())
	slog.Debug("Converting #" + num + "...")

	// Get issue content
	content := strings.Replace(issue.GetBody(), "\r", "", -1)
	if !strings.HasSuffix(content, "\n") {
		content = content + "\n"
	}

	// Get front matter and remove it from original content
	frontMatter := func() []string {
		re := regexp.MustCompile("(?s)^\\s*```\\n([^`]*)\\n```")
		match := re.FindStringSubmatch(content)
		if len(match) > 0 {
			return match
		}
		return nil
	}()
	if frontMatter != nil {
		content = strings.Replace(content, frontMatter[0], "", 1)
	} else {
		frontMatter = []string{"", ""}
	}
	content = strings.TrimLeft(content, "\n")

	// Replace image URL of Markdown style
	time := issue.GetCreatedAt().Format("2006-01-02_150405")
	re := regexp.MustCompile(`!\[image*]\((.*)\)`)
	match := re.FindAllStringSubmatch(content, -1)
	for i, m := range match {
		replaced := "![" + m[1] + "](images/" + time + "/" + strconv.Itoa(i) + ".png)"
		c.Client.downloadImage(m[1], time, i)
		content = strings.Replace(content, m[0], replaced, -1)
	}

	// Replace image URL of HTML style
	re = regexp.MustCompile(`<img width="\d+" alt="(\w+)" src="(\S+)">`)
	match = re.FindAllStringSubmatch(content, -1)
	for i, m := range match {
		replaced := "![" + m[1] + "](images/" + time + "/" + strconv.Itoa(i) + ".png)"

		fmt.Println("Replace: " + m[2])
		c.Client.downloadImage(m[2], time, i)

		content = strings.Replace(content, m[0], replaced, -1)
	}

	// Tags
	var tags []string
	for _, label := range issue.Labels {
		tags = append(tags, label.GetName())
	}

	return &Article{
		Author:           issue.GetUser().GetLogin(),
		Title:            issue.GetTitle(),
		Date:             issue.GetCreatedAt().Format("2006-01-02T15:04:05Z"),
		Category:         issue.GetMilestone().GetTitle(),
		Draft:            issue.GetState() != "closed",
		Content:          content,
		ExtraFrontMatter: frontMatter[1],
		Tags:             tags,
	}
}
