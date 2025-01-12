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
	"image"
	"image/color"
	"math"
	"slices"
	"time"
)

type TimePoint struct {
	Time    time.Time
	Value   int
	Details utils.InterpolatableLongFormatable
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

func (t TimeFrame) ProportionOfTarget(target time.Time) float64 {
	if target.Before(t.From) {
		return 0
	} else if target.After(t.To) {
		return 1
	}

	timeWidth := t.LengthSeconds()
	secondsFromLeft := target.Sub(t.From).Seconds()
	return secondsFromLeft / timeWidth
}

func (t TimeFrame) TimeAtProportion(target float64) time.Time {
	timeWidth := t.LengthSeconds()
	targetSeconds := timeWidth * target
	return t.From.Add(
		time.Microsecond * time.Duration(math.Round(targetSeconds*1000000)),
	)
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

func NewTimeBasedChart(name string) *TimeBasedChart {
	return &TimeBasedChart{
		Name: name,
	}
}

type TimeBasedChart struct {
	DataPoints        []TimePoint
	DisplayTimeFrame  TimeFrame
	DisplayValueRange DataRange
	Name              string

	CalculatedPoints []f32.Point
	dirty            bool
	lastSize         image.Point
	lastTimeFrame    TimeFrame
	lastValueRange   DataRange

	hoveredPoint  *TimePoint
	hoveredPointX int
	hoverPosition image.Point
	hoverBounds   f32.Point

	// Pointer tracking stuff
	pid     pointer.ID
	hovered bool
}

func (tc *TimeBasedChart) Add(point TimePoint) {
	tc.DataPoints = append(tc.DataPoints, point)

	slices.SortFunc(tc.DataPoints, func(a, b TimePoint) int {
		return a.Time.Compare(b.Time)
	})

	tc.dirty = true
}

func (tc *TimeBasedChart) AddWithBounds(point TimePoint) {
	tc.Add(point)

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

func ValueAtTime(points []TimePoint, target time.Time) int {
	pointsLength := len(points)

	if pointsLength == 0 {
		return 0
	} else if pointsLength == 1 {
		return points[0].Value
	} else {
		previousPoint := points[0]

		for i := 1; i < pointsLength; i++ {
			point := points[i]

			// Interpolate between previous point and current point to get Y value
			if target.Before(point.Time) {
				secondsDifference := point.Time.Sub(previousPoint.Time).Seconds()

				// There's no point to interpolate as there's no distance between points
				if secondsDifference == 0 {
					return point.Value
				}

				secondsOffset := target.Sub(previousPoint.Time).Seconds()
				proportion := secondsOffset / secondsDifference

				verticalDifference := float64(point.Value - previousPoint.Value)
				interpolatedDifference := int(math.Round(verticalDifference * proportion))

				return previousPoint.Value + interpolatedDifference
			}

			previousPoint = point
		}

		// Target is outside of data range
		return previousPoint.Value
	}
}

func ClosestPointToTarget(points []TimePoint, target time.Time) (TimePoint, int) {
	pointsLength := len(points)

	if pointsLength == 0 {
		return TimePoint{}, -1
	} else if pointsLength == 1 {
		return points[0], 0
	} else {
		previousPoint := points[0]

		for i := 1; i < pointsLength; i++ {
			point := points[i]

			// Select point based on proximity
			if target.Before(point.Time) {
				secondsDifference := point.Time.Sub(previousPoint.Time).Seconds()

				// There's no point to interpolate as there's no distance between points
				if secondsDifference == 0 {
					return point, i
				}

				secondsOffset := target.Sub(previousPoint.Time).Seconds()
				proportion := secondsOffset / secondsDifference

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

func CalculateTimeChartPoints(
	dataPoints []TimePoint,
	timeFrame TimeFrame,
	valueRange DataRange,
	tolerance float32,
	width, height int,
) []f32.Point {
	floatingWidth := float64(width)
	totalTimeFrame := timeFrame.LengthSeconds()

	// Filter and plot data
	var filteredData []TimePoint
	resultPointArray := []FloatXIntY{
		{
			X: 0,
			Y: ValueAtTime(dataPoints, timeFrame.From),
		},
	}

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
	}

	resultPointArray = append(
		resultPointArray,
		FloatXIntY{
			X: floatingWidth,
			Y: ValueAtTime(dataPoints, timeFrame.To),
		},
	)

	floatingHeight := float64(height)
	conversionValue := floatingHeight / float64(valueRange.Difference())

	var newPoints []f32.Point
	for _, point := range resultPointArray {
		valueFromMin := point.Y - valueRange.Min
		adjustedValue := max(0, min(float64(valueFromMin)*conversionValue, floatingHeight))

		newPoints = append(newPoints, f32.Point{
			X: float32(point.X),
			Y: float32(adjustedValue),
		})
	}

	newPoints = utils.SimplifyPoints(newPoints, tolerance)

	return newPoints
}

func StyleTimeBasedChart(theme *material.Theme, chart *TimeBasedChart) TimeBasedChartStyle {
	return TimeBasedChartStyle{
		Color:           color.NRGBA{R: 255, G: 255, B: 255, A: 255},
		Alpha:           140,
		BackgroundAlpha: 40,
		TooltipTextSize: 10,
		MinWidth:        300,
		MinHeight:       100,
		Inset:           layout.UniformInset(5),
		chart:           chart,
		theme:           theme,
	}
}

type TimeBasedChartStyle struct {
	Color           color.NRGBA
	Alpha           uint8
	BackgroundAlpha uint8

	TooltipTextSize unit.Sp
	LongFormat      bool

	MinWidth  unit.Dp
	MinHeight unit.Dp

	Inset layout.Inset

	theme *material.Theme
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
		ts.chart.hoverBounds = f32.Point{}
		ts.chart.RecalculatePoints(0.1, width, height)
		ts.chart.dirty = false
	}
}

func (ts TimeBasedChartStyle) update(gtx layout.Context) {
	ch := ts.chart
	for {
		ev, ok := gtx.Source.Event(pointer.Filter{
			Target: ts.chart,
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

				pointsLength := len(ch.DataPoints)

				floatingWidth := float64(ch.lastSize.X)

				cursorProportion := float64(e.Position.X) / floatingWidth
				cursorTime := ch.DisplayTimeFrame.TimeAtProportion(cursorProportion)
				closestPoint, pointIndex := ClosestPointToTarget(ch.DataPoints, cursorTime)
				pointPosition := floatingWidth * ch.DisplayTimeFrame.ProportionOfTarget(closestPoint.Time)

				ch.hoveredPointX = int(math.Round(pointPosition))
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
						leftProportion := ch.DisplayTimeFrame.ProportionOfTarget(ch.DataPoints[leftIndex].Time)
						leftPosition := floatingWidth * leftProportion
						leftBound := (pointPosition-leftPosition)/2 + leftPosition
						ch.hoverBounds.X = float32(leftBound)
					}

					// Set right bound
					if pointIndex >= pointsLength-1 {
						ch.hoverBounds.Y = float32(floatingWidth)
					} else {
						rightIndex := pointIndex + 1
						rightProportion := ch.DisplayTimeFrame.ProportionOfTarget(ch.DataPoints[rightIndex].Time)
						rightPosition := floatingWidth * rightProportion
						rightBound := (rightPosition-pointPosition)/2 + pointPosition
						ch.hoverBounds.Y = float32(rightBound)
					}
				}
			}
		default:
		}
	}
}

func (ts TimeBasedChartStyle) tooltip(point TimePoint) layout.Widget {
	return func(gtx layout.Context) layout.Dimensions {
		return layout.Background{}.Layout(
			gtx,
			utils.MakeRoundedBG(10, utils.SecondBG),
			func(gtx layout.Context) layout.Dimensions {
				return layout.UniformInset(utils.CommonSpacing).Layout(
					gtx,
					func(gtx layout.Context) layout.Dimensions {
						return layout.Flex{
							Axis: layout.Vertical,
						}.Layout(
							gtx,
							layout.Rigid(material.Label(ts.theme, ts.TooltipTextSize, point.Time.Format(time.DateTime)).Layout),
							layout.Rigid(material.Label(ts.theme, ts.TooltipTextSize, point.Details.StringCL(ts.LongFormat)).Layout),
						)
					},
				)
			},
		)
	}
}

func (ts TimeBasedChartStyle) Layout(gtx layout.Context) layout.Dimensions {
	ts.update(gtx)

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
	event.Op(gtx.Ops, ts.chart)
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

	if ts.chart.hoveredPoint != nil {
		spacing := gtx.Dp(utils.CommonSpacing)
		hoverPosition := ts.chart.hoverPosition
		hoverX := ts.chart.hoveredPointX
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
			Y: float32(size.Y),
		})

		paint.FillShape(gtx.Ops, utils.ChartLineColor, clip.Stroke{
			Path:  line.End(),
			Width: 2,
		}.Op())

		// Draw tooltip
		cgtx := gtx
		cgtx.Constraints.Min = image.Point{}

		tooltipMacro := op.Record(gtx.Ops)
		tooltipDims := ts.tooltip(*ts.chart.hoveredPoint)(cgtx)
		tooltipCall := tooltipMacro.Stop()

		xOffset := spacing * 2
		if hoverPosition.X+spacing*3+tooltipDims.Size.X > size.X {
			xOffset = -(tooltipDims.Size.X + spacing*2)
		}
		tooltipX := hoverPosition.X + xOffset

		yOffset := min(spacing*2, size.Y-(hoverPosition.Y+tooltipDims.Size.Y+spacing))
		tooltipY := hoverPosition.Y + yOffset

		trans := op.Offset(image.Point{
			X: tooltipX,
			Y: tooltipY,
		}).Push(gtx.Ops)
		tooltipCall.Add(gtx.Ops)
		trans.Pop()
	}

	return layout.Dimensions{
		Size: size,
	}
}
