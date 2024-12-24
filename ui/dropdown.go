package ui

import (
	"fmt"
	"gioui.org/layout"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
)

type Dropdown struct {
	Value    fmt.Stringer
	Options  []fmt.Stringer
	TextSize unit.Sp
	Inset    layout.Inset

	open     bool
	button   *widget.Clickable
	itemList *widget.List
}

func NewDropdown(theme *material.Theme, first fmt.Stringer, other ...fmt.Stringer) *Dropdown {
	return &Dropdown{
		Value:    first,
		Options:  append([]fmt.Stringer{first}, other...),
		TextSize: theme.TextSize,
		Inset: layout.Inset{
			Top: 10, Bottom: 10,
			Left: 12, Right: 12,
		},
	}
}
