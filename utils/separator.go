package utils

import (
	"gioui.org/f32"
	"gioui.org/layout"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/unit"
	"image"
)

func MakeVerticalSeparator(thickness unit.Dp, width unit.Dp) layout.Widget {
	return func(gtx layout.Context) layout.Dimensions {
		width := gtx.Dp(width)
		spacing := gtx.Dp(CommonSpacing)

		var path clip.Path
		path.Begin(gtx.Ops)

		path.MoveTo(f32.Point{
			X: float32(spacing),
		})
		path.LineTo(f32.Point{
			X: float32(width - spacing),
		})

		defer clip.Stroke{Path: path.End(), Width: float32(gtx.Dp(thickness))}.Op().Push(gtx.Ops).Pop()
		paint.Fill(gtx.Ops, LessContrastBg)

		return layout.Dimensions{
			Size: image.Point{
				X: width + spacing*2,
				Y: spacing,
			},
		}
	}
}
