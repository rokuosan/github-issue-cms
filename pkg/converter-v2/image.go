package converter_v2

import (
	"github.com/yuin/goldmark/ast"
)

type Image interface {
	Destination() string
	Title() string
	AST() *ast.Image
}

type image struct {
	*ast.Image
}

func (i *image) Destination() string {
	return string(i.Image.Destination)
}

func (i *image) Title() string {
	return string(i.Image.Title)
}

func (i *image) AST() *ast.Image {
	return i.Image
}

func FindImages(node ast.Node, source []byte) []Image {
	var images []Image

	ast.Walk(node, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}

		if img, ok := n.(*ast.Image); ok {
			images = append(images, &image{Image: img})
		}

		return ast.WalkContinue, nil
	})

	return images
}
