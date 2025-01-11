package drawing

import (
	"PGCombatTracker/ui/components"
	"PGCombatTracker/utils"
	"gioui.org/f32"
	"gioui.org/layout"
	"github.com/fogleman/gg"
	"image/color"
	"math"
)

func StyleAreaChart(chart *components.TimeBasedChart, color color.NRGBA) SingleAreaChart {
	return SingleAreaChart{
		Chart:           chart,
		Color:           color,
		Alpha:           255,
		BackgroundAlpha: 120,
		MinHeight:       300,
	}
}

type SingleAreaChart struct {
	Chart           *components.TimeBasedChart
	Color           color.NRGBA
	Alpha           uint8
	BackgroundAlpha uint8
	MinHeight       float64
}

func (ac SingleAreaChart) Layout() Widget {
	return func(ltx Context) Result {
		size := F64(
			ltx.Max.X,
			max(ltx.Min.Y, ac.MinHeight),
		)
		points := components.CalculateTimeChartPoints(
			ac.Chart.DataPoints,
			ac.Chart.DisplayTimeFrame,
			ac.Chart.DisplayValueRange,
			0.1,
			int(math.Floor(size.X)),
			int(math.Floor(size.Y)),
		)

		return Result{
			Size: size,
			Draw: func(gg *gg.Context) {
				gg.Push()

				// Do the clip
				gg.DrawRoundedRectangle(0, 0, size.X, size.Y, CommonSpacing*2)
				gg.ClipPreserve()

				// Background
				gg.SetColor(color.NRGBA{
					R: ac.Color.R,
					G: ac.Color.G,
					B: ac.Color.B,
					A: ac.BackgroundAlpha,
				})
				gg.Fill()

				// Draw area
				if len(points) > 0 {
					gg.MoveTo(float64(points[0].X), size.Y-float64(points[0].Y))

					for i := 1; i < len(points); i++ {
						gg.LineTo(float64(points[i].X), size.Y-float64(points[i].Y))
					}

					gg.LineTo(size.X, size.Y)
					gg.LineTo(0, size.Y)
					gg.ClosePath()

					gg.SetColor(color.NRGBA{
						R: ac.Color.R,
						G: ac.Color.G,
						B: ac.Color.B,
						A: ac.Alpha,
					})
					gg.Fill()
				}

				gg.ResetClip()

				gg.Pop()
			},
		}
	}
}

func StyleStackedAreaChart(styledFonts *StyledFontPack, chart *components.StackedTimeBasedChart) StackedAreaChart {
	return StackedAreaChart{
		Chart:           chart,
		BackgroundColor: utils.LesserContrastBg,
		Alpha:           255,
		MinHeight:       300,
		ColorBoxSize:    CommonSpacing * 4,
		TextStyle:       styledFonts.Body,
		SubTextStyle:    MakeTextStyle(styledFonts.Smaller.Face, utils.GrayText),
	}
}

type StackedAreaChart struct {
	Chart           *components.StackedTimeBasedChart
	BackgroundColor color.NRGBA
	Alpha           uint8
	MinHeight       float64
	ColorBoxSize    float64
	TextStyle       TextStyle
	SubTextStyle    TextStyle
	LongFormat      bool
}

func (sac StackedAreaChart) legend(bds []components.StackedBreakdown) Widget {
	return func(ltx Context) Result {
		if len(bds) <= 0 {
			return Empty(ltx)
		}

		items := bds[len(bds)-1].Items

		widgetItems := make([]Widget, 0, max(len(items), 1))

		for i := len(items) - 1; i >= 0; i-- {
			item := items[i]
			widgetItems = append(
				widgetItems,
				Flex{
					Axis:      layout.Horizontal,
					Alignment: layout.Middle,
				}.Layout(
					Rigid(func(ltx Context) Result {
						size := F64(sac.ColorBoxSize, sac.ColorBoxSize)

						return Result{
							Size: size,
							Draw: func(gg *gg.Context) {
								gg.SetColor(components.StringToColor(item.Name))
								gg.DrawRoundedRectangle(0, 0, sac.ColorBoxSize, sac.ColorBoxSize, CommonSpacing)
								gg.Fill()
							},
						}
					}),
					FlexHSpacer(CommonSpacing),
					Rigid(Flex{
						Axis: layout.Vertical,
					}.Layout(
						Rigid(sac.TextStyle.Layout(item.Name)),
						FlexVSpacer(CommonSpacing),
						Rigid(sac.SubTextStyle.Layout(item.Details.StringCL(sac.LongFormat))),
					)),
				),
			)
		}

		return HorizontalWrap{
			Alignment:   layout.Middle,
			Spacing:     CommonSpacing * 4,
			LineSpacing: CommonSpacing * 2,
		}.Layout(
			widgetItems...,
		)(ltx)
	}
}

func (sac StackedAreaChart) Layout() Widget {
	return func(ltx Context) Result {
		size := F64(
			ltx.Max.X,
			max(ltx.Min.Y, sac.MinHeight),
		)

		breakdowns, lines := sac.Chart.CalculatePoints(
			0.1,
			int(math.Floor(size.X)),
			int(math.Floor(size.Y)),
		)

		return Flex{
			Axis: layout.Vertical,
		}.Layout(
			Rigid(func(ltx Context) Result {
				return Result{
					Size: size,
					Draw: func(gg *gg.Context) {
						gg.Push()

						// Do the clip
						gg.DrawRoundedRectangle(0, 0, size.X, size.Y, CommonSpacing*2)
						gg.ClipPreserve()

						// Background
						gg.SetColor(sac.BackgroundColor)
						gg.Fill()

						// Draw area
						var previousLine []f32.Point
						for lineIndex, line := range lines {
							if len(line) > 0 {
								gg.MoveTo(float64(line[0].X), size.Y-float64(line[0].Y))

								for i := 1; i < len(line); i++ {
									gg.LineTo(float64(line[i].X), size.Y-float64(line[i].Y))
								}

								if len(previousLine) > 0 {
									for i := len(previousLine) - 1; i >= 0; i-- {
										gg.LineTo(float64(previousLine[i].X), size.Y-float64(previousLine[i].Y))
									}
								} else {
									gg.LineTo(size.X, size.Y)
									gg.LineTo(0, size.Y)
								}

								gg.ClosePath()

								areaColor := components.StringToColor(sac.Chart.SourceNames[lineIndex])
								areaColor.A = sac.Alpha

								gg.SetColor(areaColor)
								gg.Fill()

								previousLine = line
							}
						}

						gg.ResetClip()

						gg.Pop()
					},
				}
			}),
			FlexVSpacer(CommonSpacing),
			Rigid(sac.legend(breakdowns)),
		)(ltx)
	}
}
