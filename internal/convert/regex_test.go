package convert

import "testing"

func TestRegex_FrontMatter(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
		test     map[string]func(t *testing.T)
	}{
		{
			name: "先頭にあるフロントマターを取れる",
			input: "```\nsample front matter\n```\n" +
				"sample content",
			expected: "```\nsample front matter\n```",
			test: map[string]func(t *testing.T){
				"中身が取れる": func(t *testing.T) {
					got := regex.FrontMatter.FindStringSubmatch(
						"```\nsample front matter\n```\n" +
							"sample content",
					)
					if got[1] != "sample front matter" {
						t.Errorf("regex.FrontMatter.FindStringSubmatch()[1] = %#v, want %#v", got[1], "sample front matter")
					}
				},
			},
		},
		{
			name: "先頭以外にあるフロントマターは取れない",
			input: "sample content\n" +
				"```\nsample front matter\n```\n" +
				"sample content",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := regex.FrontMatter.FindString(tt.input); got != tt.expected {
				t.Errorf("regex.FrontMatter.FindString() = %#v, want %#v", got, tt.expected)
			}
		})
		for name, test := range tt.test {
			t.Run(name, test)
		}
	}
}
