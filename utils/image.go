package utils

import (
	"image"
	"image/color"
)

type CheckerImage struct {
	Size image.Point
}

func (c CheckerImage) ColorModel() color.Model {
	return color.NRGBAModel
}

func (c CheckerImage) Bounds() image.Rectangle {
	return image.Rectangle{
		Max: c.Size,
	}
}

func (c CheckerImage) At(x, y int) color.Color {
	x, y = x/2, y/2
	if (x+y)%2 == 1 {
		return LessContrastBg
	} else {
		return SecondBG
	}
}
