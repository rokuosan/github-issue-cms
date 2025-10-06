package converter_v2

import (
	"github.com/yuin/goldmark/ast"
)

func FindImages(node ast.Node, source []byte) []*ast.Image {
	var images []*ast.Image

	ast.Walk(node, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}

		if img, ok := n.(*ast.Image); ok {
			images = append(images, img)
		}

		return ast.WalkContinue, nil
	})

	return images
}
