package core

import (
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

// ParseArticleFromMarkdown reads a Hugo-compatible markdown file (with YAML
// frontmatter delimited by "---") and returns an Article.
func ParseArticleFromMarkdown(path string) (*Article, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read markdown file %s: %w", path, err)
	}

	article, err := parseArticleContent(string(data))
	if err != nil {
		return nil, fmt.Errorf("parse markdown %s: %w", path, err)
	}

	return article, nil
}

func parseArticleContent(content string) (*Article, error) {
	trimmed := strings.TrimPrefix(content, "---\n")
	end := strings.Index(trimmed, "\n---\n")
	if end < 0 {
		return nil, fmt.Errorf("invalid markdown: missing closing frontmatter delimiter")
	}

	fmRaw := trimmed[:end]
	body := trimmed[end+5:] // skip "\n---\n"

	article := &Article{
		Content: body,
	}

	if err := yaml.Unmarshal([]byte(fmRaw), article); err != nil {
		return nil, fmt.Errorf("parse frontmatter: %w", err)
	}

	return article, nil
}
