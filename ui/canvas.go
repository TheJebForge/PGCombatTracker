package ui

import (
	"gioui.org/layout"
	"gioui.org/op"
	"image"
)

type Canvas struct {
	ExpandHorizontal bool
	ExpandVertical   bool
	MinSize          image.Point
}

type CanvasItem struct {
	Anchor layout.Direction
	Offset image.Point
	Widget layout.Widget

	call op.CallOp
	dims layout.Dimensions
}

func (canvas Canvas) Layout(gtx layout.Context, children ...CanvasItem) layout.Dimensions {
	var maxSize image.Point

	if canvas.ExpandHorizontal {
		maxSize.X = gtx.Constraints.Max.X
	}

	if canvas.ExpandVertical {
		maxSize.Y = gtx.Constraints.Max.Y
	}

	if canvas.MinSize.X > 0 {
		maxSize.X = canvas.MinSize.X
	}

	if canvas.MinSize.Y > 0 {
		maxSize.Y = canvas.MinSize.Y
	}

	cgtx := gtx
	for i, c := range children {
		cgtx.Constraints.Min = image.Point{}

		switch c.Anchor {
		case layout.N, layout.S:
			cgtx.Constraints.Min.X = maxSize.X
		case layout.E, layout.W:
			cgtx.Constraints.Min.Y = maxSize.Y
		case layout.Center:
			cgtx.Constraints.Min = maxSize
		default:
		}

		macro := op.Record(gtx.Ops)
		dims := c.Widget(cgtx)
		call := macro.Stop()

		switch c.Anchor {
		case layout.NW:
			if w := dims.Size.X + c.Offset.X; w > maxSize.X {
				maxSize.X = w
			}

			if h := dims.Size.Y + c.Offset.Y; h > maxSize.Y {
				maxSize.Y = h
			}
		case layout.N:
			if w := dims.Size.X; w > maxSize.X {
				maxSize.X = w
			}

			if h := dims.Size.Y + c.Offset.Y; h > maxSize.Y {
				maxSize.Y = h
			}
		case layout.NE:
			if w := maxSize.X + c.Offset.X; w > maxSize.X {
				maxSize.X = w
			}

			if h := dims.Size.Y + c.Offset.Y; h > maxSize.Y {
				maxSize.Y = h
			}
		case layout.E:
			if w := maxSize.X + c.Offset.X; w > maxSize.X {
				maxSize.X = w
			}

			if h := dims.Size.Y; h > maxSize.Y {
				maxSize.Y = h
			}
		case layout.SE:
			if w := maxSize.X + c.Offset.X; w > maxSize.X {
				maxSize.X = w
			}

			if h := maxSize.Y + c.Offset.Y; h > maxSize.Y {
				maxSize.Y = h
			}
		case layout.S:
			if w := dims.Size.X; w > maxSize.X {
				maxSize.X = w
			}

			if h := maxSize.Y + c.Offset.Y; h > maxSize.Y {
				maxSize.Y = h
			}
		case layout.SW:
			if w := dims.Size.X + c.Offset.X; w > maxSize.X {
				maxSize.X = w
			}

			if h := maxSize.Y + c.Offset.Y; h > maxSize.Y {
				maxSize.Y = h
			}
		case layout.W:
			if w := dims.Size.X + c.Offset.X; w > maxSize.X {
				maxSize.X = w
			}

			if h := dims.Size.Y; h > maxSize.Y {
				maxSize.Y = h
			}
		case layout.Center:
			if w := dims.Size.X; w > maxSize.X {
				maxSize.X = w
			}

			if h := dims.Size.Y; h > maxSize.Y {
				maxSize.Y = h
			}
		}

		children[i].call = call
		children[i].dims = dims
	}

	maxSize = gtx.Constraints.Constrain(maxSize)
	for _, c := range children {
		var offset image.Point

		switch c.Anchor {
		case layout.NW:
			offset = c.Offset
		case layout.N:
			offset.X = maxSize.X/2 - c.dims.Size.X/2
			offset.Y = c.Offset.Y
		case layout.NE:
			offset.X = maxSize.X - c.dims.Size.X + c.Offset.X
			offset.Y = c.Offset.Y
		case layout.E:
			offset.X = maxSize.X - c.dims.Size.X + c.Offset.X
			offset.Y = maxSize.Y/2 - c.dims.Size.Y/2
		case layout.SE:
			offset.X = maxSize.X - c.dims.Size.X + c.Offset.X
			offset.Y = maxSize.Y - c.dims.Size.Y + c.Offset.Y
		case layout.S:
			offset.X = maxSize.X/2 - c.dims.Size.X/2
			offset.Y = maxSize.Y - c.dims.Size.Y + c.Offset.Y
		case layout.SW:
			offset.X = c.Offset.X
			offset.Y = maxSize.Y - c.dims.Size.Y + c.Offset.Y
		case layout.W:
			offset.X = c.Offset.X
			offset.Y = maxSize.Y/2 - c.dims.Size.Y/2
		case layout.Center:
			offset.X = maxSize.X/2 - c.dims.Size.X/2
			offset.Y = maxSize.Y/2 - c.dims.Size.Y/2
		}

		trans := op.Offset(offset).Push(gtx.Ops)
		c.call.Add(gtx.Ops)
		trans.Pop()
	}

	return layout.Dimensions{
		Size: maxSize,
	}
}
