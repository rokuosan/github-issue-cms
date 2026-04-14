package core

import (
	"fmt"

	"gopkg.in/yaml.v3"
)

// ArticleRenderer renders an article into a serialized representation.
type ArticleRenderer interface {
	Render(article *Article) (string, error)
}

// HugoArticleRenderer renders articles as Hugo-compatible markdown.
type HugoArticleRenderer struct{}

// NewHugoArticleRenderer creates a HugoArticleRenderer.
func NewHugoArticleRenderer() ArticleRenderer {
	return HugoArticleRenderer{}
}

// Render renders an article as Hugo-compatible markdown.
func (HugoArticleRenderer) Render(article *Article) (string, error) {
	extra := article.FrontMatter.Values()
	rendered := article.Clone()
	applyFrontMatterOverrides(rendered, extra)

	extraFrontMatter, err := NewFrontMatter(extra).MarshalYAML()
	if err != nil {
		return "", err
	}

	partial, err := yaml.Marshal(rendered)
	if err != nil {
		return "", err
	}

	frontMatter := string(partial) + string(extraFrontMatter)
	return fmt.Sprintf("---\n%s---\n\n%s\n", frontMatter, rendered.Content), nil
}

func applyFrontMatterOverrides(article *Article, extra map[string]any) {
	if author, ok := stringValue(extra["author"]); ok {
		article.Author = author
		delete(extra, "author")
	}
	if title, ok := stringValue(extra["title"]); ok {
		article.Title = title
		delete(extra, "title")
	}
	if date, ok := stringValue(extra["date"]); ok {
		article.Date = date
		delete(extra, "date")
	}
	if category, ok := categoryValue(extra["categories"]); ok {
		article.Category = category
		delete(extra, "categories")
	}
	if tags, ok := stringSliceValue(extra["tags"]); ok {
		article.Tags = tags
		delete(extra, "tags")
	}
	if draft, ok := boolValue(extra["draft"]); ok {
		article.Draft = draft
		delete(extra, "draft")
	}
}

func stringValue(value any) (string, bool) {
	s, ok := value.(string)
	return s, ok
}

func boolValue(value any) (bool, bool) {
	b, ok := value.(bool)
	return b, ok
}

func categoryValue(value any) (string, bool) {
	if s, ok := value.(string); ok {
		return s, true
	}
	values, ok := stringSliceValue(value)
	if !ok || len(values) == 0 {
		return "", false
	}
	return values[0], true
}

func stringSliceValue(value any) ([]string, bool) {
	switch typed := value.(type) {
	case []string:
		return append([]string(nil), typed...), true
	case []any:
		values := make([]string, 0, len(typed))
		for _, item := range typed {
			s, ok := item.(string)
			if !ok {
				return nil, false
			}
			values = append(values, s)
		}
		return values, true
	default:
		return nil, false
	}
}
