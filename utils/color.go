package utils

import "image/color"

func AverageColor(colors ...color.NRGBA) color.NRGBA {
	var r, g, b, a int

	for _, c := range colors {
		r += int(c.R)
		g += int(c.G)
		b += int(c.B)
		a += int(c.A)
	}

	colorsLen := len(colors)

	r /= colorsLen
	g /= colorsLen
	b /= colorsLen
	a /= colorsLen

	return color.NRGBA{
		R: uint8(r),
		G: uint8(g),
		B: uint8(b),
		A: uint8(a),
	}
}

func RGB(c uint32) color.NRGBA {
	return ARGB(0xff000000 | c)
}

func ARGB(c uint32) color.NRGBA {
	return color.NRGBA{A: uint8(c >> 24), R: uint8(c >> 16), G: uint8(c >> 8), B: uint8(c)}
}
