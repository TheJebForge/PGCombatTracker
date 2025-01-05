package components

import (
	"PGCombatTracker/utils"
	"gioui.org/f32"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/unit"
	"gioui.org/widget/material"
	"github.com/samber/lo"
	"image"
	"image/color"
	"slices"
)

func NewStackedTimeBasedChart() *StackedTimeBasedChart {
	return &StackedTimeBasedChart{}
}

type StackedTimeBasedChart struct {
	Sources           []*TimeBasedChart
	SourceNames       []string
	DisplayTimeFrame  TimeFrame
	DisplayValueRange DataRange

	CalculatedLines [][]f32.Point
	dirty           bool
	lastSize        image.Point
	lastTimeFrame   TimeFrame
	lastValueRange  DataRange
}

func (stc *StackedTimeBasedChart) Add(source *TimeBasedChart, name string) {
	stc.Sources = append(stc.Sources, source)
	stc.SourceNames = append(stc.SourceNames, name)
	stc.dirty = true
}

func YatX(points []f32.Point, targetX float32) float32 {
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
				proportion := targetOffset / width

				verticalDifference := point.Y - previousPoint.Y

				return previousPoint.Y + verticalDifference*proportion
			}

			previousPoint = point
		}

		// Target is outside of data range
		return previousPoint.Y
	}
}

func (stc *StackedTimeBasedChart) RecalculatePoints(tolerance float32, width, height int) {
	var sourceLines [][]f32.Point

	for _, source := range stc.Sources {
		sourceLines = append(
			sourceLines,
			CalculateTimeChartPoints(
				source.DataPoints,
				stc.DisplayTimeFrame,
				stc.DisplayValueRange,
				tolerance,
				width,
				height,
			),
		)
	}

	var xLocations []float32

	for _, line := range sourceLines {
		for _, point := range line {
			xLocations = append(xLocations, point.X)
		}
	}

	slices.Sort(xLocations)
	xLocations = lo.Uniq(xLocations)

	newLines := make([][]f32.Point, len(sourceLines))

	for _, x := range xLocations {
		var currentHeight float32

		for i, sourceLine := range sourceLines {
			y := YatX(sourceLine, x)

			newLines[i] = append(newLines[i], f32.Point{
				X: x,
				Y: y + currentHeight,
			})
			currentHeight += y
		}
	}

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
		sts.chart.RecalculatePoints(0.5, width, height)
		sts.chart.dirty = false
	}
}

func (sts StackedTimeBasedChartStyle) legend(
	gtx layout.Context,
) layout.Dimensions {
	items := sts.chart.SourceNames

	colorBoxSize := gtx.Dp(sts.ColorBoxSize)
	flexItems := make([]layout.FlexChild, 0, max(len(items)*2-1, 1))

	for i, itemName := range items {
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
					paint.Fill(gtx.Ops, StringToColor(itemName))

					return layout.Dimensions{
						Size: size,
					}
				}),
				utils.FlexSpacerW(utils.CommonSpacing),
				layout.Rigid(material.Label(sts.theme, sts.TextSize, itemName).Layout),
			)
		}))
	}

	return layout.Flex{
		Axis: layout.Vertical,
	}.Layout(gtx, flexItems...)
}

func (sts StackedTimeBasedChartStyle) Layout(gtx layout.Context) layout.Dimensions {
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
