package utils

import (
	"gioui.org/text"
	"gioui.org/widget/material"
	"image/color"
)

func WithColor(style material.LabelStyle, c color.NRGBA) material.LabelStyle {
	style.Color = c
	return style
}

func WithAlignment(style material.LabelStyle, alignment text.Alignment) material.LabelStyle {
	style.Alignment = alignment
	return style
}
