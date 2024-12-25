package utils

import (
	"gioui.org/layout"
	"gioui.org/unit"
	"image"
)

func LayoutMinimalSize(size image.Point, dim layout.Dimensions) layout.Dimensions {
	return layout.Dimensions{
		Size: image.Point{
			X: max(dim.Size.X, size.X),
			Y: max(dim.Size.Y, size.Y),
		},
		Baseline: dim.Baseline,
	}
}

func LayoutMinimalX(width int, dim layout.Dimensions) layout.Dimensions {
	return LayoutMinimalSize(
		image.Point{
			X: width,
			Y: 0,
		},
		dim,
	)
}

func LayoutMinimalY(height int, dim layout.Dimensions) layout.Dimensions {
	return LayoutMinimalSize(
		image.Point{
			X: 0,
			Y: height,
		},
		dim,
	)
}

func LayoutDefinedHeight(gtx layout.Context, height int, widget layout.Widget) layout.Dimensions {
	cgtx := gtx
	cgtx.Constraints.Min = image.Point{}
	cgtx.Constraints.Max.Y = height

	dims := widget(cgtx)

	return LayoutMinimalSize(image.Point{
		X: gtx.Constraints.Max.X,
		Y: height,
	}, dims)
}

func FlexSpacerW(width unit.Dp) layout.FlexChild {
	return layout.Rigid(layout.Spacer{Width: width}.Layout)
}

func FlexSpacerH(height unit.Dp) layout.FlexChild {
	return layout.Rigid(layout.Spacer{Height: height}.Layout)
}
