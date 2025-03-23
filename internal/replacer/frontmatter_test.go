package replacer

import (
	"testing"
)

func TestFrontMatterReplacer_Replace(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "with front matter",
			input:    "```\n---\ntitle: Test\n---\n```\n\nThis is content.",
			expected: "This is content.",
		},
		{
			name:     "with front matter and extra newlines",
			input:    "```\ntitle: Test\nauthor: John\n```\n\n\nThis is content.",
			expected: "This is content.",
		},
		{
			name:     "without front matter",
			input:    "This is content without front matter.",
			expected: "This is content without front matter.",
		},
		{
			name:     "with empty front matter",
			input:    "```\n\n```\nContent after empty front matter.",
			expected: "Content after empty front matter.",
		},
		{
			name:     "with front matter with indentation",
			input:    "  ```\ntitle: Test\n```\nContent with indented front matter.",
			expected: "Content with indented front matter.",
		},
	}

	replacer := NewFrontMatterReplacer()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := replacer.Replace(tt.input)
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if result != tt.expected {
				t.Errorf("expected: %q, got: %q", tt.expected, result)
			}
		})
	}
}
