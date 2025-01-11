package collectors

import (
	"PGCombatTracker/utils/drawing"
	"fmt"
	"gioui.org/layout"
	"github.com/fogleman/gg"
	"image/color"
)

func layoutTitle(styledFonts *drawing.StyledFontPack, tabName string, parameters, inner drawing.Widget) drawing.Widget {
	return drawing.Flex{
		Axis: layout.Vertical,
	}.Layout(
		drawing.Rigid(drawing.UniformInset(drawing.CommonSpacing*2).Layout(
			drawing.Flex{Axis: layout.Vertical}.Layout(
				drawing.Rigid(styledFonts.Heading.Layout(fmt.Sprintf("%v Tab", tabName))),
				drawing.FlexVSpacer(drawing.CommonSpacing*2),
				drawing.Rigid(parameters),
			),
		)),
		drawing.FlexVSpacer(drawing.CommonSpacing),
		drawing.Rigid(inner),
		drawing.FlexVSpacer(drawing.CommonSpacing),
		drawing.Rigid(drawing.UniformInset(drawing.CommonSpacing*2).Layout(
			drawing.Flex{
				Axis:      layout.Horizontal,
				Alignment: layout.End,
				ExpandW:   true,
			}.Layout(
				drawing.Rigid(styledFonts.Smaller.Layout("PGCombatTracker")),
				drawing.FlexHSpacer(drawing.CommonSpacing),
				drawing.Rigid(styledFonts.Smallest.Layout("by TheJebForge")),
				drawing.Flexer(1),
				drawing.Rigid(styledFonts.Smaller.Layout("github.com/TheJebForge/PGCombatTracker")),
			),
		)),
	)
}

func percentBar(progress float64, color color.NRGBA) drawing.Widget {
	return func(ltx drawing.Context) drawing.Result {
		size := ltx.Min
		percentWidth := size.X * progress

		return drawing.Result{
			Size: size,
			Draw: func(gg *gg.Context) {
				gg.SetColor(color)
				gg.DrawRoundedRectangle(0, 0, percentWidth, size.Y, drawing.CommonSpacing*2)
				gg.Fill()
			},
		}
	}
}
