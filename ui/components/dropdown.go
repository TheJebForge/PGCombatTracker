package components

import (
	"PGCombatTracker/utils"
	"fmt"
	"gioui.org/layout"
	"gioui.org/text"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"golang.org/x/exp/shiny/materialdesign/icons"
	"image"
)

type Dropdown struct {
	Title   string
	Value   fmt.Stringer
	options []fmt.Stringer

	changed           bool
	valueIndex        int
	currentModalLayer *ModalLayer // nil if not open
	button            *widget.Clickable
	itemList          *widget.List
	itemButtons       []*widget.Clickable

	downIcon *widget.Icon
	upIcon   *widget.Icon
}

func NewDropdown(title string, first fmt.Stringer, other ...fmt.Stringer) (*Dropdown, error) {
	options := append([]fmt.Stringer{first}, other...)

	downIcon, err := widget.NewIcon(icons.NavigationArrowDropDown)
	if err != nil {
		return nil, err
	}

	upIcon, err := widget.NewIcon(icons.NavigationArrowDropUp)
	if err != nil {
		return nil, err
	}

	return &Dropdown{
		Title:   title,
		Value:   first,
		options: options,

		button: &widget.Clickable{},
		itemList: &widget.List{
			List: layout.List{
				Axis: layout.Vertical,
			},
		},
		itemButtons: utils.MakeClickableArray(len(options)),

		downIcon: downIcon,
		upIcon:   upIcon,
	}, nil
}

func (d *Dropdown) Options() []fmt.Stringer {
	return d.options
}

func (d *Dropdown) SetOptions(options []fmt.Stringer) {
	d.options = options
	d.itemButtons = utils.MakeClickableArray(len(options))
}

type DropdownStyle struct {
	TextSize       unit.Sp
	MaxWidth       unit.Dp
	DialogTextSize unit.Sp
	Inset          layout.Inset
	DialogMinWidth unit.Dp
	NoLabel        bool

	modalLayer *ModalLayer
	dropdown   *Dropdown
	theme      *material.Theme
}

func StyleDropdown(theme *material.Theme, modalLayer *ModalLayer, dropdown *Dropdown) DropdownStyle {
	return DropdownStyle{
		TextSize:       theme.TextSize,
		MaxWidth:       200,
		DialogTextSize: theme.TextSize,
		Inset: layout.Inset{
			Top: 10, Bottom: 10,
			Left: 12, Right: 12,
		},
		DialogMinWidth: 200,

		modalLayer: modalLayer,
		theme:      theme,
		dropdown:   dropdown,
	}
}

func (d DropdownStyle) dropdownButton() layout.Widget {
	return func(gtx layout.Context) layout.Dimensions {
		return material.ButtonLayout(d.theme, d.dropdown.button).Layout(
			gtx,
			func(gtx layout.Context) layout.Dimensions {
				return d.Inset.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
					return layout.Flex{
						Axis:      layout.Horizontal,
						Alignment: layout.Middle,
					}.Layout(
						gtx,
						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							cgtx := gtx
							cgtx.Constraints.Max.X = gtx.Dp(d.MaxWidth)
							return material.Label(d.theme, d.TextSize, d.dropdown.Value.String()).Layout(cgtx)
						}),
						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							cgtx := gtx
							cgtx.Constraints.Min.X = gtx.Dp(unit.Dp(d.TextSize * 1.25))
							if d.dropdown.currentModalLayer != nil {
								return d.dropdown.upIcon.Layout(cgtx, d.theme.Fg)
							} else {
								return d.dropdown.downIcon.Layout(cgtx, d.theme.Fg)
							}
						}),
					)
				})
			},
		)
	}
}

func (d DropdownStyle) Layout(gtx layout.Context) layout.Dimensions {
	if d.dropdown.button.Clicked(gtx) {
		d.Open(d.modalLayer)
	}

	if d.NoLabel {
		return d.dropdownButton()(gtx)
	}

	return layout.Flex{
		Axis:      layout.Horizontal,
		Alignment: layout.Middle,
	}.Layout(
		gtx,
		layout.Rigid(material.Label(d.theme, d.TextSize, d.dropdown.Title).Layout),
		utils.FlexSpacerW(utils.CommonSpacing),
		layout.Rigid(d.dropdownButton()),
	)
}

func (d DropdownStyle) Open(modalLayer *ModalLayer) {
	modalLayer.CurrentModal = d
	d.dropdown.currentModalLayer = modalLayer
}

func (d *Dropdown) SelectItem(index int) {
	d.currentModalLayer.CurrentModal = nil
	d.currentModalLayer = nil

	d.valueIndex = index
	d.Value = d.options[index]
	d.changed = true
}

func (d DropdownStyle) surface(inner layout.Widget) layout.Widget {
	return func(gtx layout.Context) layout.Dimensions {
		return utils.LayoutMinimalSize(
			image.Point{X: gtx.Dp(d.DialogMinWidth)},
			layout.Background{}.Layout(
				gtx,
				utils.MakeRoundedBG(10, utils.SecondBG),
				func(gtx layout.Context) layout.Dimensions {
					return layout.Inset{
						Top:    utils.CommonSpacing * 2,
						Right:  utils.CommonSpacing * 1,
						Bottom: utils.CommonSpacing * 2,
						Left:   utils.CommonSpacing * 2,
					}.Layout(
						gtx,
						inner,
					)
				},
			),
		)
	}
}

func (d DropdownStyle) ModalLayout(gtx layout.Context) layout.Dimensions {
	return d.surface(func(gtx layout.Context) layout.Dimensions {
		cgtx := gtx
		cgtx.Constraints.Min.X = gtx.Dp(d.DialogMinWidth)

		return layout.Flex{
			Axis: layout.Vertical,
		}.Layout(
			cgtx,
			layout.Rigid(utils.WithAlignment(material.Label(d.theme, d.DialogTextSize, d.dropdown.Title), text.Middle).Layout),
			utils.FlexSpacerH(utils.CommonSpacing),
			layout.Rigid(utils.MakeVerticalSeparator(2, d.DialogMinWidth)),
			utils.FlexSpacerH(utils.CommonSpacing),
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				amount := len(d.dropdown.options)

				return material.List(d.theme, d.dropdown.itemList).Layout(
					gtx,
					amount,
					func(gtx layout.Context, index int) layout.Dimensions {
						item := d.dropdown.options[index]
						button := d.dropdown.itemButtons[index]

						if button.Clicked(gtx) {
							d.dropdown.SelectItem(index)
						}

						cgtx := gtx
						cgtx.Constraints.Min.X = gtx.Constraints.Max.X

						style := material.Button(d.theme, button, item.String())
						style.TextSize = d.DialogTextSize

						if d.dropdown.valueIndex != index {
							style.Background = utils.SecondBG
						}

						return style.Layout(gtx)
					},
				)
			}),
		)
	})(gtx)
}

func (d DropdownStyle) Close(layer *ModalLayer) {
	layer.CurrentModal = nil
	d.dropdown.currentModalLayer = nil
}

func (d *Dropdown) Changed() bool {
	if d.changed {
		d.changed = false
		return true
	}

	return false
}
