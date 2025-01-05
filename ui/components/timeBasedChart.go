package components

import (
	"PGCombatTracker/utils"
	"gioui.org/f32"
	"gioui.org/layout"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/unit"
	"image"
	"image/color"
	"math"
	"slices"
	"time"
)

type TimePoint struct {
	Time  time.Time
	Value int
}

func (t TimePoint) Interpolate(other TimePoint, interpolant time.Time) TimePoint {
	var interpolationValue float64

	if interpolant.After(other.Time) {
		interpolationValue = 1
	} else if interpolant.Before(t.Time) {
		interpolationValue = 0
	} else {
		total := other.Time.Sub(t.Time).Seconds()
		interpolantDiff := interpolant.Sub(t.Time).Seconds()

		interpolationValue = interpolantDiff / total
	}

	valueDifference := other.Value - t.Value
	interpolatedDifference := int(math.Floor(float64(valueDifference) * interpolationValue))

	interpolatedValue := t.Value + interpolatedDifference

	return TimePoint{
		Time:  interpolant,
		Value: interpolatedValue,
	}
}

type TimeFrame struct {
	From time.Time
	To   time.Time
}

func (t TimeFrame) Expand(time time.Time) TimeFrame {
	fromTime := t.From
	toTime := t.To

	if fromTime.After(time) {
		fromTime = time
	}

	if time.After(toTime) {
		toTime = time
	}

	return TimeFrame{
		From: fromTime,
		To:   toTime,
	}
}

func (t TimeFrame) LengthSeconds() float64 {
	return t.To.Sub(t.From).Seconds()
}

func (t TimeFrame) Within(time time.Time) bool {
	return (time.After(t.From) || time.Equal(t.From)) && (time.Before(t.To) || time.Equal(t.To))
}

type DataRange struct {
	Min int
	Max int
}

func (r DataRange) Expand(value int) DataRange {
	return DataRange{
		Min: min(r.Min, value),
		Max: max(r.Max, value),
	}
}

func (r DataRange) Difference() int {
	return r.Max - r.Min
}

func NewTimeBasedChart() *TimeBasedChart {
	return &TimeBasedChart{}
}

type TimeBasedChart struct {
	DataPoints        []TimePoint
	DisplayTimeFrame  TimeFrame
	DisplayValueRange DataRange

	CalculatedPoints []f32.Point
	dirty            bool
	lastSize         image.Point
	lastTimeFrame    TimeFrame
	lastValueRange   DataRange
}

func (tc *TimeBasedChart) Add(point TimePoint) {
	tc.DataPoints = append(tc.DataPoints, point)

	slices.SortFunc(tc.DataPoints, func(a, b TimePoint) int {
		return a.Time.Compare(b.Time)
	})

	length := len(tc.DataPoints)

	if length == 1 {
		tc.DisplayTimeFrame = TimeFrame{
			From: point.Time,
			To:   point.Time,
		}
		tc.DisplayValueRange.Min = point.Value - 1
		tc.DisplayValueRange.Max = point.Value + 1
	} else {
		tc.DisplayTimeFrame = tc.DisplayTimeFrame.Expand(point.Time)
		tc.DisplayValueRange = tc.DisplayValueRange.Expand(point.Value)
	}

	tc.dirty = true
}

type FloatXIntY struct {
	X float64
	Y int
}

func (tc *TimeBasedChart) RecalculatePoints(tolerance float32, width, height int) {
	tc.CalculatedPoints = CalculateTimeChartPoints(
		tc.DataPoints,
		tc.DisplayTimeFrame,
		tc.DisplayValueRange,
		tolerance,
		width,
		height,
	)
}

func CalculateTimeChartPoints(
	dataPoints []TimePoint,
	timeFrame TimeFrame,
	valueRange DataRange,
	tolerance float32,
	width, height int,
) []f32.Point {
	floatingWidth := float64(width)
	totalTimeFrame := timeFrame.LengthSeconds()

	// Filter and plot data, but also keep adjacent points
	var previousPointSet bool
	var hasPreviousPoint bool
	var previousPoint TimePoint
	var hasNextPoint bool
	var nextPoint TimePoint

	var filteredData []TimePoint
	var lastDataInRange bool

	var resultPointArray []FloatXIntY

	for _, point := range dataPoints {
		dataInRange := timeFrame.Within(point.Time)

		if dataInRange {
			filteredData = append(filteredData, point)

			timePosition := point.Time.Sub(timeFrame.From).Seconds()
			timeFramePosition := timePosition / totalTimeFrame
			xPosition := floatingWidth * timeFramePosition

			value := point.Value

			resultPointArray = append(
				resultPointArray,
				FloatXIntY{
					X: xPosition,
					Y: value,
				},
			)
		}

		if dataInRange != lastDataInRange {
			if dataInRange {
				if previousPointSet {
					hasPreviousPoint = true
				}
			} else {
				nextPoint = point
				hasNextPoint = true
			}
		} else {
			if !hasPreviousPoint {
				previousPoint = point
				previousPointSet = true
			}
		}

		lastDataInRange = dataInRange
	}

	if len(resultPointArray) <= 0 {
		resultPointArray = append(
			resultPointArray,
			FloatXIntY{
				X: floatingWidth / 2,
				Y: 0,
			},
		)
	}

	// Add interpolated versions of points for previous and next
	if hasPreviousPoint {
		interpolatedValue := previousPoint.Interpolate(filteredData[0], timeFrame.From).Value

		resultPointArray = slices.Insert(
			resultPointArray,
			0,
			FloatXIntY{
				X: 0,
				Y: interpolatedValue,
			},
		)
	} else {
		resultPointArray = slices.Insert(
			resultPointArray,
			0,
			FloatXIntY{
				X: 0,
				Y: resultPointArray[0].Y,
			},
		)
	}

	if hasNextPoint {
		interpolatedValue := filteredData[len(filteredData)-1].Interpolate(nextPoint, timeFrame.To).Value

		resultPointArray = append(
			resultPointArray,
			FloatXIntY{
				X: floatingWidth,
				Y: interpolatedValue,
			},
		)
	} else {
		resultPointArray = append(
			resultPointArray,
			FloatXIntY{
				X: floatingWidth,
				Y: resultPointArray[len(resultPointArray)-1].Y,
			},
		)
	}

	floatingHeight := float64(height)
	conversionValue := floatingHeight / float64(valueRange.Difference())

	var newPoints []f32.Point
	for _, point := range resultPointArray {
		valueFromMin := point.Y - valueRange.Min
		adjustedValue := float64(valueFromMin) * conversionValue

		if adjustedValue < 0 || adjustedValue > floatingHeight {
			continue
		}

		newPoints = append(newPoints, f32.Point{
			X: float32(point.X),
			Y: float32(adjustedValue),
		})
	}

	newPoints = utils.SimplifyPoints(newPoints, tolerance)

	return newPoints
}

func StyleTimeBasedChart(chart *TimeBasedChart) TimeBasedChartStyle {
	return TimeBasedChartStyle{
		Color:           color.NRGBA{R: 255, G: 255, B: 255, A: 255},
		Alpha:           140,
		BackgroundAlpha: 40,
		MinWidth:        300,
		MinHeight:       100,
		Inset:           layout.UniformInset(5),
		chart:           chart,
	}
}

type TimeBasedChartStyle struct {
	Color           color.NRGBA
	Alpha           uint8
	BackgroundAlpha uint8

	MinWidth  unit.Dp
	MinHeight unit.Dp

	Inset layout.Inset

	chart *TimeBasedChart
}

func (ts TimeBasedChartStyle) CheckAndRecalculate(width, height int) {
	size := image.Point{
		X: width,
		Y: height,
	}

	if ts.chart.lastSize != size {
		ts.chart.lastSize = size
		ts.chart.dirty = true
	}

	if ts.chart.lastTimeFrame != ts.chart.DisplayTimeFrame {
		ts.chart.lastTimeFrame = ts.chart.DisplayTimeFrame
		ts.chart.dirty = true
	}

	if ts.chart.lastValueRange != ts.chart.DisplayValueRange {
		ts.chart.lastValueRange = ts.chart.DisplayValueRange
		ts.chart.dirty = true
	}

	if ts.chart.dirty {
		ts.chart.RecalculatePoints(0.5, width, height)
		ts.chart.dirty = false
	}
}

func (ts TimeBasedChartStyle) Layout(gtx layout.Context) layout.Dimensions {
	size := gtx.Constraints.Min

	pxWidth := gtx.Dp(ts.MinWidth)
	if size.X < pxWidth {
		size.X = pxWidth
	}

	pxHeight := gtx.Dp(ts.MinHeight)
	if size.Y < pxHeight {
		size.Y = pxHeight
	}

	ts.CheckAndRecalculate(size.X, size.Y)

	points := ts.chart.CalculatedPoints

	defer clip.UniformRRect(image.Rectangle{Max: size}, 10).Push(gtx.Ops).Pop()
	semiTransparent := ts.Color
	semiTransparent.A = ts.BackgroundAlpha
	paint.Fill(gtx.Ops, semiTransparent)

	path := clip.Path{}

	path.Begin(gtx.Ops)

	floatingSizeX := float32(size.X)
	floatingSizeY := float32(size.Y)

	if len(points) > 0 {
		path.MoveTo(f32.Point{
			X: points[0].X,
			Y: floatingSizeY - points[0].Y,
		})

		for i := 1; i < len(points); i++ {
			path.LineTo(f32.Point{
				X: points[i].X,
				Y: floatingSizeY - points[i].Y,
			})
		}

		path.LineTo(f32.Point{
			X: floatingSizeX,
			Y: floatingSizeY,
		})

		path.LineTo(f32.Point{
			X: 0,
			Y: floatingSizeY,
		})

		path.Close()
	}

	areaColor := ts.Color
	areaColor.A = ts.Alpha

	paint.FillShape(gtx.Ops, areaColor, clip.Outline{
		Path: path.End(),
	}.Op())

	return layout.Dimensions{
		Size: size,
	}
}
