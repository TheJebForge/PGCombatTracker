package drawing

import (
	"PGCombatTracker/ui/components"
	"PGCombatTracker/utils"
	"fmt"
	"gioui.org/layout"
	"github.com/fogleman/gg"
	"image/color"
	"math"
)

var TAU = math.Pi * 2

type PieChart struct {
	ColorBoxSize float64
	TextStyle    TextStyle
	SubTextStyle TextStyle
}

type PieChartItem struct {
	Value   int
	Name    string
	SubText string
}

func (pc PieChart) legend(floatingTotalValue float64, colors []color.NRGBA, items []PieChartItem) Widget {
	return func(ltx Context) Result {
		flexItems := make([]FlexChild, 0, max(len(items)*2-1, 1))

		for i, item := range items {
			percentage := float64(item.Value) / floatingTotalValue

			if i != 0 {
				flexItems = append(flexItems, FlexVSpacer(CommonSpacing))
			}

			flexItems = append(
				flexItems,
				Rigid(Flex{
					Axis:      layout.Horizontal,
					Alignment: layout.Middle,
				}.Layout(
					Rigid(func(ltx Context) Result {
						size := F64(pc.ColorBoxSize, pc.ColorBoxSize)

						return Result{
							Size: size,
							Draw: func(gg *gg.Context) {
								gg.SetColor(colors[i])
								gg.DrawRoundedRectangle(0, 0, pc.ColorBoxSize, pc.ColorBoxSize, CommonSpacing)
								gg.Fill()
							},
						}
					}),
					FlexHSpacer(CommonSpacing),
					Rigid(Flex{
						Axis: layout.Vertical,
					}.Layout(
						Rigid(pc.TextStyle.Layout(item.Name)),
						FlexVSpacer(CommonSpacing),
						Rigid(pc.SubTextStyle.Layout(fmt.Sprintf("%v (%.1f%%)", item.SubText, percentage*100))),
					)),
				)),
			)
		}

		return Flex{
			Axis: layout.Vertical,
		}.Layout(
			flexItems...,
		)(ltx)
	}
}

func (pc PieChart) Layout(totalValue int, items ...PieChartItem) Widget {
	return func(ltx Context) Result {
		if len(items) <= 0 {
			text := pc.TextStyle.Layout("No Data")(ltx)

			return Result{
				Size: F64(ltx.Max.X, text.Size.Y),
				Draw: func(gg *gg.Context) {
					gg.Push()
					gg.Translate(ltx.Max.X/2-text.Size.X/2, 0)

					text.Draw(gg)

					gg.Pop()
				},
			}
		}

		floatingTotalValue := float64(totalValue)

		angles := make([]float64, len(items))
		colors := make([]color.NRGBA, len(items))

		for i, item := range items {
			percent := float64(item.Value) / floatingTotalValue
			angles[i] = TAU * percent
			colors[i] = components.StringToColor(item.Name)
		}

		return Flex{
			Axis:      layout.Horizontal,
			Alignment: layout.Middle,
			ExpandW:   true,
		}.Layout(
			Flexed(1, func(ltx Context) Result {
				chartSideLength := min(ltx.Max.X, ltx.Max.Y)
				size := ltx.Max

				return Result{
					Size: size,
					Draw: func(gg *gg.Context) {
						gg.DrawImage(
							utils.PieImage{
								Size:   int(math.Floor(chartSideLength)),
								Angles: angles,
								Colors: colors,
							},
							int(math.Floor(size.X/2-chartSideLength/2)),
							int(math.Floor(size.Y/2-chartSideLength/2)),
						)
					},
				}
			}),
			FlexHSpacer(CommonSpacing),
			Rigid(pc.legend(floatingTotalValue, colors, items)),
		)(ltx)
	}
}
