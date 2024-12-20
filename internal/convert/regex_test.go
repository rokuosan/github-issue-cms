package convert

import "testing"

func assertFindStringSubMatch(t *testing.T, got, want []string) {
	t.Helper()
	if len(got) != len(want) {
		t.Errorf("len(got) = %d, want %d", len(got), len(want))
	}
	for i := range got {
		if got[i] != want[i] {
			t.Errorf("got[%d] = %#v, want %#v", i, got[i], want[i])
		}
	}
}

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

func TestRegex_MarkdownLink(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{
			name:     "マークダウンリンクを取れる",
			input:    "sample [link](https://example.com) content",
			expected: []string{"[link](https://example.com)", "link", "https://example.com"},
		},
		{
			name:     "マークダウンリンクがない",
			input:    "sample content",
			expected: nil,
		},
		{
			name: "途中で改行がある場合はマークダウンリンクではない",
			input: "sample [link](https://example\n.com)\n" +
				"sample content",
			expected: nil,
		},
		{
			name:     "無効なURLが含まれたマークダウンリンク",
			input:    "[link](https://example.com[])\n",
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := regex.MarkdownLink.FindStringSubmatch(tt.input); got != nil {
				assertFindStringSubMatch(t, got, tt.expected)
			}
		})
	}
}
