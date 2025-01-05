package components

import (
	"PGCombatTracker/utils"
	"fmt"
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
		MinHeight: 200,

		ColorBoxSize: utils.CommonSpacing * 2,
		TextSize:     theme.TextSize,
		TextColor:    theme.Fg,
		SubTextColor: utils.GrayText,

		theme: theme,
	}
}

type PieChart struct {
	MinHeight unit.Dp

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
	flexItems := make([]layout.FlexChild, 0, max(len(items)*2-1, 1))

	for i, item := range items {
		percentage := float64(item.Value) / floatingTotalValue

		if i != 0 {
			flexItems = append(flexItems, utils.FlexSpacerH(utils.CommonSpacing))
		}

		flexItems = append(flexItems, layout.Rigid(func(gtx layout.Context) layout.Dimensions {
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
		}))
	}

	return layout.Flex{
		Axis: layout.Vertical,
	}.Layout(gtx, flexItems...)
}

func (pc PieChart) Layout(gtx layout.Context, totalValue int, items ...PieChartItem) layout.Dimensions {
	// Say 'No data' if empty
	if len(items) <= 0 {
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

	horizonalSpace := width - legendDims.Size.X - spacing
	verticalSpace := max(legendDims.Size.Y, gtx.Dp(pc.MinHeight))
	pieSize := min(horizonalSpace, verticalSpace)

	// Render pie chart
	xOffset := horizonalSpace/2 - pieSize/2
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
	paint.NewImageOp(utils.PieImage{
		Size:   pieSize,
		Angles: angles,
		Colors: colors,
	}).Add(gtx.Ops)
	paint.PaintOp{}.Add(gtx.Ops)
	clipOp.Pop()
	transOp.Pop()

	// Render the legend
	trans := op.Offset(image.Point{
		X: horizonalSpace + spacing,
		Y: verticalSpace/2 - legendDims.Size.Y/2,
	}).Push(gtx.Ops)
	legendCall.Add(gtx.Ops)
	trans.Pop()

	return layout.Dimensions{
		Size: image.Point{
			X: width,
			Y: verticalSpace,
		},
	}
}
