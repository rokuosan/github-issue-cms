package converter_v2

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
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

func TestImage_Download(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/test-image.png", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("image data"))
	})
	server := httptest.NewServer(mux)
	defer server.Close()

	t.Run("download image", func(t *testing.T) {
		md := goldmark.New()
		markdown := "![alt text](" + server.URL + "/test-image.png)"
		source := []byte(markdown)
		doc := md.Parser().Parse(text.NewReader(source))

		images := FindImages(doc, source)
		assert.Equal(t, 1, len(images))

		img := images[0]
		var buf []byte
		writer := bytes.NewBuffer(buf)
		ctx := context.Background()

		err := img.Download(ctx, server.Client(), writer)
		assert.NoError(t, err)
		assert.Equal(t, "image data", writer.String())
	})

	t.Run("Download to temporary file", func(t *testing.T) {
		md := goldmark.New()
		markdown := "![alt text](" + server.URL + "/test-image.png)"
		source := []byte(markdown)
		doc := md.Parser().Parse(text.NewReader(source))

		images := FindImages(doc, source)
		assert.Equal(t, 1, len(images))

		img := images[0]

		filename := filepath.Join(t.TempDir(), "downloaded_image.png")
		fmt.Println(filename)
		file, err := os.Create(filename)
		assert.NoError(t, err)
		defer file.Close()

		ctx := context.Background()
		err = img.Download(ctx, server.Client(), file)
		assert.NoError(t, err)

		data, err := os.ReadFile(filename)
		assert.NoError(t, err)
		assert.Equal(t, "image data", string(data))
	})
}
