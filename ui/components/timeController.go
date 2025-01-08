package components

import (
	"PGCombatTracker/utils"
	"gioui.org/gesture"
	"gioui.org/io/event"
	"gioui.org/io/key"
	"gioui.org/io/pointer"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"golang.org/x/exp/shiny/materialdesign/icons"
	"image"
	"image/color"
	"math"
	"time"
)

type TimeMode uint8

const (
	OverviewTimeMode TimeMode = iota
	RealTimeMode
)

func NewTimeController(baseChart *TimeBasedChart) (*TimeController, error) {
	overviewIcon, err := widget.NewIcon(icons.ActionTimeline)
	if err != nil {
		return nil, err
	}

	realTimeIcon, err := widget.NewIcon(icons.ActionHistory)
	if err != nil {
		return nil, err
	}

	defaultOptions := []RealTimeOption{
		{
			Label:   "30s",
			Seconds: 30,
		},
		{
			Label:   "1m",
			Seconds: 60,
		},
		{
			Label:   "3m",
			Seconds: 180,
		},
		{
			Label:   "5m",
			Seconds: 300,
		},
		{
			Label:   "10m",
			Seconds: 600,
		},
		{
			Label:   "15m",
			Seconds: 900,
		},
	}

	controller := &TimeController{
		BaseChart:       baseChart,
		overviewIcon:    overviewIcon,
		realTimeIcon:    realTimeIcon,
		overviewFramer:  newOverviewTimeFramer(),
		modeButton:      &widget.Clickable{},
		LookBackSeconds: 300,
		realTimeOptions: defaultOptions,
		realTimeButtons: utils.MakeClickableArray(len(defaultOptions)),
	}
	controller.overviewFramer.ctrl = controller

	return controller, nil
}

type RealTimeOption struct {
	Label   string
	Seconds int
}

type TimeController struct {
	BaseChart      *TimeBasedChart
	FullTimeFrame  TimeFrame
	FullValueRange DataRange

	CurrentTimeFrame TimeFrame

	CurrentMode TimeMode

	LookBackSeconds int

	overviewIcon   *widget.Icon
	realTimeIcon   *widget.Icon
	overviewFramer *overviewTimeFramer
	modeButton     *widget.Clickable

	realTimeOptions []RealTimeOption
	realTimeButtons []*widget.Clickable
}

func (tc *TimeController) Add(point TimePoint) {
	tc.BaseChart.Add(point)

	length := len(tc.BaseChart.DataPoints)

	if length == 1 {
		tc.FullTimeFrame = TimeFrame{
			From: point.Time,
			To:   point.Time,
		}
		tc.FullValueRange.Min = point.Value - 1
		tc.FullValueRange.Max = point.Value + 1
	} else {
		tc.FullTimeFrame = tc.FullTimeFrame.Expand(point.Time)
		tc.FullValueRange = tc.FullValueRange.Expand(point.Value)
	}

	tc.RecalculateTimeFrame()
}

func (tc *TimeController) recalculateRealTimeFrame() {
	now := time.Now()

	tc.CurrentTimeFrame.From = now.Add(time.Second * time.Duration(-tc.LookBackSeconds))
	tc.CurrentTimeFrame.To = now
}

func (tc *TimeController) RecalculateTimeFrame() {
	if tc.CurrentMode == OverviewTimeMode {
		tc.overviewFramer.recalculateControllerBounds()
	} else {
		tc.recalculateRealTimeFrame()
	}
}

func StyleTimeController(theme *material.Theme, controller *TimeController) TimeControllerStyle {
	return TimeControllerStyle{
		HandleThickness:      3,
		HandleColor:          color.NRGBA{R: 80, G: 80, B: 80, A: 255},
		SelectionColor:       color.NRGBA{G: 255, B: 255, A: 50},
		MinHeight:            40,
		OverviewChartStyle:   StyleTimeBasedChart(theme, controller.BaseChart),
		PixelUpdateThreshold: 3,
		TextSize:             12,
		theme:                theme,
		controller:           controller,
	}
}

type TimeControllerStyle struct {
	HandleThickness      unit.Dp
	HandleColor          color.NRGBA
	SelectionColor       color.NRGBA
	MinHeight            unit.Dp
	OverviewChartStyle   TimeBasedChartStyle
	PixelUpdateThreshold int
	TextSize             unit.Sp

	lastRequestTime time.Time
	theme           *material.Theme
	controller      *TimeController
}

func (tcs TimeControllerStyle) Layout(gtx layout.Context) layout.Dimensions {
	tcs.controller.overviewFramer.Update(gtx)

	if tcs.controller.modeButton.Clicked(gtx) {
		if tcs.controller.CurrentMode == OverviewTimeMode {
			tcs.controller.CurrentMode = RealTimeMode
		} else {
			tcs.controller.CurrentMode = OverviewTimeMode
			tcs.controller.overviewFramer.recalculateControllerBounds()
		}
	}

	if tcs.controller.CurrentMode == RealTimeMode && time.Now().After(tcs.lastRequestTime) {
		tcs.controller.recalculateRealTimeFrame()

		millisUntilUpdate := (tcs.controller.LookBackSeconds * 1000 * tcs.PixelUpdateThreshold) / gtx.Constraints.Max.X
		updateAt := time.Now().Add(time.Millisecond * time.Duration(millisUntilUpdate))
		tcs.lastRequestTime = updateAt
		gtx.Execute(op.InvalidateCmd{At: updateAt})
	}

	return layout.Flex{
		Axis:      layout.Horizontal,
		Alignment: layout.Middle,
	}.Layout(
		gtx,
		layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
			switch tcs.controller.CurrentMode {
			case OverviewTimeMode:
				return tcs.controller.overviewFramer.Layout(gtx, tcs)
			case RealTimeMode:
				flexItems := make([]layout.FlexChild, 0, len(tcs.controller.realTimeOptions)*2-1)

				for i, item := range tcs.controller.realTimeOptions {
					button := tcs.controller.realTimeButtons[i]

					if button.Clicked(gtx) {
						tcs.controller.LookBackSeconds = item.Seconds
						tcs.controller.recalculateRealTimeFrame()
					}

					if i != 0 {
						flexItems = append(flexItems, utils.FlexSpacerW(utils.CommonSpacing))
					}

					flexItems = append(flexItems, layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
						style := material.Button(tcs.theme, button, item.Label)
						style.TextSize = tcs.TextSize
						style.Inset = layout.UniformInset(utils.CommonSpacing)

						if tcs.controller.LookBackSeconds != item.Seconds {
							style.Background = utils.LesserContrastBg
						}

						return style.Layout(gtx)
					}))
				}

				return layout.Flex{
					Axis:      layout.Horizontal,
					Alignment: layout.Middle,
				}.Layout(gtx, flexItems...)
			}
			return layout.Dimensions{}
		}),
		utils.FlexSpacerW(utils.CommonSpacing),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			icon := tcs.controller.overviewIcon
			if tcs.controller.CurrentMode == RealTimeMode {
				icon = tcs.controller.realTimeIcon
			}

			style := material.IconButton(
				tcs.theme,
				tcs.controller.modeButton,
				icon,
				"Mode Button",
			)
			style.Inset = layout.UniformInset(utils.CommonSpacing)
			style.Size = tcs.MinHeight - utils.CommonSpacing*2
			return style.Layout(gtx)
		}),
	)
}

func newOverviewTimeFramer() *overviewTimeFramer {
	return &overviewTimeFramer{
		rightDragNormalizedPosition: 1,
	}
}

type overviewTimeFramer struct {
	startLeftDragNP  float64
	startRightDragNP float64

	leftDragNormalizedPosition  float64
	rightDragNormalizedPosition float64

	framePass              bool
	lastWidth              int
	lastDraggableThickness int

	startPosition image.Point
	outerDragging bool
	innerDragging bool
	leftDragging  bool
	rightDragging bool

	clicker gesture.Click
	dragger gesture.Drag

	ctrl *TimeController
}

func (ots *overviewTimeFramer) recalculateControllerBounds() {
	if ots.ctrl == nil {
		return
	}

	fullTimeFrame := ots.ctrl.FullTimeFrame
	timeFrameLength := fullTimeFrame.LengthSeconds()

	leftTimeOffset := ots.leftDragNormalizedPosition * timeFrameLength
	rightTimeOffset := ots.rightDragNormalizedPosition * timeFrameLength

	newLeftTime := fullTimeFrame.From.Add(time.Microsecond * time.Duration(math.Round(leftTimeOffset*1000000)))
	newRightTime := fullTimeFrame.From.Add(time.Microsecond * time.Duration(math.Round(rightTimeOffset*1000000)))

	ots.ctrl.CurrentTimeFrame = TimeFrame{
		From: newLeftTime,
		To:   newRightTime,
	}
}

func (ots *overviewTimeFramer) processPointerEvent(gtx layout.Context, ev event.Event) bool {
	if !ots.framePass {
		return false
	}

	var position image.Point
	var clicked, dragged, released bool

	switch evt := ev.(type) {
	case gesture.ClickEvent:
		if (evt.Kind == gesture.KindPress && evt.Source == pointer.Mouse) ||
			(evt.Kind == gesture.KindClick && evt.Source != pointer.Mouse) {
			position = evt.Position
			clicked = true
			gtx.Execute(key.FocusCmd{Tag: ots})
		} else {
			return false
		}
	case pointer.Event:
		position = evt.Position.Round()

		switch {
		case evt.Kind == pointer.Release && evt.Source == pointer.Mouse:
			released = true
		case evt.Kind == pointer.Drag && evt.Source == pointer.Mouse:
			dragged = true
		default:
			return false
		}
	}

	if clicked {
		ots.startPosition = position
		return false
	}

	thickness := ots.lastDraggableThickness
	halfThinkness := thickness / 2
	handleBounding := thickness + halfThinkness

	floatingInnerWidth := float64(ots.lastWidth - thickness)
	minimumGap := 3 / floatingInnerWidth

	startX := ots.startPosition.X
	boundedStartX := min(max(0, startX-halfThinkness), ots.lastWidth-thickness)
	normalizedStartX := float64(boundedStartX) / floatingInnerWidth

	boundedX := min(max(0, position.X-halfThinkness), ots.lastWidth-thickness)
	normalizedX := float64(boundedX) / floatingInnerWidth

	anyDragging := ots.innerDragging || ots.outerDragging || ots.leftDragging || ots.rightDragging

	if dragged {
		if !anyDragging {
			leftHandlePixelX := int(math.Floor(floatingInnerWidth*ots.leftDragNormalizedPosition)) + halfThinkness
			rightHandlePixelX := int(math.Ceil(floatingInnerWidth*ots.rightDragNormalizedPosition)) + halfThinkness

			switch {
			case startX > leftHandlePixelX-handleBounding && startX < leftHandlePixelX+handleBounding:
				// Left handle
				ots.leftDragging = true
			case startX > rightHandlePixelX-handleBounding && startX < rightHandlePixelX+handleBounding:
				// Right handle
				ots.rightDragging = true
			case startX > leftHandlePixelX && startX < rightHandlePixelX:
				// Inner area
				ots.innerDragging = true
				ots.startLeftDragNP = ots.leftDragNormalizedPosition
				ots.startRightDragNP = ots.rightDragNormalizedPosition
			default:
				// Outer area
				ots.outerDragging = true
			}
		}

		switch {
		case ots.leftDragging:
			ots.leftDragNormalizedPosition = min(normalizedX, ots.rightDragNormalizedPosition-minimumGap)
		case ots.rightDragging:
			ots.rightDragNormalizedPosition = max(normalizedX, ots.leftDragNormalizedPosition+minimumGap)
		case ots.innerDragging:
			leftBound := normalizedStartX - ots.startLeftDragNP
			rightBound := 1 - (ots.startRightDragNP - normalizedStartX)

			boundedX := min(max(leftBound, normalizedX), rightBound)
			difference := boundedX - normalizedStartX

			ots.leftDragNormalizedPosition = ots.startLeftDragNP + difference
			ots.rightDragNormalizedPosition = ots.startRightDragNP + difference
		case ots.outerDragging:
			if normalizedX > normalizedStartX {
				ots.leftDragNormalizedPosition = normalizedStartX
				ots.rightDragNormalizedPosition = max(normalizedStartX+minimumGap, normalizedX)
			} else {
				ots.leftDragNormalizedPosition = min(normalizedStartX-minimumGap, normalizedX)
				ots.rightDragNormalizedPosition = normalizedStartX
			}
		}

		ots.recalculateControllerBounds()
		return true
	}

	if released {
		ots.leftDragging = false
		ots.rightDragging = false
		ots.innerDragging = false
		ots.outerDragging = false

		if !anyDragging {
			leftHandlePixelX := int(math.Floor(floatingInnerWidth*ots.leftDragNormalizedPosition)) + halfThinkness
			rightHandlePixelX := int(math.Ceil(floatingInnerWidth*ots.rightDragNormalizedPosition)) + halfThinkness

			switch {
			case startX > leftHandlePixelX && startX < rightHandlePixelX:
				distanceToLeft := math.Abs(normalizedX - ots.leftDragNormalizedPosition)
				distanceToRight := math.Abs(normalizedX - ots.rightDragNormalizedPosition)

				if distanceToLeft <= distanceToRight {
					ots.leftDragNormalizedPosition = min(normalizedX, ots.rightDragNormalizedPosition-minimumGap)
				} else {
					ots.rightDragNormalizedPosition = max(normalizedX, ots.leftDragNormalizedPosition+minimumGap)
				}
			default:
				if normalizedX < ots.leftDragNormalizedPosition {
					ots.leftDragNormalizedPosition = min(normalizedX, ots.rightDragNormalizedPosition-minimumGap)
				} else {
					ots.rightDragNormalizedPosition = max(normalizedX, ots.leftDragNormalizedPosition+minimumGap)
				}
			}

			gtx.Execute(op.InvalidateCmd{})
			ots.recalculateControllerBounds()
			return true
		}
	}

	return false
}

func (ots *overviewTimeFramer) Update(gtx layout.Context) bool {
	for {
		ev, ok := ots.clicker.Update(gtx.Source)
		if !ok {
			break
		}
		if ots.processPointerEvent(gtx, ev) {
			return true
		}
	}

	for {
		ev, ok := ots.dragger.Update(gtx.Metric, gtx.Source, gesture.Both)
		if !ok {
			break
		}
		if ots.processPointerEvent(gtx, ev) {
			return true
		}
	}

	return false
}

func (ots *overviewTimeFramer) Layout(gtx layout.Context, style TimeControllerStyle) layout.Dimensions {
	size := image.Point{
		X: gtx.Constraints.Min.X,
		Y: max(gtx.Constraints.Min.Y, gtx.Dp(style.MinHeight)),
	}

	handleThickness := gtx.Dp(style.HandleThickness)
	halfThickness := handleThickness / 2
	spacing := gtx.Dp(utils.CommonSpacing)
	halfSpacing := spacing / 2

	// Draw the chart (1st thing to be drawn)
	chartSize := image.Point{
		X: size.X - handleThickness,
		Y: size.Y - spacing,
	}

	style.OverviewChartStyle.MinHeight = style.MinHeight - utils.CommonSpacing
	style.OverviewChartStyle.chart.DisplayTimeFrame = style.controller.FullTimeFrame
	style.OverviewChartStyle.chart.DisplayValueRange = style.controller.FullValueRange

	cgtx := gtx
	cgtx.Constraints.Min = chartSize
	cgtx.Constraints.Max = chartSize

	trans := op.Offset(image.Point{
		X: halfThickness,
		Y: halfSpacing,
	}).Push(gtx.Ops)
	style.OverviewChartStyle.Layout(cgtx)
	trans.Pop()

	// Draw invisible box to receive all events (2nd)
	eventClip := clip.Rect{Max: size}.Push(gtx.Ops)
	ots.clicker.Add(gtx.Ops)
	ots.dragger.Add(gtx.Ops)
	event.Op(gtx.Ops, ots)
	eventClip.Pop()

	// Calculate screen positions for handles
	floatingWidth := float64(size.X)
	leftPosition := int(math.Floor(ots.leftDragNormalizedPosition * floatingWidth))
	rightPosition := int(math.Ceil(ots.rightDragNormalizedPosition * floatingWidth))

	// Draw area between handles (3rd)
	areaClip := clip.UniformRRect(image.Rectangle{
		Min: image.Point{
			X: leftPosition,
			Y: halfSpacing,
		},
		Max: image.Point{
			X: rightPosition,
			Y: size.Y - halfSpacing,
		},
	}, 10).Push(gtx.Ops)
	paint.Fill(gtx.Ops, style.SelectionColor)
	areaClip.Pop()

	// Draw handles (4th and 5th)
	drawHandle := func(atX int) {
		handleClip := clip.UniformRRect(image.Rectangle{
			Min: image.Point{
				X: atX - halfThickness,
			},
			Max: image.Point{
				X: atX + handleThickness,
				Y: size.Y,
			},
		}, handleThickness).Push(gtx.Ops)
		paint.Fill(gtx.Ops, style.HandleColor)
		handleClip.Pop()
	}

	drawHandle(leftPosition)
	drawHandle(rightPosition)

	ots.framePass = true
	ots.lastWidth = size.X
	ots.lastDraggableThickness = handleThickness
	ots.ctrl = style.controller

	return layout.Dimensions{
		Size: size,
	}
}
