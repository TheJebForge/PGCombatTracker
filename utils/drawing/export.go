package drawing

import (
	"gioui.org/widget/material"
	"github.com/fogleman/gg"
	"image"
	"log"
)

func ExportImage(theme *material.Theme, baseWidget Widget, max F64Point) image.Image {
	ltx := NewContext(max)

	result := baseWidget(ltx)

	imageSize := result.Size.Round()
	log.Println(imageSize)
	dc := gg.NewContext(imageSize.X, imageSize.Y)
	dc.SetColor(theme.Bg)
	dc.Clear()

	result.Draw(dc)

	return dc.Image()
}
