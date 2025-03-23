package replacer

import (
	"regexp"
	"strings"
)

type yamlFrontMatterReplacer struct {
	frontMatterRegex *regexp.Regexp
	frontMatter      string
}

func NewFrontMatterReplacer() Replacer {
	return &yamlFrontMatterReplacer{
		frontMatterRegex: regexp.MustCompile("(?s)^\\s*```\\n([^`]*)\\n```"),
	}
}

var _ Replacer = &yamlFrontMatterReplacer{}

func (f *yamlFrontMatterReplacer) Replace(text string) (string, error) {
	frontMatter := f.frontMatterRegex.FindString(text)
	if frontMatter != "" {
		// ヒットする一つだけを削除して、前の空白を削除する
		text = strings.Replace(text, frontMatter, "", 1)
		text = strings.TrimLeft(text, "\n")
	}
	return text, nil
}
