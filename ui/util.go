package ui

import (
	"PGCombatTracker/abstract"
	"gioui.org/io/system"
	"gioui.org/layout"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/text"
	"gioui.org/unit"
	"gioui.org/widget/material"
	"image"
	"image/color"
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

func MakeColoredBG(nrgba color.NRGBA) layout.Widget {
	return func(gtx layout.Context) layout.Dimensions {
		defer clip.Rect{Max: gtx.Constraints.Min}.Push(gtx.Ops).Pop()
		paint.Fill(gtx.Ops, nrgba)
		return layout.Dimensions{Size: gtx.Constraints.Min}
	}
}

func MakeColoredAndDragBG(nrgba color.NRGBA) layout.Widget {
	return func(gtx layout.Context) layout.Dimensions {
		defer clip.Rect{Max: gtx.Constraints.Min}.Push(gtx.Ops).Pop()
		system.ActionInputOp(system.ActionMove).Add(gtx.Ops)
		paint.Fill(gtx.Ops, nrgba)
		return layout.Dimensions{Size: gtx.Constraints.Min}
	}
}

func MakeColoredAndOptionalDragBG(nrgba color.NRGBA, draggable bool) layout.Widget {
	return func(gtx layout.Context) layout.Dimensions {
		defer clip.Rect{Max: gtx.Constraints.Min}.Push(gtx.Ops).Pop()

		if draggable {
			system.ActionInputOp(system.ActionMove).Add(gtx.Ops)
		}

		paint.Fill(gtx.Ops, nrgba)
		return layout.Dimensions{Size: gtx.Constraints.Min}
	}
}

func MakeRoundedBG(radius int, nrgba color.NRGBA) layout.Widget {
	return func(gtx layout.Context) layout.Dimensions {
		defer clip.UniformRRect(
			image.Rectangle{Max: gtx.Constraints.Min},
			radius,
		).Push(gtx.Ops).Pop()
		paint.Fill(gtx.Ops, nrgba)
		return layout.Dimensions{Size: gtx.Constraints.Min}
	}
}

func WithColor(style material.LabelStyle, c color.NRGBA) material.LabelStyle {
	style.Color = c
	return style
}

func WithAlignment(style material.LabelStyle, alignment text.Alignment) material.LabelStyle {
	style.Alignment = alignment
	return style
}

func FlexSpacerW(width unit.Dp) layout.FlexChild {
	return layout.Rigid(layout.Spacer{Width: width}.Layout)
}

func FlexSpacerH(height unit.Dp) layout.FlexChild {
	return layout.Rigid(layout.Spacer{Height: height}.Layout)
}

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
		return abstract.LessContrastBg
	} else {
		return abstract.SecondBG
	}
}
