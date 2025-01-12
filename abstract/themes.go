package abstract

import (
	"PGCombatTracker/utils"
	"fmt"
	"gioui.org/widget/material"
	"image/color"
)

type PGCTTheme struct {
	LesserContrastBg    color.NRGBA
	LessContrastBg      color.NRGBA
	BG                  color.NRGBA
	TextColor           color.NRGBA
	ContrastBG          color.NRGBA
	ContrastTextColor   color.NRGBA
	HalfBG              color.NRGBA
	SecondBG            color.NRGBA
	GrayText            color.NRGBA
	RedText             color.NRGBA
	ChartLineColor      color.NRGBA
	ChartSelectionColor color.NRGBA
	RandomColorBase     int
	RandomColorWidth    int
}

func (th PGCTTheme) Apply(theme *material.Theme) {
	utils.LesserContrastBg = th.LesserContrastBg
	utils.LessContrastBg = th.LessContrastBg
	utils.BG = th.BG
	theme.Bg = th.BG
	theme.Fg = th.TextColor
	theme.ContrastBg = th.ContrastBG
	theme.ContrastFg = th.ContrastTextColor
	utils.HalfBG = th.HalfBG
	utils.SecondBG = th.SecondBG
	utils.GrayText = th.GrayText
	utils.RedText = th.RedText
	utils.RandomColorBase = th.RandomColorBase
	utils.RandomColorWidth = th.RandomColorWidth
	utils.ChartLineColor = th.ChartLineColor
	utils.ChartSelectionColor = th.ChartSelectionColor
}

type PGCTThemeSelection uint8

const (
	PGCTGrayTheme PGCTThemeSelection = iota
	PGCTLightTheme
)

func PGCTThemeOptions() []fmt.Stringer {
	return []fmt.Stringer{
		PGCTGrayTheme,
		PGCTLightTheme,
	}
}

func (s PGCTThemeSelection) Theme() PGCTTheme {
	switch s {
	case PGCTGrayTheme:
		return PGCTTheme{
			LesserContrastBg:    utils.RGB(0x373737),
			LessContrastBg:      utils.RGB(0x464646),
			BG:                  utils.RGB(0x191919),
			TextColor:           utils.RGB(0xffffff),
			ContrastBG:          utils.RGB(0x646464),
			ContrastTextColor:   utils.RGB(0xffffff),
			HalfBG:              utils.RGB(0x202020),
			SecondBG:            utils.RGB(0x282828),
			GrayText:            utils.RGB(0x8c8c8c),
			RedText:             utils.RGB(0xff3030),
			RandomColorBase:     60,
			RandomColorWidth:    50,
			ChartLineColor:      utils.ARGB(0xffffff14),
			ChartSelectionColor: utils.ARGB(0x00ffff32),
		}
	case PGCTLightTheme:
		return PGCTTheme{
			LesserContrastBg:    utils.RGB(0xcdcdcd),
			LessContrastBg:      utils.RGB(0xbdbdbd),
			BG:                  utils.RGB(0xffffff),
			TextColor:           utils.RGB(0x050505),
			ContrastBG:          utils.RGB(0xaaaaaa),
			ContrastTextColor:   utils.RGB(0x111111),
			HalfBG:              utils.RGB(0xeeeeee),
			SecondBG:            utils.RGB(0xdddddd),
			GrayText:            utils.RGB(0x666666),
			RedText:             utils.RGB(0xff3030),
			RandomColorBase:     160,
			RandomColorWidth:    50,
			ChartLineColor:      utils.ARGB(0x14000000),
			ChartSelectionColor: utils.ARGB(0x99006666),
		}
	}

	return PGCTTheme{}
}

func (s PGCTThemeSelection) String() string {
	switch s {
	case PGCTGrayTheme:
		return "Gray"
	case PGCTLightTheme:
		return "Light"
	}

	return ""
}
