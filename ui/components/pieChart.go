package components

import (
	"PGCombatTracker/utils"
	"fmt"
	"gioui.org/f32"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/unit"
	"gioui.org/widget/material"
	"image"
	"image/color"
	"math"
)

var TAU = math.Pi * 2

type PieChartItem struct {
	Value   int
	Name    string
	SubText string
}

func StylePieChart(theme *material.Theme) PieChart {
	return PieChart{
		MinHeight:     200,
		OverflowLimit: 15,

		ColorBoxSize: utils.CommonSpacing * 2,
		TextSize:     theme.TextSize,
		TextColor:    theme.Fg,
		SubTextColor: utils.GrayText,

		theme: theme,
	}
}

type PieChart struct {
	MinHeight     unit.Dp
	OverflowLimit int

	ColorBoxSize unit.Dp
	TextSize     unit.Sp
	TextColor    color.NRGBA
	SubTextColor color.NRGBA

	theme *material.Theme
}

func (pc PieChart) legend(
	gtx layout.Context,
	floatingTotalValue float64,
	colors []color.NRGBA,
	items []PieChartItem,
) layout.Dimensions {
	colorBoxSize := gtx.Dp(pc.ColorBoxSize)
	widgetItems := make([]layout.Widget, 0, max(len(items), 1))

	for i, item := range items {
		percentage := float64(item.Value) / floatingTotalValue

		widgetItems = append(widgetItems, func(gtx layout.Context) layout.Dimensions {
			return layout.Flex{
				Axis:      layout.Horizontal,
				Alignment: layout.Middle,
			}.Layout(
				gtx,
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					size := image.Point{
						X: colorBoxSize,
						Y: colorBoxSize,
					}

					defer clip.UniformRRect(image.Rectangle{Max: size}, 5).Push(gtx.Ops).Pop()
					paint.Fill(gtx.Ops, colors[i])

					return layout.Dimensions{
						Size: size,
					}
				}),
				utils.FlexSpacerW(utils.CommonSpacing),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return layout.Flex{
						Axis: layout.Vertical,
					}.Layout(
						gtx,
						layout.Rigid(material.Label(pc.theme, pc.TextSize, item.Name).Layout),
						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							style := material.Label(
								pc.theme,
								pc.TextSize,
								fmt.Sprintf("%v (%.1f%%)", item.SubText, percentage*100),
							)
							style.Color = pc.SubTextColor
							return style.Layout(gtx)
						}),
					)
				}),
			)
		})
	}

	if len(items) >= pc.OverflowLimit {
		return HorizontalWrap{
			Alignment:   layout.Middle,
			Spacing:     utils.CommonSpacing,
			LineSpacing: utils.CommonSpacing,
		}.Layout(gtx, widgetItems...)
	} else {
		flexItems := make([]layout.FlexChild, 0, min(1, len(widgetItems)*2-1))

		for i, item := range widgetItems {
			if i != 0 {
				flexItems = append(flexItems, utils.FlexSpacerH(utils.CommonSpacing))
			}

			flexItems = append(flexItems, layout.Rigid(item))
		}

		return layout.Flex{
			Axis: layout.Vertical,
		}.Layout(gtx, flexItems...)
	}

}

func (pc PieChart) Layout(gtx layout.Context, totalValue int, items ...PieChartItem) layout.Dimensions {
	itemsLength := len(items)

	// Say 'No data' if empty
	if itemsLength <= 0 {
		return Canvas{
			ExpandHorizontal: true,
			MinSize: image.Point{
				Y: gtx.Dp(pc.MinHeight),
			},
		}.Layout(
			gtx,
			CanvasItem{
				Anchor: layout.Center,
				Widget: func(gtx layout.Context) layout.Dimensions {
					style := material.Label(pc.theme, pc.TextSize, "No Data")
					style.Color = pc.SubTextColor
					return style.Layout(gtx)
				},
			},
		)
	}

	floatingTotalValue := float64(totalValue)

	angles := make([]float64, len(items))
	colors := make([]color.NRGBA, len(items))

	for i, item := range items {
		percent := float64(item.Value) / floatingTotalValue
		angles[i] = TAU * percent
		colors[i] = StringToColor(item.Name)
	}

	cgtx := gtx
	cgtx.Constraints.Min = image.Point{}

	macro := op.Record(gtx.Ops)
	legendDims := pc.legend(cgtx, floatingTotalValue, colors, items)
	legendCall := macro.Stop()

	spacing := gtx.Dp(utils.CommonSpacing)
	width := gtx.Constraints.Max.X

	var horizontalSpace, verticalSpace int

	overflowed := itemsLength >= pc.OverflowLimit

	if overflowed {
		horizontalSpace = width
		verticalSpace = max(width, gtx.Dp(pc.MinHeight))
	} else {
		horizontalSpace = width - spacing - legendDims.Size.X
		verticalSpace = max(legendDims.Size.Y, gtx.Dp(pc.MinHeight))
	}
	pieSize := min(horizontalSpace, verticalSpace)

	// Render pie chart
	xOffset := horizontalSpace/2 - pieSize/2
	yOffset := verticalSpace/2 - pieSize/2

	transOp := op.Offset(image.Point{
		X: xOffset,
		Y: yOffset,
	}).Push(gtx.Ops)
	clipOp := clip.Rect{
		Max: image.Point{
			X: pieSize,
			Y: pieSize,
		},
	}.Push(gtx.Ops)

	floatingPieSize := float32(pieSize)
	halfPie := floatingPieSize / 2

	pieTransOp := op.Offset(f32.Pt(
		halfPie,
		halfPie,
	).Round()).Push(gtx.Ops)

	previousAngle := -(math.Pi / 2)
	for i, angle := range angles {
		clr := colors[i]

		vy, vx := math.Sincos(previousAngle)
		pen := f32.Pt(float32(vx), float32(vy)).Mul(halfPie)

		path := clip.Path{}
		path.Begin(gtx.Ops)

		path.MoveTo(f32.Pt(0, 0))
		path.LineTo(pen)

		center := f32.Pt(0, 0).Sub(pen)
		path.Arc(center, center, float32(angle))
		path.Close()

		paint.FillShape(gtx.Ops, clr, clip.Outline{
			Path: path.End(),
		}.Op())

		previousAngle += angle
	}

	pieTransOp.Pop()
	clipOp.Pop()
	transOp.Pop()

	// Render the legend
	var transformOffset image.Point
	if overflowed {
		transformOffset = image.Point{
			X: 0,
			Y: verticalSpace + spacing,
		}
	} else {
		transformOffset = image.Point{
			X: horizontalSpace + spacing,
			Y: verticalSpace/2 - legendDims.Size.Y/2,
		}
	}

	trans := op.Offset(transformOffset).Push(gtx.Ops)
	legendCall.Add(gtx.Ops)
	trans.Pop()

	if overflowed {
		return layout.Dimensions{
			Size: image.Point{
				X: width,
				Y: verticalSpace + spacing + legendDims.Size.Y,
			},
		}
	} else {
		return layout.Dimensions{
			Size: image.Point{
				X: width,
				Y: verticalSpace,
			},
		}
	}

}
