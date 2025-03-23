package replacer

import (
	"testing"
)

func TestHTMLReplacer_Replace(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "basic image tag with alt",
			input:    `<img src="https://example.com/image.jpg" alt="example">`,
			expected: `![example](https://example.com/image.jpg)`,
		},
		{
			name:     "image tag with single quotes",
			input:    `<img src='https://example.com/image.jpg' alt='example'>`,
			expected: `![example](https://example.com/image.jpg)`,
		},
		{
			name:     "image tag with no alt",
			input:    `<img src="https://example.com/image.jpg">`,
			expected: `![image](https://example.com/image.jpg)`,
		},
		{
			name:     "image tag with other attributes",
			input:    `<img width="100" src="https://example.com/image.jpg" class="img" alt="example">`,
			expected: `![example](https://example.com/image.jpg)`,
		},
		{
			name:     "multiple image tags",
			input:    `<p><img src="https://example.com/image1.jpg" alt="first"> and <img src="https://example.com/image2.jpg" alt="second"></p>`,
			expected: `<p>![first](https://example.com/image1.jpg) and ![second](https://example.com/image2.jpg)</p>`,
		},
		{
			name:     "image tag with no src",
			input:    `<img alt="example">`,
			expected: `<img alt="example">`,
		},
		{
			name:     "non-image HTML",
			input:    `<p>This is <strong>HTML</strong> content.</p>`,
			expected: `<p>This is <strong>HTML</strong> content.</p>`,
		},
		{
			name:     "mixed content",
			input:    `<p>Text with <img src="https://example.com/image.jpg" alt="example"> embedded.</p>`,
			expected: `<p>Text with ![example](https://example.com/image.jpg) embedded.</p>`,
		},
	}

	replacer := NewHTMLReplacer()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := replacer.Replace(tt.input)
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if result != tt.expected {
				t.Errorf("got %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestHTMLReplacer_ReplaceFunc(t *testing.T) {
	replacer := NewHTMLReplacer().(*htmlReplacer)

	tests := []struct {
		name     string
		imgTag   string
		expected string
	}{
		{
			name:     "basic image tag",
			imgTag:   `<img src="https://example.com/image.jpg" alt="example">`,
			expected: `![example](https://example.com/image.jpg)`,
		},
		{
			name:     "no alt attribute",
			imgTag:   `<img src="https://example.com/image.jpg">`,
			expected: `![image](https://example.com/image.jpg)`,
		},
		{
			name:     "no src attribute",
			imgTag:   `<img alt="example">`,
			expected: `<img alt="example">`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := replacer.replaceFunc(tt.imgTag)
			if result != tt.expected {
				t.Errorf("got %q, want %q", result, tt.expected)
			}
		})
	}
}
