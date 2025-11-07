package converter_v2

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/text"
)

func stringPtr(s string) *string {
	return &s
}

func TestExtractFrontMatter(t *testing.T) {
	tests := []struct {
		name     string
		markdown string
		expected *string
		error    error
	}{
		{
			name:     "with front matter",
			markdown: "```yaml\ntitle: Test Article\nauthor: John Doe\ndate: 2024-01-01\n```\n\n# Content\n\nThis is the content.",
			expected: stringPtr("title: Test Article\nauthor: John Doe\ndate: 2024-01-01\n"),
			error:    nil,
		},
		{
			name:     "with front matter without language",
			markdown: "```\ntitle: Another Article\ndate: 2024-02-02\n```\n\n# Content\n\nThis is another content.",
			expected: stringPtr("title: Another Article\ndate: 2024-02-02\n"),
			error:    nil,
		},
		{
			name:     "without front matter",
			markdown: "# Content\n\nThis is the content without front matter.",
			expected: nil,
			error:    nil,
		},
		{
			name:     "empty front matter",
			markdown: "```\n```\n\n# Content\n\nThis is the content with empty front matter.",
			expected: nil,
			error:    nil,
		},
		{
			name:     "non-fenced code block at start",
			markdown: "> This is a blockquote.\n\n```\ntitle: Test Article\n```\n\n# Content",
			expected: nil,
			error:    nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			md := goldmark.New(
				goldmark.WithParserOptions(
					parser.WithAutoHeadingID(),
				),
			)

			source := []byte(tt.markdown)
			doc := md.Parser().Parse(text.NewReader(source))

			got, err := ExtractFrontMatter(doc, source)
			if tt.error != nil {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				if tt.expected == nil {
					assert.Nil(t, got)
				} else {
					assert.NotNil(t, got)
					assert.Equal(t, *tt.expected, got.String())
				}
			}
		})
	}
}
