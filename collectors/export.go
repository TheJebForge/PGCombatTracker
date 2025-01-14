package collectors

import (
	"PGCombatTracker/ui/components"
	"PGCombatTracker/utils"
	"PGCombatTracker/utils/drawing"
	"fmt"
	"gioui.org/layout"
	"github.com/fogleman/gg"
	"image/color"
	"math"
	"time"
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
		percentWidth := size.X * max(min(1, progress), 0)

		return drawing.Result{
			Size: size,
			Draw: func(gg *gg.Context) {
				gg.SetColor(color)
				gg.DrawRoundedRectangle(0, 0, percentWidth, size.Y, min(drawing.CommonSpacing*2, percentWidth))
				gg.Fill()
			},
		}
	}
}

func exportUniversalStatsTextItself(
	styledFonts *drawing.StyledFontPack,
	sideText utils.LongFormatable, amount int,
	name, amountFormat string,
	long bool,
) drawing.Widget {
	return drawing.Flex{
		ExpandW: true,
	}.Layout(
		drawing.Rigid(func(ltx drawing.Context) drawing.Result {
			if amount == 0 {
				return styledFonts.Body.Layout(name)(ltx)
			} else {
				return drawing.Flex{
					Axis: layout.Vertical,
				}.Layout(
					drawing.Rigid(styledFonts.Body.Layout(name)),
					drawing.FlexVSpacer(drawing.CommonSpacing),
					drawing.Rigid(styledFonts.Smaller.Layout(fmt.Sprintf(amountFormat, amount))),
				)(ltx)
			}
		}),
		drawing.Flexer(1),
		drawing.Rigid(styledFonts.Body.Layout(sideText.StringCL(long))),
	)
}

func exportUniversalStatsTextAsSurface(
	styledFonts *drawing.StyledFontPack,
	sideText utils.LongFormatable,
	background drawing.Widget, amount int,
	name, amountFormat string,
	long bool,
) drawing.Widget {
	return drawing.CustomSurface(
		background,
		exportUniversalStatsTextItself(styledFonts, sideText, amount, name, amountFormat, long),
	)
}

func exportUniversalStatsTextAsStack(
	styledFonts *drawing.StyledFontPack,
	sideText utils.LongFormatable,
	background drawing.Widget, amount int,
	name, amountFormat string,
	long bool,
) drawing.Widget {
	return drawing.Stack{
		Wide: true,
	}.Layout(
		background,
		drawing.UniformInset(drawing.CommonSpacing*2).Layout(exportUniversalStatsTextItself(
			styledFonts,
			sideText,
			amount,
			name,
			amountFormat,
			long,
		)),
	)
}

func exportUniversalBar(
	styledFonts *drawing.StyledFontPack,
	sideText utils.LongFormatable,
	value, max, amount int,
	name, amountFormat string,
	long bool,
) drawing.Widget {
	return exportUniversalStatsTextAsSurface(
		styledFonts,
		sideText,
		func(ltx drawing.Context) drawing.Result {
			progress := float64(value) / float64(max)
			if math.IsNaN(progress) {
				progress = 1
			}

			return percentBar(progress, components.StringToColor(name))(ltx)
		}, amount,
		name, amountFormat,
		long,
	)
}

func exportTimeFrame(styledFonts *drawing.StyledFontPack, timeFrame components.TimeFrame) drawing.Widget {
	return drawing.RoundedSurface(utils.LesserContrastBg, drawing.Flex{
		Axis:      layout.Horizontal,
		Alignment: layout.Middle,
		ExpandW:   true,
	}.Layout(
		drawing.Rigid(styledFonts.Smaller.Layout(fmt.Sprintf(
			"Data from %v",
			timeFrame.From.Format(time.DateTime),
		))),
		drawing.Flexer(1),
		drawing.Rigid(styledFonts.Smaller.Layout(fmt.Sprintf(
			"to %v",
			timeFrame.To.Format(time.DateTime),
		))),
	))
}
