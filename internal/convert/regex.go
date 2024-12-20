package convert

import "regexp"

var regex = struct {
	FrontMatter  *regexp.Regexp
	MarkdownLink *regexp.Regexp
}{
	FrontMatter:  regexp.MustCompile("(?s)^\\s*```\\n([^`]*)\\n```"),
	MarkdownLink: regexp.MustCompile(`\[(^[\n\r]*)]\((^[\n\r]*)\)`),
}
