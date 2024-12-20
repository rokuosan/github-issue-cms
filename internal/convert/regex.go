package convert

import "regexp"

var regex = struct {
	FrontMatter             *regexp.Regexp
	MarkdownLink            *regexp.Regexp
	MarkdownCodeBlock       *regexp.Regexp
	MarkdownInlineCodeBlock *regexp.Regexp
	HTMLImage               *regexp.Regexp
}{
	FrontMatter:             regexp.MustCompile("(?s)^\\s*```\\n([^`]*)\\n```"),
	MarkdownLink:            regexp.MustCompile(`\[(^[\n\r]*)]\((^[\n\r]*)\)`),
	MarkdownCodeBlock:       regexp.MustCompile("```[\\s\\S]*?```"),
	MarkdownInlineCodeBlock: regexp.MustCompile("`{1,2}[^`]*`{1,2}"),
	HTMLImage:               regexp.MustCompile(`<img[^>]*\b(?:alt="([^"]*)"[^>]*\bsrc="([^"]*)"|src="([^"]*)"[^>]*\balt="([^"]*)")[^>]*>`),
}
