package components

import (
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/widget"
	"image"
	"image/color"
)

type Modal interface {
	ModalLayout(gtx layout.Context) layout.Dimensions
	Close(layer *ModalLayer)
}

func NewModalLayer() *ModalLayer {
	return &ModalLayer{
		outerArea: &widget.Clickable{},
		innerArea: &widget.Clickable{},
	}
}

type ModalLayer struct {
	CurrentModal Modal
	outerArea    *widget.Clickable
	innerArea    *widget.Clickable
}

func (ml *ModalLayer) Overlay(bottom layout.Widget) layout.Widget {
	return func(gtx layout.Context) layout.Dimensions {
		if ml.outerArea.Clicked(gtx) {
			if ml.CurrentModal != nil {
				ml.CurrentModal.Close(ml)
			}
		}

		return layout.Stack{}.Layout(
			gtx,
			layout.Expanded(bottom),
			layout.Stacked(func(gtx layout.Context) layout.Dimensions {
				if modal := ml.CurrentModal; modal != nil {
					// Layout an outer area, so we can catch clicks outside of the modal
					// and prevent clicks from going to bottom layer
					ml.outerArea.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
						return layout.Dimensions{
							Size: gtx.Constraints.Max,
						}
					})

					// Layout the modal into a macro, so we can record
					// dimensions and keep it for later
					cgtx := gtx
					cgtx.Constraints.Min = image.Point{}

					macro := op.Record(gtx.Ops)
					dims := modal.ModalLayout(gtx)
					call := macro.Stop()

					// Layout inner area into a macro based on modal dimensions
					areaMacro := op.Record(gtx.Ops)

					ml.innerArea.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
						return layout.Dimensions{
							Size: dims.Size,
						}
					})

					areaCall := areaMacro.Stop()

					// Calculate top left location based on received dimensions
					location := image.Point{
						X: cgtx.Constraints.Max.X/2 - dims.Size.X/2,
						Y: cgtx.Constraints.Max.Y/2 - dims.Size.Y/2,
					}

					// Draw semi-transparent background to bring visual focus to the modal
					background := clip.Rect{Max: gtx.Constraints.Max}.Push(gtx.Ops)
					paint.Fill(gtx.Ops, color.NRGBA{A: 150})
					background.Pop()

					// Offset everything by previously calculated location
					trans := op.Offset(location).Push(gtx.Ops)

					// Render the inner area
					areaCall.Add(gtx.Ops)

					// Render the modal
					call.Add(gtx.Ops)

					// Pop transformation stack
					trans.Pop()
				}

				return layout.Dimensions{
					Size: gtx.Constraints.Max,
				}
			}),
		)
	}

}
