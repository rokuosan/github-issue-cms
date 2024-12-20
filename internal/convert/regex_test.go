package convert

import (
	"reflect"
	"testing"
)

func assertFindStringSubMatch(t *testing.T, got, want []string) {
	t.Helper()
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %#v, want %#v", got, want)
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
			name:     "リンクを取れる",
			input:    "sample [link](https://example.com) content",
			expected: []string{"[link](https://example.com)", "link", "https://example.com"},
		},
		{
			name:     "リンクがない",
			input:    "sample content",
			expected: nil,
		},
		{
			name: "途中で改行がある場合はリンクではない",
			input: "sample [link](https://example\n.com)\n" +
				"sample content",
			expected: nil,
		},
		{
			name:     "無効なURLが含まれたリンク",
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

func TestRegex_HTMLImage(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{
			name:     "HTMLのimage タグを取れる",
			input:    `<img alt="sample" src="https://example.com">`,
			expected: []string{`<img alt="sample" src="https://example.com">`, "sample", "https://example.com", "", ""},
		},
		{
			name:     "HTMLのimage タグのaltとsrcが逆でも取れる",
			input:    `<img alt="sample" src="https://example.com">`,
			expected: []string{`<img alt="sample" src="https://example.com">`, "sample", "https://example.com", "", ""},
		},
		{
			name:     "HTMLのimageがないときはヒットしない",
			input:    "sample content",
			expected: nil,
		},
		{
			name:     "途中で改行があっても、HTMLのimageタグとしてみる",
			input:    "<img alt=\"sample\" \n src=\"https://example.com\">",
			expected: []string{"<img alt=\"sample\" \n src=\"https://example.com\">", "sample", "https://example.com", "", ""},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := regex.HTMLImage.FindStringSubmatch(tt.input); got != nil {
				assertFindStringSubMatch(t, got, tt.expected)
			}
		})
	}
}

func TestRegex_MarkdownCodeBlock(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name: "コードブロックを取れる",
			input: "sample\n" +
				"```\n" +
				"sample code\n" +
				"```\n" +
				"sample content",
			expected: "```\nsample code\n```",
		},
		{
			name:     "コードブロックがない",
			input:    "sample content",
			expected: "",
		},
		{
			name: "タグがついていてもコードブロックを取れる",
			input: "sample\n" +
				"```type\n" +
				"sample code\n" +
				"```\n" +
				"sample content",
			expected: "```type\nsample code\n```",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := regex.MarkdownCodeBlock.FindString(tt.input); got != tt.expected {
				t.Errorf("regex.MarkdownCodeBlock.FindString() = %#v, want %#v", got, tt.expected)
			}
		})
	}
}

func TestRegex_MarkdownInlineCodeBlock(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name: "バッククオートが1つのインラインコードブロックを取れる",
			input: "sample\n" +
				"`sample code`\n" +
				"sample content",
			expected: "`sample code`",
		},
		{
			name: "バッククオートが2つのインラインコードブロックを取れる",
			input: "sample\n" +
				"``sample code``\n" +
				"sample content",
			expected: "``sample code``",
		},
		{
			name:     "インラインコードブロックがない",
			input:    "sample content",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := regex.MarkdownInlineCodeBlock.FindString(tt.input); got != tt.expected {
				t.Errorf("regex.MarkdownInlineCodeBlock.FindString() = %#v, want %#v", got, tt.expected)
			}
		})
	}
}
