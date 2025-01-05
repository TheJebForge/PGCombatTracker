package components

import (
	"cmp"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/unit"
	"github.com/samber/lo"
	"image"
	"slices"
)

// HorizontalWrap Container that layouts all the children horizontally and wraps them to next line if they don't fit
type HorizontalWrap struct {
	Alignment   layout.Alignment
	Spacing     unit.Dp
	LineSpacing unit.Dp
}

type horizontalLine struct {
	calls       []op.CallOp
	dims        []layout.Dimensions
	maxBaseline int
	size        image.Point
}

func (hw HorizontalWrap) Layout(gtx layout.Context, children ...layout.Widget) layout.Dimensions {
	spacing := gtx.Dp(hw.Spacing)
	lineSpacing := gtx.Dp(hw.LineSpacing)
	maxWidth := gtx.Constraints.Max.X

	var lines []horizontalLine
	var currentLine horizontalLine

	cgtx := gtx
	cgtx.Constraints.Min = image.Point{}

	for _, c := range children {
		macro := op.Record(gtx.Ops)
		dim := c(cgtx)
		call := macro.Stop()

		currentSpacing := spacing
		if len(currentLine.calls) <= 0 {
			currentSpacing = 0
		}

		baseline := dim.Size.Y - dim.Baseline

		if currentLine.size.X+currentSpacing+dim.Size.X < maxWidth || len(currentLine.calls) <= 0 {
			currentLine.size.X += currentSpacing + dim.Size.X
			currentLine.size.Y = max(currentLine.size.Y, dim.Size.Y)

			currentLine.dims = append(currentLine.dims, dim)
			currentLine.calls = append(currentLine.calls, call)

			currentLine.maxBaseline = max(currentLine.maxBaseline, baseline)
		} else {
			lines = append(lines, currentLine)

			currentLine = horizontalLine{
				size: dim.Size,
				dims: []layout.Dimensions{
					dim,
				},
				calls: []op.CallOp{
					call,
				},
				maxBaseline: baseline,
			}
		}
	}

	lines = append(lines, currentLine)

	maxSize := image.Point{
		X: slices.MaxFunc(lines, func(a, b horizontalLine) int {
			return cmp.Compare(a.size.X, b.size.X)
		}).size.X,
		Y: lo.SumBy(lines, func(item horizontalLine) int {
			return item.size.Y
		}) + ((len(lines) - 1) * lineSpacing),
	}
	maxSize = gtx.Constraints.Constrain(maxSize)

	currentY := 0
	for _, line := range lines {
		currentX := 0

		for i, call := range line.calls {
			dim := line.dims[i]
			baseline := dim.Size.Y - dim.Baseline

			offset := image.Point{
				X: currentX,
				Y: currentY,
			}

			switch hw.Alignment {
			case layout.Middle:
				offset.Y += line.size.Y/2 - dim.Size.Y/2
			case layout.End:
				offset.Y += line.size.Y - dim.Size.Y
			case layout.Baseline:
				offset.Y += line.maxBaseline - baseline
			default:
			}

			trans := op.Offset(offset).Push(gtx.Ops)
			call.Add(gtx.Ops)
			trans.Pop()

			currentX += spacing + dim.Size.X
		}

		currentY += line.size.Y + lineSpacing
	}

	allBaseline := 0
	if len(lines) == 1 {
		allBaseline = maxSize.Y - lines[0].maxBaseline
	}

	return layout.Dimensions{
		Size:     maxSize,
		Baseline: allBaseline,
	}
}
