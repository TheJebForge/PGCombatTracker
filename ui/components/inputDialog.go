package components

import (
	"PGCombatTracker/utils"
	"gioui.org/layout"
	"gioui.org/text"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"image"
)

func NewInputDialog(theme *material.Theme, title, prompt, hint string, callback func(string)) *InputDialog {
	return &InputDialog{
		Title:         title,
		Prompt:        prompt,
		Hint:          hint,
		Callback:      callback,
		TextSize:      theme.TextSize,
		MinWidth:      200,
		theme:         theme,
		editor:        &widget.Editor{},
		confirmButton: &widget.Clickable{},
	}
}

type InputDialog struct {
	Title    string
	Prompt   string
	Hint     string
	Callback func(string)
	TextSize unit.Sp
	MinWidth unit.Dp

	theme         *material.Theme
	modalLayer    *ModalLayer
	editor        *widget.Editor
	confirmButton *widget.Clickable
}

func (i *InputDialog) Open(layer *ModalLayer) {
	layer.CurrentModal = i
	i.modalLayer = layer
}

func (i *InputDialog) surface(inner layout.Widget) layout.Widget {
	return func(gtx layout.Context) layout.Dimensions {
		return utils.LayoutMinimalSize(
			image.Point{X: gtx.Dp(i.MinWidth)},
			layout.Background{}.Layout(
				gtx,
				utils.MakeRoundedBG(10, utils.SecondBG),
				func(gtx layout.Context) layout.Dimensions {
					return layout.UniformInset(utils.CommonSpacing*2).Layout(
						gtx,
						inner,
					)
				},
			),
		)
	}
}

func (i *InputDialog) textEditor(theme *material.Theme, editor *widget.Editor, hint string) layout.Widget {
	return func(gtx layout.Context) layout.Dimensions {
		return layout.Background{}.Layout(
			gtx,
			utils.MakeRoundedBG(10, utils.LessContrastBg),
			func(gtx layout.Context) layout.Dimensions {
				return layout.UniformInset(utils.CommonSpacing*2).Layout(
					gtx,
					func(gtx layout.Context) layout.Dimensions {
						style := material.Editor(theme, editor, hint)
						style.TextSize = i.TextSize
						style.HintColor = utils.GrayText
						return style.Layout(gtx)
					},
				)
			},
		)
	}
}

func (i *InputDialog) ModalLayout(gtx layout.Context) layout.Dimensions {
	return i.surface(func(gtx layout.Context) layout.Dimensions {
		if i.confirmButton.Clicked(gtx) {
			i.Callback(i.editor.Text())
			i.Close(i.modalLayer)
		}

		cgtx := gtx
		cgtx.Constraints.Min.X = gtx.Dp(i.MinWidth)

		return layout.Flex{
			Axis:      layout.Vertical,
			Alignment: layout.Middle,
		}.Layout(
			cgtx,
			layout.Rigid(utils.WithAlignment(material.Label(i.theme, i.TextSize, i.Title), text.Middle).Layout),
			utils.FlexSpacerH(utils.CommonSpacing),
			layout.Rigid(utils.MakeVerticalSeparator(2, i.MinWidth)),
			utils.FlexSpacerH(utils.CommonSpacing),
			layout.Rigid(utils.WithAlignment(material.Label(i.theme, i.TextSize, i.Prompt), text.Middle).Layout),
			utils.FlexSpacerH(utils.CommonSpacing),
			layout.Rigid(i.textEditor(i.theme, i.editor, i.Hint)),
			utils.FlexSpacerH(utils.CommonSpacing),
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				style := material.Button(i.theme, i.confirmButton, "Confirm")
				style.TextSize = i.TextSize
				return style.Layout(gtx)
			}),
		)
	})(gtx)
}

func (i *InputDialog) Close(layer *ModalLayer) {
	if layer != nil {
		layer.CurrentModal = nil
	}
	i.modalLayer = nil
}
