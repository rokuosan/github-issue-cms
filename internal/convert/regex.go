package convert

import "regexp"

var regex = struct {
	FrontMatter *regexp.Regexp
}{
	FrontMatter: regexp.MustCompile("(?s)^\\s*```\\n([^`]*)\\n```"),
}
