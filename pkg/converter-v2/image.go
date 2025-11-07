package converter_v2

import (
	"context"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/yuin/goldmark/ast"
)

type Image interface {
	Destination() string
	Title() string
	AST() *ast.Image
	URL() (*url.URL, error)
	Download(context.Context, *http.Client, io.Writer) error
}

type image struct {
	image *ast.Image
	url   string
}

var _ Image = (*image)(nil)

func (i *image) Destination() string {
	return string(i.image.Destination)
}

func (i *image) Title() string {
	return string(i.image.Title)
}

func (i *image) AST() *ast.Image {
	return i.image
}

func (i *image) URL() (*url.URL, error) {
	return url.Parse(i.url)
}

func (i *image) Download(ctx context.Context, client *http.Client, w io.Writer) error {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	u, err := i.URL()
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return err
	}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	_, err = io.Copy(w, resp.Body)
	if err != nil {
		return err
	}
	return nil
}

func FindImages(node ast.Node, source []byte) []Image {
	var images []Image

	ast.Walk(node, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}

		if img, ok := n.(*ast.Image); ok {
			images = append(images, &image{
				image: img,
				url:   string(img.Destination),
			})
		}

		return ast.WalkContinue, nil
	})

	return images
}
