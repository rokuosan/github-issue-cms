package converter_v2

import (
	"fmt"

	"github.com/yuin/goldmark/ast"
	"gopkg.in/yaml.v3"
)

type FrontMatter string

func ExtractFrontMatter(node ast.Node, source []byte) (*FrontMatter, error) {
	if node.ChildCount() == 0 {
		return nil, nil
	}

	firstChild := node.FirstChild()
	if fcb, ok := firstChild.(*ast.FencedCodeBlock); ok {
		lines := fcb.Lines()
		if lines.Len() == 0 {
			return nil, nil
		}

		var content []byte
		for i := range lines.Len() {
			line := lines.At(i)
			content = append(content, line.Value(source)...)
		}

		fm := FrontMatter(content)
		return &fm, nil
	}

	return nil, nil
}

func (f *FrontMatter) String() string {
	if f == nil {
		return ""
	}
	return string(*f)
}

func (f *FrontMatter) StringWithBackQuotes() string {
	if f == nil {
		return "```\n```"
	}
	return fmt.Sprintf("```\n%s\n```", string(*f))
}

func (f *FrontMatter) ParseYAML() map[string]any {
	var result map[string]any
	if f != nil {
		_ = yaml.Unmarshal([]byte(*f), &result)
	}
	return result
}
