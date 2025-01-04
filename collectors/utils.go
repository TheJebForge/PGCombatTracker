package collectors

import (
	"PGCombatTracker/abstract"
	"PGCombatTracker/ui/components"
	"PGCombatTracker/utils"
	"fmt"
	"gioui.org/layout"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"image"
	"strings"
	"unicode"
)

func defaultDropdownStyle(state abstract.LayeredState, dropdown *components.Dropdown) components.DropdownStyle {
	style := components.StyleDropdown(state.Theme(), state.ModalLayer(), dropdown)

	style.TextSize = 12
	style.Inset = layout.UniformInset(3)
	style.DialogTextSize = 12
	style.MaxWidth = 150
	style.DialogMinWidth = 250

	return style
}

func defaultCheckboxStyle(state abstract.GlobalState, bool *widget.Bool, label string) material.CheckBoxStyle {
	style := material.CheckBox(state.Theme(), bool, label)
	style.TextSize = 12
	style.Size = 18
	return style
}

func defaultLabelStyle(state abstract.LayeredState, text string) material.LabelStyle {
	return material.Label(state.Theme(), 12, text)
}

func topBarSurface(inner layout.Widget) layout.Widget {
	return func(gtx layout.Context) layout.Dimensions {
		return layout.Background{}.Layout(
			gtx,
			utils.MakeColoredBG(utils.HalfBG),
			func(gtx layout.Context) layout.Dimensions {
				return layout.UniformInset(utils.CommonSpacing).Layout(gtx, inner)
			},
		)
	}
}

func drawUniversalStatsText(
	state abstract.LayeredState,
	sideText utils.LongFormatable,
	background layout.Widget, amount int,
	name, amountFormat string,
	size unit.Dp,
	long bool,
) layout.Widget {
	return func(gtx layout.Context) layout.Dimensions {
		return components.Canvas{
			ExpandHorizontal: true,
			MinSize: image.Point{
				Y: gtx.Dp(size + utils.CommonSpacing),
			},
		}.Layout(
			gtx,
			components.CanvasItem{
				Anchor: layout.N,
				Widget: func(gtx layout.Context) layout.Dimensions {
					cgtx := gtx
					cgtx.Constraints.Min.X = cgtx.Constraints.Max.X

					return background(cgtx)
				},
			},
			components.CanvasItem{
				Anchor: layout.W,
				Offset: image.Point{
					X: gtx.Dp(utils.CommonSpacing),
					Y: gtx.Dp(-2.5),
				},
				Widget: func(gtx layout.Context) layout.Dimensions {
					if amount == 0 {
						return material.Label(state.Theme(), 12, name).Layout(gtx)
					} else {
						return layout.Flex{
							Axis: layout.Vertical,
						}.Layout(
							gtx,
							layout.Rigid(material.Label(state.Theme(), 12, name).Layout),
							layout.Rigid(material.Label(state.Theme(), 12, fmt.Sprintf(amountFormat,
								amount)).Layout),
						)
					}
				},
			},
			components.CanvasItem{
				Anchor: layout.E,
				Offset: image.Point{
					X: gtx.Dp(-utils.CommonSpacing),
					Y: gtx.Dp(-2.5),
				},
				Widget: func(gtx layout.Context) layout.Dimensions {
					return material.Label(state.Theme(), 12, sideText.StringCL(long)).Layout(gtx)
				},
			},
		)
	}
}

func drawUniversalBar(
	state abstract.LayeredState,
	sideText utils.LongFormatable,
	value, max, amount int,
	name, amountFormat string,
	size unit.Dp,
	long bool,
) layout.Widget {
	return drawUniversalStatsText(
		state,
		sideText,
		func(gtx layout.Context) layout.Dimensions {
			var progress = float64(value) / float64(max)

			return components.BarWidget(components.StringToColor(name), size, progress)(gtx)
		}, amount,
		name, amountFormat,
		size,
		long,
	)
}

// SplitOffId split off any numbers or ids on the end of the string
func SplitOffId(name string) string {
	parts := strings.Split(name, " ")

	if len(parts) <= 0 {
		return ""
	}

	lastPart := parts[len(parts)-1]

	for _, c := range []rune(lastPart) {
		if !unicode.IsDigit(c) && c != '#' {
			return name
		}
	}

	return strings.Join(parts[:len(parts)-1], " ")
}

type subjectChoice string

func (s subjectChoice) String() string {
	if s == "" {
		return "All"
	}
	return string(s)
}
