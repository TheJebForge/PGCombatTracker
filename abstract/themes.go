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
	PGCTNightTheme PGCTThemeSelection = iota
	PGCTGrayTheme
	PGCTLightTheme
	PGCTBluePurpleTheme
	PGCTSkyBlueTheme
	PGCTPinkTheme
	PGCTDarkRedTheme
	PGCTOliveTheme
	PGCTBrownTheme
	PGCTSkyOrangeTheme
	PGCTPurpleTheme
	PGCTSkyRedTheme
	PGCTSkyTealTheme
	PGCTBlueTheme
	PGCTGreenTheme
	PGCTPGTheme
	PGCTLightPGTheme
)

func PGCTThemeOptions() []fmt.Stringer {
	return []fmt.Stringer{
		PGCTNightTheme,
		PGCTGrayTheme,
		PGCTLightTheme,
		PGCTPGTheme,
		PGCTLightPGTheme,
		PGCTSkyBlueTheme,
		PGCTPinkTheme,
		PGCTSkyRedTheme,
		PGCTSkyOrangeTheme,
		PGCTSkyTealTheme,
		PGCTDarkRedTheme,
		PGCTBrownTheme,
		PGCTOliveTheme,
		PGCTGreenTheme,
		PGCTBlueTheme,
		PGCTBluePurpleTheme,
		PGCTPurpleTheme,
	}
}

func (s PGCTThemeSelection) Theme() PGCTTheme {
	switch s {
	case PGCTGrayTheme:
		return PGCTTheme{
			BG:                  utils.RGB(0x191919),
			HalfBG:              utils.RGB(0x202020),
			SecondBG:            utils.RGB(0x282828),
			LesserContrastBg:    utils.RGB(0x373737),
			LessContrastBg:      utils.RGB(0x464646),
			ContrastBG:          utils.RGB(0x646464),
			TextColor:           utils.RGB(0xffffff),
			ContrastTextColor:   utils.RGB(0xffffff),
			GrayText:            utils.RGB(0x8c8c8c),
			RedText:             utils.RGB(0xff3030),
			RandomColorBase:     60,
			RandomColorWidth:    50,
			ChartLineColor:      utils.ARGB(0x14ffffff),
			ChartSelectionColor: utils.ARGB(0x3200ffff),
		}
	case PGCTLightTheme:
		return PGCTTheme{
			BG:                  utils.RGB(0xffffff),
			HalfBG:              utils.RGB(0xe3e3e3),
			SecondBG:            utils.RGB(0xcecece),
			LesserContrastBg:    utils.RGB(0xbababa),
			LessContrastBg:      utils.RGB(0xadadad),
			ContrastBG:          utils.RGB(0x9a9a9a),
			TextColor:           utils.RGB(0x050505),
			ContrastTextColor:   utils.RGB(0x050505),
			GrayText:            utils.RGB(0x666666),
			RedText:             utils.RGB(0xff3030),
			RandomColorBase:     160,
			RandomColorWidth:    50,
			ChartLineColor:      utils.ARGB(0x50000000),
			ChartSelectionColor: utils.ARGB(0x55006666),
		}
	case PGCTNightTheme:
		return PGCTTheme{
			BG:                  utils.RGB(0x000000),
			HalfBG:              utils.RGB(0x101010),
			SecondBG:            utils.RGB(0x151515),
			LesserContrastBg:    utils.RGB(0x202020),
			LessContrastBg:      utils.RGB(0x303030),
			ContrastBG:          utils.RGB(0x404040),
			TextColor:           utils.RGB(0xffffff),
			ContrastTextColor:   utils.RGB(0xffffff),
			GrayText:            utils.RGB(0x8c8c8c),
			RedText:             utils.RGB(0xff3030),
			RandomColorBase:     60,
			RandomColorWidth:    50,
			ChartLineColor:      utils.ARGB(0x14ffffff),
			ChartSelectionColor: utils.ARGB(0x3200ffff),
		}
	case PGCTBluePurpleTheme:
		return PGCTTheme{
			BG:                  utils.RGB(0x0E2F47),
			HalfBG:              utils.RGB(0x1C3052),
			SecondBG:            utils.RGB(0x2A325D),
			LesserContrastBg:    utils.RGB(0x373369),
			LessContrastBg:      utils.RGB(0x453574),
			ContrastBG:          utils.RGB(0x53367F),
			TextColor:           utils.RGB(0xffffff),
			ContrastTextColor:   utils.RGB(0xffffff),
			GrayText:            utils.RGB(0x8c8c8c),
			RedText:             utils.RGB(0xff3030),
			RandomColorBase:     60,
			RandomColorWidth:    50,
			ChartLineColor:      utils.ARGB(0x14ffffff),
			ChartSelectionColor: utils.ARGB(0x3200ffff),
		}
	case PGCTSkyBlueTheme:
		return PGCTTheme{
			BG:                  utils.RGB(0xFAFAFA),
			HalfBG:              utils.RGB(0xE3EEF6),
			SecondBG:            utils.RGB(0xCCE2F2),
			LesserContrastBg:    utils.RGB(0xB5D7EF),
			LessContrastBg:      utils.RGB(0x9ECBEB),
			ContrastBG:          utils.RGB(0x87BFE7),
			TextColor:           utils.RGB(0x050505),
			ContrastTextColor:   utils.RGB(0x050505),
			GrayText:            utils.RGB(0x666666),
			RedText:             utils.RGB(0xff3030),
			RandomColorBase:     160,
			RandomColorWidth:    50,
			ChartLineColor:      utils.ARGB(0x50000000),
			ChartSelectionColor: utils.ARGB(0x55006666),
		}
	case PGCTPinkTheme:
		return PGCTTheme{
			BG:                  utils.RGB(0xFAFAFA),
			HalfBG:              utils.RGB(0xF7E8EE),
			SecondBG:            utils.RGB(0xF4D6E3),
			LesserContrastBg:    utils.RGB(0xF1C4D7),
			LessContrastBg:      utils.RGB(0xEEB2CC),
			ContrastBG:          utils.RGB(0xEBA0C0),
			TextColor:           utils.RGB(0x050505),
			ContrastTextColor:   utils.RGB(0x050505),
			GrayText:            utils.RGB(0x666666),
			RedText:             utils.RGB(0xff3030),
			RandomColorBase:     160,
			RandomColorWidth:    50,
			ChartLineColor:      utils.ARGB(0x50000000),
			ChartSelectionColor: utils.ARGB(0x55006666),
		}
	case PGCTDarkRedTheme:
		return PGCTTheme{
			BG:                  utils.RGB(0x2A1616),
			HalfBG:              utils.RGB(0x341818),
			SecondBG:            utils.RGB(0x3D1A1A),
			LesserContrastBg:    utils.RGB(0x471D1D),
			LessContrastBg:      utils.RGB(0x501F1F),
			ContrastBG:          utils.RGB(0x5A2121),
			TextColor:           utils.RGB(0xffffff),
			ContrastTextColor:   utils.RGB(0xffffff),
			GrayText:            utils.RGB(0x8c8c8c),
			RedText:             utils.RGB(0xff3030),
			RandomColorBase:     60,
			RandomColorWidth:    50,
			ChartLineColor:      utils.ARGB(0x50ffffff),
			ChartSelectionColor: utils.ARGB(0x3200ffff),
		}
	case PGCTOliveTheme:
		return PGCTTheme{
			BG:                  utils.RGB(0x172915),
			HalfBG:              utils.RGB(0x243217),
			SecondBG:            utils.RGB(0x313B1A),
			LesserContrastBg:    utils.RGB(0x3F451C),
			LessContrastBg:      utils.RGB(0x4C4E1F),
			ContrastBG:          utils.RGB(0x595721),
			TextColor:           utils.RGB(0xffffff),
			ContrastTextColor:   utils.RGB(0xffffff),
			GrayText:            utils.RGB(0x8c8c8c),
			RedText:             utils.RGB(0xff3030),
			RandomColorBase:     60,
			RandomColorWidth:    50,
			ChartLineColor:      utils.ARGB(0x14ffffff),
			ChartSelectionColor: utils.ARGB(0x3200ffff),
		}
	case PGCTBrownTheme:
		return PGCTTheme{
			BG:                  utils.RGB(0x292415),
			HalfBG:              utils.RGB(0x332A17),
			SecondBG:            utils.RGB(0x3C2F1A),
			LesserContrastBg:    utils.RGB(0x46351C),
			LessContrastBg:      utils.RGB(0x4F3A1F),
			ContrastBG:          utils.RGB(0x594021),
			TextColor:           utils.RGB(0xffffff),
			ContrastTextColor:   utils.RGB(0xffffff),
			GrayText:            utils.RGB(0x8c8c8c),
			RedText:             utils.RGB(0xff3030),
			RandomColorBase:     60,
			RandomColorWidth:    50,
			ChartLineColor:      utils.ARGB(0x14ffffff),
			ChartSelectionColor: utils.ARGB(0x3200ffff),
		}
	case PGCTSkyOrangeTheme:
		return PGCTTheme{
			BG:                  utils.RGB(0xFAFAFA),
			HalfBG:              utils.RGB(0xF6F2E5),
			SecondBG:            utils.RGB(0xF2EAD1),
			LesserContrastBg:    utils.RGB(0xEFE1BC),
			LessContrastBg:      utils.RGB(0xEBD9A8),
			ContrastBG:          utils.RGB(0xE7D193),
			TextColor:           utils.RGB(0x050505),
			ContrastTextColor:   utils.RGB(0x050505),
			GrayText:            utils.RGB(0x666666),
			RedText:             utils.RGB(0xff3030),
			RandomColorBase:     160,
			RandomColorWidth:    50,
			ChartLineColor:      utils.ARGB(0x50000000),
			ChartSelectionColor: utils.ARGB(0x55006666),
		}
	case PGCTPurpleTheme:
		return PGCTTheme{
			BG:                  utils.RGB(0x251333),
			HalfBG:              utils.RGB(0x30133F),
			SecondBG:            utils.RGB(0x3C134B),
			LesserContrastBg:    utils.RGB(0x471458),
			LessContrastBg:      utils.RGB(0x531464),
			ContrastBG:          utils.RGB(0x5E1470),
			TextColor:           utils.RGB(0xffffff),
			ContrastTextColor:   utils.RGB(0xffffff),
			GrayText:            utils.RGB(0x8c8c8c),
			RedText:             utils.RGB(0xff3030),
			RandomColorBase:     60,
			RandomColorWidth:    50,
			ChartLineColor:      utils.ARGB(0x14ffffff),
			ChartSelectionColor: utils.ARGB(0x3200ffff),
		}
	case PGCTSkyRedTheme:
		return PGCTTheme{
			BG:                  utils.RGB(0xFAFAFA),
			HalfBG:              utils.RGB(0xFBE1E1),
			SecondBG:            utils.RGB(0xFCC7C7),
			LesserContrastBg:    utils.RGB(0xFDAEAE),
			LessContrastBg:      utils.RGB(0xFE9494),
			ContrastBG:          utils.RGB(0xFF7B7B),
			TextColor:           utils.RGB(0x050505),
			ContrastTextColor:   utils.RGB(0x050505),
			GrayText:            utils.RGB(0x666666),
			RedText:             utils.RGB(0xff3030),
			RandomColorBase:     160,
			RandomColorWidth:    50,
			ChartLineColor:      utils.ARGB(0x50000000),
			ChartSelectionColor: utils.ARGB(0x55006666),
		}
	case PGCTSkyTealTheme:
		return PGCTTheme{
			BG:                  utils.RGB(0xFAFAFA),
			HalfBG:              utils.RGB(0xE4FBF3),
			SecondBG:            utils.RGB(0xCFFCEC),
			LesserContrastBg:    utils.RGB(0xB9FDE4),
			LessContrastBg:      utils.RGB(0xA4FEDD),
			ContrastBG:          utils.RGB(0x8EFFD6),
			TextColor:           utils.RGB(0x050505),
			ContrastTextColor:   utils.RGB(0x050505),
			GrayText:            utils.RGB(0x666666),
			RedText:             utils.RGB(0xff3030),
			RandomColorBase:     160,
			RandomColorWidth:    50,
			ChartLineColor:      utils.ARGB(0x50000000),
			ChartSelectionColor: utils.ARGB(0x55006666),
		}
	case PGCTBlueTheme:
		return PGCTTheme{
			BG:                  utils.RGB(0x040922),
			HalfBG:              utils.RGB(0x0A132F),
			SecondBG:            utils.RGB(0x101C3D),
			LesserContrastBg:    utils.RGB(0x17264A),
			LessContrastBg:      utils.RGB(0x1D2F58),
			ContrastBG:          utils.RGB(0x233965),
			TextColor:           utils.RGB(0xffffff),
			ContrastTextColor:   utils.RGB(0xffffff),
			GrayText:            utils.RGB(0x8c8c8c),
			RedText:             utils.RGB(0xff3030),
			RandomColorBase:     60,
			RandomColorWidth:    50,
			ChartLineColor:      utils.ARGB(0x14ffffff),
			ChartSelectionColor: utils.ARGB(0x3200ffff),
		}
	case PGCTGreenTheme:
		return PGCTTheme{
			BG:                  utils.RGB(0x042109),
			HalfBG:              utils.RGB(0x0C2F0E),
			SecondBG:            utils.RGB(0x133D14),
			LesserContrastBg:    utils.RGB(0x1B4A19),
			LessContrastBg:      utils.RGB(0x22581F),
			ContrastBG:          utils.RGB(0x2A6624),
			TextColor:           utils.RGB(0xffffff),
			ContrastTextColor:   utils.RGB(0xffffff),
			GrayText:            utils.RGB(0x8c8c8c),
			RedText:             utils.RGB(0xff3030),
			RandomColorBase:     60,
			RandomColorWidth:    50,
			ChartLineColor:      utils.ARGB(0x14ffffff),
			ChartSelectionColor: utils.ARGB(0x3200ffff),
		}
	case PGCTPGTheme:
		return PGCTTheme{
			BG:                  utils.RGB(0x201c19),
			HalfBG:              utils.RGB(0x241e1c),
			SecondBG:            utils.RGB(0x251f1c),
			LesserContrastBg:    utils.RGB(0x2f2623),
			LessContrastBg:      utils.RGB(0x3e180e),
			ContrastBG:          utils.RGB(0x4d2215),
			TextColor:           utils.RGB(0xbfad84),
			ContrastTextColor:   utils.RGB(0xbfad84),
			GrayText:            utils.RGB(0x8c8c8c),
			RedText:             utils.RGB(0x86826d),
			RandomColorBase:     60,
			RandomColorWidth:    50,
			ChartLineColor:      utils.ARGB(0x14ffffff),
			ChartSelectionColor: utils.ARGB(0x3200ffff),
		}
	case PGCTLightPGTheme:
		return PGCTTheme{
			BG:                  utils.RGB(0x9b957c),
			HalfBG:              utils.RGB(0x877650),
			SecondBG:            utils.RGB(0x796a48),
			LesserContrastBg:    utils.RGB(0x62563c),
			LessContrastBg:      utils.RGB(0x876631),
			ContrastBG:          utils.RGB(0xa6793a),
			TextColor:           utils.RGB(0x25221d),
			ContrastTextColor:   utils.RGB(0x25221d),
			GrayText:            utils.RGB(0x444444),
			RedText:             utils.RGB(0xff3030),
			RandomColorBase:     90,
			RandomColorWidth:    50,
			ChartLineColor:      utils.ARGB(0x50000000),
			ChartSelectionColor: utils.ARGB(0x3200ffff),
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
	case PGCTNightTheme:
		return "Night"
	case PGCTBluePurpleTheme:
		return "Blue/Purple"
	case PGCTSkyBlueTheme:
		return "Sky Blue"
	case PGCTPinkTheme:
		return "Sky Pink"
	case PGCTDarkRedTheme:
		return "Dark Red"
	case PGCTOliveTheme:
		return "Olive"
	case PGCTBrownTheme:
		return "Brown"
	case PGCTSkyOrangeTheme:
		return "Sky Orange"
	case PGCTPurpleTheme:
		return "Purple"
	case PGCTSkyRedTheme:
		return "Sky Red"
	case PGCTSkyTealTheme:
		return "Sky Teal"
	case PGCTBlueTheme:
		return "Blue"
	case PGCTGreenTheme:
		return "Green"
	case PGCTPGTheme:
		return "Project Gorgon"
	case PGCTLightPGTheme:
		return "Light Project Gorgon"
	}

	return ""
}
