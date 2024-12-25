package utils

import (
	"gioui.org/io/system"
	"gioui.org/layout"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"image"
	"image/color"
)

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
