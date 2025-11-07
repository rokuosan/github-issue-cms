package converter_v2

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/text"
)

func TestFindImages(t *testing.T) {
	tests := []struct {
		name     string
		markdown string
		want     int
	}{
		{
			name:     "no images",
			markdown: "# Hello\n\nThis is text without images.",
			want:     0,
		},
		{
			name:     "single image",
			markdown: "# Title\n\n![alt text](https://example.com/image.png)",
			want:     1,
		},
		{
			name:     "multiple images",
			markdown: "![first](image1.png)\n\nSome text\n\n![second](image2.jpg)\n\n![third](https://example.com/image3.gif)",
			want:     3,
		},
		{
			name:     "images with links",
			markdown: "[link](https://example.com)\n\n![image](test.png)\n\n[another link](page.html)",
			want:     1,
		},
		{
			name:     "nested markdown",
			markdown: "# Header\n\n- List item 1\n- ![nested image](nested.png)\n- List item 3\n\n> ![quoted image](quoted.jpg)",
			want:     2,
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

			images := FindImages(doc, source)
			assert.Equal(t, tt.want, len(images))
		})
	}
}

func TestFindImages_Properties(t *testing.T) {
	markdown := "![alt text](https://example.com/image.png \"title text\")"
	source := []byte(markdown)

	md := goldmark.New()
	doc := md.Parser().Parse(text.NewReader(source))

	images := FindImages(doc, source)
	assert.Equal(t, 1, len(images))

	img := images[0]
	assert.Equal(t, "https://example.com/image.png", string(img.Destination()))
	assert.Equal(t, "title text", string(img.Title()))
}
