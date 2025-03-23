package replacer

import (
	"fmt"
	"regexp"
)

type htmlReplacer struct {
	imageTagRegex *regexp.Regexp
	altAttrRegex  *regexp.Regexp
}

func NewHTMLReplacer() Replacer {
	return &htmlReplacer{
		imageTagRegex: regexp.MustCompile(`<img[^>]*src=["']([^"']+)["'][^>]*>`),
		altAttrRegex:  regexp.MustCompile(`alt=["']([^"']+)["']`),
	}
}

var _ Replacer = &htmlReplacer{}

func (h *htmlReplacer) Replace(text string) (string, error) {
	replaced := h.imageTagRegex.ReplaceAllStringFunc(text, h.replaceFunc)
	return replaced, nil
}

func (h *htmlReplacer) replaceFunc(imgTag string) string {
	srcMatches := h.imageTagRegex.FindStringSubmatch(imgTag)
	altMatches := h.altAttrRegex.FindStringSubmatch(imgTag)
	if len(srcMatches) < 2 {
		return imgTag // src がない場合は変換せずそのまま返す
	}
	altText := "image"
	if len(altMatches) > 1 {
		altText = altMatches[1]
	}
	return fmt.Sprintf("![%s](%s)", altText, srcMatches[1])
}
