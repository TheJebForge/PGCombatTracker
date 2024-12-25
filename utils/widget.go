package utils

import "gioui.org/widget"

func MakeClickableArray(amount int) []*widget.Clickable {
	clickables := make([]*widget.Clickable, amount)

	for i := 0; i < amount; i++ {
		clickables[i] = &widget.Clickable{}
	}

	return clickables
}
