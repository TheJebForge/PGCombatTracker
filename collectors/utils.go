package collectors

import (
	"PGCombatTracker/abstract"
	"PGCombatTracker/ui/components"
	"PGCombatTracker/utils"
	"gioui.org/layout"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"strings"
	"unicode"
)

func defaultDropdownStyle(state abstract.LayeredState, dropdown *components.Dropdown) components.DropdownStyle {
	style := components.StyleDropdown(state.Theme(), state.ModalLayer(), dropdown)

	style.TextSize = 12
	style.Inset = layout.UniformInset(3)
	style.DialogTextSize = 12
	style.DialogMinWidth = 250

	return style
}

func defaultCheckboxStyle(state abstract.GlobalState, bool *widget.Bool, label string) material.CheckBoxStyle {
	style := material.CheckBox(state.Theme(), bool, label)
	style.TextSize = 12
	style.Size = 18
	return style
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
