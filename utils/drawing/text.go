package drawing

import (
	"PGCombatTracker/abstract"
	"github.com/fogleman/gg"
	"golang.org/x/image/font"
	"image/color"
)

type TextStyle struct {
	Face       font.Face
	Color      color.NRGBA
	fontHeight float64
	baseline   float64
}

func (ts TextStyle) measureString(text string) F64Point {
	d := &font.Drawer{
		Face: ts.Face,
	}
	a := d.MeasureString(text)
	return F64Point{
		X: float64(a >> 6),
		Y: ts.fontHeight,
	}
}

func MakeTextStyle(face font.Face, color color.NRGBA) TextStyle {
	return TextStyle{
		Face:       face,
		Color:      color,
		fontHeight: float64(face.Metrics().Height) / 64,
		baseline:   float64(face.Metrics().Descent) / 128,
	}
}

func (ts TextStyle) Layout(text string) Widget {
	return func(ltx Context) Result {
		size := ts.measureString(text)

		return Result{
			Size: size,
			Draw: func(gg *gg.Context) {
				gg.SetColor(ts.Color)
				gg.SetFontFace(ts.Face)
				gg.DrawString(text, 0, size.Y-ts.baseline)
			},
		}
	}
}

func StyleFontPack(fontpack *abstract.FontPack, color color.NRGBA) *StyledFontPack {
	return &StyledFontPack{
		Smallest: MakeTextStyle(fontpack.Smallest, color),
		Smaller:  MakeTextStyle(fontpack.Smaller, color),
		Body:     MakeTextStyle(fontpack.Body, color),
		Heading:  MakeTextStyle(fontpack.Heading, color),
	}
}

type StyledFontPack struct {
	Smallest TextStyle
	Smaller  TextStyle
	Body     TextStyle
	Heading  TextStyle
}
