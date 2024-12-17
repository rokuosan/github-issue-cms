package convert

import (
	"github.com/rokuosan/github-issue-cms/internal/api"
	"gopkg.in/yaml.v3"
	"regexp"
	"strings"
)

type issue struct {
	*api.GitHubIssue
}

var regex = struct {
	FrontMatter *regexp.Regexp
}{
	FrontMatter: regexp.MustCompile("(?s)^\\s*```\\n([^`]*)\\n```"),
}

// ConvertToArticle converts the issue to an article
func (i *issue) ConvertToArticle() (*Article, error) {
	frontMatter, err := i.frontMatter()
	if err != nil {
		return nil, err
	}

	article := &Article{
		Title:      i.GetTitle(),
		Content:    i.content(),
		Author:     i.User.GetLogin(),
		Date:       i.CreatedAt.Format("2006-01-02T15:04:05Z"),
		Categories: []string{i.Milestone.GetTitle()},
		Tags:       i.tags(),
		Draft:      i.GetState() == "open",
	}

	if strings.TrimSpace(frontMatter) == "" {
		return article, nil
	}

	extra := make(map[string]interface{})
	if err = yaml.Unmarshal([]byte(frontMatter), extra); err != nil {
		return nil, err
	}

	// Override the article fields with the front matter
	for key, value := range extra {
		isExtra := false
		switch key {
		case "title":
			article.Title = value.(string)
		case "author":
			article.Author = value.(string)
		case "authors":
			article.Authors = value.([]string)
		case "date":
			article.Date = value.(string)
		case "categories":
			article.Categories = []string{value.(string)}
		case "tags":
			article.Tags = value.([]string)
		case "draft":
			article.Draft = value.(bool)
		default:
			isExtra = true
		}
		if !isExtra {
			delete(extra, key)
		}
	}

	if len(extra) != 0 {
		e, err := yaml.Marshal(extra)
		if err != nil {
			return nil, err
		}
		article.ExtraFrontMatter = string(e)
	}

	return article, nil
}

// body returns the body of the issue
func (i *issue) body() string {
	c := strings.Replace(i.Issue.GetBody(), "\r", "", -1)
	c = strings.TrimSpace(c)
	if !strings.HasSuffix(c, "\n") {
		c = c + "\n"
	}
	return c
}

// content returns the content of the issue without front matter
func (i *issue) content() string {
	content := i.body()
	frontMatter := regex.FrontMatter.FindString(content)
	if frontMatter != "" {
		content = strings.Replace(content, frontMatter, "", 1)
		content = strings.TrimLeft(content, "\n")
	}
	return content
}

// frontMatter returns the front matter of the issue
func (i *issue) frontMatter() (string, error) {
	match := regex.FrontMatter.FindStringSubmatch(i.body())

	if len(match) < 2 {
		return "", nil
	}

	content := []byte(match[1])

	// Verify that the front matter is valid YAML
	t := make(map[interface{}]interface{})
	err := yaml.Unmarshal(content, &t)
	if err != nil {
		return "", err
	}

	return string(content), nil
}

// tags returns the tags of the issue
func (i *issue) tags() []string {
	tags := make([]string, len(i.Labels))
	for i, l := range i.Labels {
		tags[i] = l.GetName()
	}
	return tags
}
