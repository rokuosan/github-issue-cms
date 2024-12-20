package convert

import "regexp"

var regex = struct {
	FrontMatter  *regexp.Regexp
	MarkdownLink *regexp.Regexp
	HTMLImage    *regexp.Regexp
}{
	FrontMatter:  regexp.MustCompile("(?s)^\\s*```\\n([^`]*)\\n```"),
	MarkdownLink: regexp.MustCompile(`\[(^[\n\r]*)]\((^[\n\r]*)\)`),
	HTMLImage:    regexp.MustCompile(`<img[^>]*\b(?:alt="([^"]*)"[^>]*\bsrc="([^"]*)"|src="([^"]*)"[^>]*\balt="([^"]*)")[^>]*>`),
}
