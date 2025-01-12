package components

import (
	"PGCombatTracker/utils"
	"gioui.org/f32"
	"gioui.org/io/event"
	"gioui.org/io/pointer"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/unit"
	"gioui.org/widget/material"
	"github.com/samber/lo"
	"image"
	"image/color"
	"math"
	"slices"
	"time"
)

type StackedBreakdownItem struct {
	Name    string
	Details utils.InterpolatableLongFormatable
}

type StackedBreakdown struct {
	X     float32
	Time  time.Time
	Items []StackedBreakdownItem
}

func NewStackedTimeBasedChart() *StackedTimeBasedChart {
	return &StackedTimeBasedChart{}
}

type StackedTimeBasedChart struct {
	Sources           []*TimeBasedChart
	SourceNames       []string
	DisplayTimeFrame  TimeFrame
	DisplayValueRange DataRange

	CalculatedLines     [][]f32.Point
	CalculatedBreakdown []StackedBreakdown
	dirty               bool
	lastSize            image.Point
	lastTimeFrame       TimeFrame
	lastValueRange      DataRange

	hoveredPoint  *StackedBreakdown
	hoveredPointX int
	hoverPosition image.Point
	hoverBounds   f32.Point

	// Pointer tracking stuff
	pid     pointer.ID
	hovered bool
}

func (stc *StackedTimeBasedChart) Add(source *TimeBasedChart, name string) {
	stc.Sources = append(stc.Sources, source)
	stc.SourceNames = append(stc.SourceNames, name)
	stc.dirty = true
}

func YatXOnLine(points []f32.Point, targetX float32) float32 {
	pointsLength := len(points)

	if pointsLength == 0 {
		return 0
	} else if pointsLength == 1 {
		return points[0].Y
	} else {
		previousPoint := points[0]

		for i := 1; i < pointsLength; i++ {
			point := points[i]

			// Interpolate between previous point and current point to get Y value
			if targetX < point.X {
				width := point.X - previousPoint.X

				// There's no point to interpolate as there's no distance between points
				if width == 0 {
					return point.Y
				}

				targetOffset := targetX - previousPoint.X
				proportion := min(max(0, targetOffset/width), 1)

				verticalDifference := point.Y - previousPoint.Y

				return previousPoint.Y + verticalDifference*proportion
			}

			previousPoint = point
		}

		// Target is outside of data range
		return previousPoint.Y
	}
}

func InterpolatedTimePoint(points []TimePoint, target time.Time) TimePoint {
	pointsLength := len(points)

	if pointsLength == 0 {
		return TimePoint{}
	} else if pointsLength == 1 {
		return points[0]
	} else {
		previousPoint := points[0]

		for i := 1; i < pointsLength; i++ {
			point := points[i]

			// Interpolate between previous point and current point to get Y value
			if target.Before(point.Time) {
				secondsDifference := point.Time.Sub(previousPoint.Time).Seconds()

				// There's no point to interpolate as there's no distance between points
				if secondsDifference == 0 {
					return point
				}

				secondsOffset := target.Sub(previousPoint.Time).Seconds()
				proportion := min(max(0, secondsOffset/secondsDifference), 1)

				return TimePoint{
					Time:    target,
					Value:   utils.LerpInt(previousPoint.Value, point.Value, proportion),
					Details: previousPoint.Details.InterpolateILF(point.Details, proportion),
				}
			}

			previousPoint = point
		}

		// Target is outside of data range
		return previousPoint
	}
}

func (stc *StackedTimeBasedChart) CalculatePoints(tolerance float32, width, height int) ([]StackedBreakdown, [][]f32.Point) {
	sourceLines := make([][]f32.Point, 0, len(stc.Sources))

	allLocations := 0
	for _, source := range stc.Sources {
		sourceLine := CalculateTimeChartPoints(
			source.DataPoints,
			stc.DisplayTimeFrame,
			stc.DisplayValueRange,
			tolerance,
			width,
			height,
		)
		sourceLines = append(
			sourceLines,
			sourceLine,
		)
		allLocations += len(sourceLine)
	}

	xLocations := make([]float32, 0, allLocations)

	for _, line := range sourceLines {
		for _, point := range line {
			xLocations = append(xLocations, point.X)
		}
	}

	slices.Sort(xLocations)
	xLocations = lo.Uniq(xLocations)

	floatingWidth := float64(stc.lastSize.X)

	lineCount := len(sourceLines)

	newLines := make([][]f32.Point, lineCount)
	breakdowns := make([]StackedBreakdown, len(xLocations))

	for xI, x := range xLocations {
		var currentHeight float32

		proportion := float64(x) / floatingWidth
		xTime := stc.DisplayTimeFrame.TimeAtProportion(proportion)

		breakdown := StackedBreakdown{
			X:     x,
			Time:  xTime,
			Items: make([]StackedBreakdownItem, lineCount),
		}

		for i, sourceLine := range sourceLines {
			y := max(YatXOnLine(sourceLine, x), 0)

			timePoint := InterpolatedTimePoint(stc.Sources[i].DataPoints, xTime)
			newLines[i] = append(newLines[i], f32.Point{
				X: x,
				Y: y + currentHeight,
			})
			breakdown.Items[i] = StackedBreakdownItem{
				Name:    stc.Sources[i].Name,
				Details: timePoint.Details,
			}
			currentHeight += y
		}
		slices.Reverse(breakdown.Items)
		breakdowns[xI] = breakdown
	}

	return breakdowns, newLines
}

func (stc *StackedTimeBasedChart) RecalculatePoints(tolerance float32, width, height int) {
	breakdowns, newLines := stc.CalculatePoints(tolerance, width, height)
	stc.CalculatedBreakdown = breakdowns
	stc.CalculatedLines = newLines
}

func StyleStackedTimeBasedChart(theme *material.Theme, chart *StackedTimeBasedChart) StackedTimeBasedChartStyle {
	return StackedTimeBasedChartStyle{
		Alpha:           140,
		BackgroundColor: theme.Bg,
		MinHeight:       100,
		ColorBoxSize:    utils.CommonSpacing * 2,
		TextSize:        theme.TextSize,
		TextColor:       theme.Fg,
		SubTextColor:    utils.GrayText,
		TooltipTextSize: 10,
		Inset:           layout.UniformInset(5),
		chart:           chart,
		theme:           theme,
	}
}

type StackedTimeBasedChartStyle struct {
	Alpha uint8

	BackgroundColor color.NRGBA

	MinHeight unit.Dp

	ColorBoxSize unit.Dp
	TextSize     unit.Sp
	TextColor    color.NRGBA
	SubTextColor color.NRGBA

	TooltipTextSize unit.Sp
	LongFormat      bool

	Inset layout.Inset

	chart *StackedTimeBasedChart
	theme *material.Theme
}

func (sts StackedTimeBasedChartStyle) CheckAndRecalculate(width, height int) {
	size := image.Point{
		X: width,
		Y: height,
	}

	if sts.chart.lastSize != size {
		sts.chart.lastSize = size
		sts.chart.dirty = true
	}

	if sts.chart.lastTimeFrame != sts.chart.DisplayTimeFrame {
		sts.chart.lastTimeFrame = sts.chart.DisplayTimeFrame
		sts.chart.dirty = true
	}

	if sts.chart.lastValueRange != sts.chart.DisplayValueRange {
		sts.chart.lastValueRange = sts.chart.DisplayValueRange
		sts.chart.dirty = true
	}

	if sts.chart.dirty {
		sts.chart.hoverBounds = f32.Point{}
		sts.chart.RecalculatePoints(0.1, width, height)
		sts.chart.dirty = false
	}
}

func ClosestBreakdownToTarget(points []StackedBreakdown, target float32) (StackedBreakdown, int) {
	pointsLength := len(points)

	if pointsLength == 0 {
		return StackedBreakdown{}, -1
	} else if pointsLength == 1 {
		return points[0], 0
	} else {
		previousPoint := points[0]

		for i := 1; i < pointsLength; i++ {
			point := points[i]

			// Select point based on proximity
			if target < point.X {
				xDifference := point.X - previousPoint.X

				// There's no point to interpolate as there's no distance between points
				if xDifference == 0 {
					return point, i
				}

				xOffset := target - previousPoint.X
				proportion := xOffset / xDifference

				if proportion > .5 {
					return point, i
				} else {
					return previousPoint, i - 1
				}
			}

			previousPoint = point
		}

		// Target is outside of data range
		return previousPoint, pointsLength - 1
	}
}

func (sts StackedTimeBasedChartStyle) update(gtx layout.Context) {
	ch := sts.chart
	for {
		ev, ok := gtx.Source.Event(pointer.Filter{
			Target: sts.chart,
			Kinds:  pointer.Enter | pointer.Move | pointer.Leave | pointer.Cancel,
		})
		if !ok {
			break
		}
		e, ok := ev.(pointer.Event)
		if !ok {
			continue
		}

		switch e.Kind {
		case pointer.Leave, pointer.Cancel:
			if ch.hovered && ch.pid == e.PointerID {
				ch.hovered = false
				ch.hoveredPoint = nil
				ch.hoverBounds = f32.Point{}
			}
		case pointer.Enter:
			if !ch.hovered {
				ch.pid = e.PointerID
			}

			if ch.pid == e.PointerID {
				ch.hovered = true
			}
		case pointer.Move:
			if ch.pid == e.PointerID {
				ch.hoverPosition = e.Position.Round()

				if e.Position.X > ch.hoverBounds.X && e.Position.X < ch.hoverBounds.Y {
					continue
				}

				pointsLength := len(ch.CalculatedBreakdown)

				floatingWidth := float64(ch.lastSize.X)

				closestPoint, pointIndex := ClosestBreakdownToTarget(ch.CalculatedBreakdown, e.Position.X)
				pointPosition := closestPoint.X

				ch.hoveredPointX = int(math.Round(float64(pointPosition)))
				ch.hoveredPoint = &closestPoint

				// Set the bounds that will make the hover change, so we don't do expensive shit on every mouse move
				if pointsLength <= 1 {
					ch.hoverBounds = f32.Point{
						X: 0,
						Y: float32(floatingWidth),
					}
				} else {
					// Set left bound
					if pointIndex <= 0 {
						ch.hoverBounds.X = 0
					} else {
						leftIndex := pointIndex - 1
						leftPosition := ch.CalculatedBreakdown[leftIndex].X
						leftBound := (pointPosition-leftPosition)/2 + leftPosition
						ch.hoverBounds.X = leftBound
					}

					// Set right bound
					if pointIndex >= pointsLength-1 {
						ch.hoverBounds.Y = float32(floatingWidth)
					} else {
						rightIndex := pointIndex + 1
						rightPosition := ch.CalculatedBreakdown[rightIndex].X
						rightBound := (rightPosition-pointPosition)/2 + pointPosition
						ch.hoverBounds.Y = rightBound
					}
				}
			}
		default:
		}
	}
}

func (sts StackedTimeBasedChartStyle) tooltip(point StackedBreakdown) layout.Widget {
	return func(gtx layout.Context) layout.Dimensions {
		colorBoxSize := gtx.Dp(sts.ColorBoxSize)

		return layout.Background{}.Layout(
			gtx,
			utils.MakeRoundedBG(10, utils.SecondBG),
			func(gtx layout.Context) layout.Dimensions {
				return layout.UniformInset(utils.CommonSpacing).Layout(
					gtx,
					func(gtx layout.Context) layout.Dimensions {
						flexItems := make([]layout.FlexChild, 0, len(point.Items)+1)

						flexItems = append(flexItems,
							layout.Rigid(material.Label(sts.theme, sts.TooltipTextSize, point.Time.Format(time.DateTime)).Layout),
						)

						for _, item := range point.Items {
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
										paint.Fill(gtx.Ops, StringToColor(item.Name))

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
											layout.Rigid(material.Label(sts.theme, sts.TooltipTextSize, item.Name).Layout),
											layout.Rigid(material.Label(sts.theme, sts.TooltipTextSize, item.Details.StringCL(sts.LongFormat)).Layout),
										)
									}),
								)
							}))
						}

						return layout.Flex{
							Axis: layout.Vertical,
						}.Layout(
							gtx,
							flexItems...,
						)
					},
				)
			},
		)
	}
}

func (sts StackedTimeBasedChartStyle) legend(
	gtx layout.Context,
) layout.Dimensions {
	bds := sts.chart.CalculatedBreakdown
	if len(bds) <= 0 {
		return layout.Dimensions{}
	}

	items := bds[len(bds)-1].Items

	colorBoxSize := gtx.Dp(sts.ColorBoxSize)
	flexItems := make([]layout.FlexChild, 0, max(len(items)*2-1, 1))

	for i, item := range items {
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
					paint.Fill(gtx.Ops, StringToColor(item.Name))

					return layout.Dimensions{
						Size: size,
					}
				}),
				utils.FlexSpacerW(utils.CommonSpacing),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					sub := material.Label(sts.theme, sts.TextSize, item.Details.StringCL(sts.LongFormat))
					sub.Color = sts.SubTextColor
					return layout.Flex{
						Axis: layout.Vertical,
					}.Layout(
						gtx,
						layout.Rigid(material.Label(sts.theme, sts.TextSize, item.Name).Layout),
						layout.Rigid(sub.Layout),
					)
				}),
			)
		}))
	}

	return layout.Flex{
		Axis: layout.Vertical,
	}.Layout(gtx, flexItems...)
}

func (sts StackedTimeBasedChartStyle) Layout(gtx layout.Context) layout.Dimensions {
	sts.update(gtx)

	// Legend
	cgtx := gtx
	cgtx.Constraints.Min = image.Point{}

	macro := op.Record(gtx.Ops)
	legendDims := sts.legend(cgtx)
	legendCall := macro.Stop()

	spacing := gtx.Dp(utils.CommonSpacing)

	// Chart rendering calculations
	size := image.Point{
		X: gtx.Constraints.Max.X,
		Y: max(gtx.Dp(sts.MinHeight), legendDims.Size.Y),
	}

	chartSize := image.Point{
		X: size.X - spacing - legendDims.Size.X,
		Y: size.Y,
	}

	sts.CheckAndRecalculate(chartSize.X, chartSize.Y)

	lines := sts.chart.CalculatedLines

	// Render chart
	chartClip := clip.UniformRRect(image.Rectangle{Max: chartSize}, 10).Push(gtx.Ops)
	paint.Fill(gtx.Ops, sts.BackgroundColor)
	event.Op(gtx.Ops, sts.chart)

	floatingSizeX := float32(chartSize.X)
	floatingSizeY := float32(chartSize.Y)

	var previousLine []f32.Point
	for lineIndex, line := range lines {
		path := clip.Path{}

		path.Begin(gtx.Ops)

		if len(line) > 0 {
			path.MoveTo(f32.Point{
				X: line[0].X,
				Y: floatingSizeY - line[0].Y,
			})

			for i := 1; i < len(line); i++ {
				path.LineTo(f32.Point{
					X: line[i].X,
					Y: floatingSizeY - line[i].Y,
				})
			}

			if len(previousLine) > 0 {
				for i := len(previousLine) - 1; i >= 0; i-- {
					path.LineTo(f32.Point{
						X: previousLine[i].X,
						Y: floatingSizeY - previousLine[i].Y,
					})
				}
			} else {
				path.LineTo(f32.Point{
					X: floatingSizeX,
					Y: floatingSizeY,
				})

				path.LineTo(f32.Point{
					X: 0,
					Y: floatingSizeY,
				})
			}

			path.Close()
		}

		areaColor := StringToColor(sts.chart.SourceNames[lineIndex])
		areaColor.A = sts.Alpha

		paint.FillShape(gtx.Ops, areaColor, clip.Outline{
			Path: path.End(),
		}.Op())

		previousLine = line
	}

	if sts.chart.hoveredPoint != nil {
		hoverPosition := sts.chart.hoverPosition
		hoverX := sts.chart.hoveredPointX
		floatingHoverX := float32(hoverX)

		// Draw line
		line := clip.Path{}
		line.Begin(gtx.Ops)

		line.MoveTo(f32.Point{
			X: floatingHoverX,
			Y: 0,
		})
		line.LineTo(f32.Point{
			X: floatingHoverX,
			Y: float32(chartSize.Y),
		})

		paint.FillShape(gtx.Ops, utils.ChartLineColor, clip.Stroke{
			Path:  line.End(),
			Width: 2,
		}.Op())

		// Draw tooltip
		cgtx := gtx
		cgtx.Constraints.Min = image.Point{}

		tooltipMacro := op.Record(gtx.Ops)
		tooltipDims := sts.tooltip(*sts.chart.hoveredPoint)(cgtx)
		tooltipCall := tooltipMacro.Stop()

		xOffset := spacing * 2
		if hoverPosition.X+spacing*3+tooltipDims.Size.X > chartSize.X {
			xOffset = -(tooltipDims.Size.X + spacing*2)
		}
		tooltipX := max(hoverPosition.X+xOffset, spacing)

		yOffset := min(spacing*2, chartSize.Y-(hoverPosition.Y+tooltipDims.Size.Y+spacing))
		tooltipY := hoverPosition.Y + yOffset

		trans := op.Offset(image.Point{
			X: tooltipX,
			Y: tooltipY,
		}).Push(gtx.Ops)
		tooltipCall.Add(gtx.Ops)
		trans.Pop()
	}

	chartClip.Pop()

	// Render legend
	trans := op.Offset(image.Point{
		X: chartSize.X + spacing,
		Y: size.Y/2 - legendDims.Size.Y/2,
	}).Push(gtx.Ops)
	legendCall.Add(gtx.Ops)
	trans.Pop()

	return layout.Dimensions{
		Size: size,
	}
}
