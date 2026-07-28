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
	// Normalize line endings: Windows CRLF → LF.
	content = strings.ReplaceAll(content, "\r\n", "\n")

	// Ensure the content starts with the opening delimiter.
	if !strings.HasPrefix(content, "---\n") && !strings.HasPrefix(content, "---\r") {
		return nil, fmt.Errorf("invalid markdown: missing opening frontmatter delimiter")
	}

	// Strip the opening "---\n".
	trimmed := strings.TrimPrefix(content, "---\n")

	// Find the closing delimiter. Try both "\n---\n" and "\n---" (EOF without trailing \n).
	end := strings.Index(trimmed, "\n---\n")
	if end < 0 {
		end = strings.Index(trimmed, "\n---")
	}
	if end < 0 {
		return nil, fmt.Errorf("invalid markdown: missing closing frontmatter delimiter")
	}

	fmRaw := trimmed[:end]

	// Calculate body start: skip past the closing delimiter.
	bodyStart := end
	if strings.HasPrefix(trimmed[bodyStart:], "\n---\n") {
		bodyStart += 5 // len("\n---\n")
	} else if strings.HasPrefix(trimmed[bodyStart:], "\n---") {
		bodyStart += 4 // len("\n---")
	}
	body := trimmed[bodyStart:]

	article := &Article{
		Content: body,
	}

	if err := yaml.Unmarshal([]byte(fmRaw), article); err != nil {
		return nil, fmt.Errorf("parse frontmatter: %w", err)
	}

	return article, nil
}
