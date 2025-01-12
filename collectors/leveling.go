package collectors

import (
	"PGCombatTracker/abstract"
	"PGCombatTracker/ui/components"
	"PGCombatTracker/utils"
	"PGCombatTracker/utils/drawing"
	"cmp"
	"fmt"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/unit"
	"gioui.org/widget"
	"image"
	"log"
	"slices"
	"time"
)

type skillXP struct {
	name   string
	xp     int
	levels int
	chart  *components.TimeBasedChart
}

type subjectiveSkillsXP struct {
	name           string
	skills         []skillXP
	timeController *components.TimeController
	totalXP        int
	maxXP          int
	maxRange       components.DataRange
}

func NewLevelingCollector() *LevelingCollector {
	subjectDropdown, err := components.NewDropdown(
		"Subject",
		subjectChoice(""),
	)
	if err != nil {
		log.Fatalln(err)
	}

	displayDropdown, err := components.NewDropdown(
		"Display",
		DisplayBars,
		DisplayPie,
		DisplayGraphs,
	)
	if err != nil {
		log.Fatalln(err)
	}

	return &LevelingCollector{
		all: subjectiveSkillsXP{
			timeController: components.NewTimeControllerOrCrash(components.NewTimeBasedChart("Total")),
		},
		subjectDropdown: subjectDropdown,
		displayDropdown: displayDropdown,
		longFormatBool:  &widget.Bool{},
	}
}

type LevelingCollector struct {
	all      subjectiveSkillsXP
	subjects []subjectiveSkillsXP

	currentSubject  string
	currentDisplay  displayChoice
	subjectDropdown *components.Dropdown
	displayDropdown *components.Dropdown
	longFormatBool  *widget.Bool
}

func (l *LevelingCollector) Reset(info abstract.StatisticsInformation) {
	l.all = subjectiveSkillsXP{
		timeController: components.NewTimeControllerOrCrash(components.NewTimeBasedChart("Total")),
	}
	l.subjects = nil
	l.subjectDropdown.SetOptions([]fmt.Stringer{subjectChoice("")})
	l.currentSubject = ""
}

func (l *LevelingCollector) Tick(info abstract.StatisticsInformation, at time.Time) {

}

func (l *LevelingCollector) Collect(info abstract.StatisticsInformation, event *abstract.ChatEvent) error {
	xp, xpOk := event.Contents.(*abstract.XPGained)
	leveledXP, levelOk := event.Contents.(*abstract.XPGainedLeveledUp)

	var skillName string
	var gainedXP int
	switch {
	case xpOk:
		skillName = xp.Skill
		gainedXP = xp.XP
	case levelOk:
		skillName = leveledXP.Skill
		gainedXP = leveledXP.XP
	default:
		return nil
	}

	findSkillXp := func(skill skillXP) bool {
		return skill.name == skillName
	}
	createSkillXp := func(leveled bool) func() skillXP {
		return func() skillXP {
			var level int
			if leveled {
				level = 1
			}

			chart := components.NewTimeBasedChart(skillName)
			chart.Add(components.TimePoint{
				Time:    event.Time,
				Value:   gainedXP,
				Details: XPValue(gainedXP),
			})

			return skillXP{
				name:   skillName,
				xp:     gainedXP,
				levels: level,
				chart:  chart,
			}
		}
	}
	updateSkillXp := func(leveled bool) func(skillXP) skillXP {
		return func(skill skillXP) skillXP {
			skill.xp += gainedXP
			if leveled {
				skill.levels++
			}
			skill.chart.Add(components.TimePoint{
				Time:    event.Time,
				Value:   skill.xp,
				Details: XPValue(skill.xp),
			})

			return skill
		}
	}
	skillXpSort := func(a, b skillXP) int {
		return cmp.Compare(b.xp, a.xp)
	}
	skillXpMax := func(a, b skillXP) int {
		return cmp.Compare(a.xp, b.xp)
	}
	processSubjectiveSkillXp := func(stats *subjectiveSkillsXP, leveled bool) {
		stats.skills = utils.CreateUpdate(
			stats.skills,
			findSkillXp,
			createSkillXp(leveled),
			updateSkillXp(leveled),
		)
		stats.totalXP += gainedXP
		stats.timeController.Add(components.TimePoint{
			Time:  event.Time,
			Value: stats.totalXP,
		})
		slices.SortFunc(stats.skills, skillXpSort)
		stats.maxXP = slices.MaxFunc(stats.skills, skillXpMax).xp
		stats.maxRange = stats.maxRange.Expand(stats.maxXP)
	}
	processSubjects := func(leveled bool) {
		l.subjects = utils.CreateUpdate(
			l.subjects,
			func(subject subjectiveSkillsXP) bool {
				return subject.name == info.CurrentUsername()
			},
			func() subjectiveSkillsXP {
				timeController := components.NewTimeControllerOrCrash(components.NewTimeBasedChart("Total"))
				timeController.Add(components.TimePoint{
					Time:  event.Time,
					Value: gainedXP,
				})
				return subjectiveSkillsXP{
					name: info.CurrentUsername(),
					skills: []skillXP{
						createSkillXp(leveled)(),
					},
					totalXP:        gainedXP,
					maxXP:          gainedXP,
					timeController: timeController,
					maxRange:       components.DataRange{Max: gainedXP},
				}
			},
			func(subject subjectiveSkillsXP) subjectiveSkillsXP {
				processSubjectiveSkillXp(&subject, leveled)
				return subject
			},
		)
	}

	if xpOk {
		processSubjectiveSkillXp(&l.all, false)
		processSubjects(false)
	} else {
		processSubjectiveSkillXp(&l.all, true)
		processSubjects(true)
	}

	// Add subjects to dropdown
	l.subjectDropdown.SetOptions(utils.CreateUpdate(
		l.subjectDropdown.Options(),
		func(item fmt.Stringer) bool {
			casted := item.(subjectChoice)
			return string(casted) == info.CurrentUsername()
		},
		func() fmt.Stringer {
			return subjectChoice(info.CurrentUsername())
		},
		func(stringer fmt.Stringer) fmt.Stringer {
			return stringer
		},
	))

	return nil
}

func (l *LevelingCollector) TabName() string {
	return "XP Gained"
}

type XPValue int

func (xp XPValue) StringCL(long bool) string {
	if long {
		return fmt.Sprintf("%d XP", xp)
	} else {
		return fmt.Sprintf("%v XP", utils.FormatNumber(int(xp)))
	}
}
func (xp XPValue) Interpolate(other utils.Interpolatable, t float64) utils.Interpolatable {
	otherXP, ok := other.(XPValue)
	if !ok {
		return other
	}

	return skillUseCounter(utils.LerpInt(
		int(xp),
		int(otherXP),
		t,
	))
}
func (xp XPValue) InterpolateILF(other utils.InterpolatableLongFormatable, t float64) utils.InterpolatableLongFormatable {
	return xp.Interpolate(other, t).(utils.InterpolatableLongFormatable)
}

func (l *LevelingCollector) drawWidget(state abstract.LayeredState, skill skillXP, widget layout.Widget, size unit.Dp) layout.Widget {
	return drawUniversalStatsText(
		state, XPValue(skill.xp),
		widget, skill.levels,
		skill.name, "leveled %v times",
		size, l.longFormatBool.Value,
	)
}

func (l *LevelingCollector) drawBar(state abstract.LayeredState, skill skillXP, maxXP int, size unit.Dp) layout.Widget {
	return drawUniversalBar(
		state, XPValue(skill.xp),
		skill.xp, maxXP, skill.levels,
		skill.name, "leveled %v times",
		size, l.longFormatBool.Value,
	)
}

func (l *LevelingCollector) UI(state abstract.LayeredState) (layout.Widget, []layout.Widget) {
	if l.subjectDropdown.Changed() {
		l.currentSubject = string(l.subjectDropdown.Value.(subjectChoice))
	}

	if l.displayDropdown.Changed() {
		l.currentDisplay = l.displayDropdown.Value.(displayChoice)
	}

	subject := l.all
	for _, possibleSubject := range l.subjects {
		if possibleSubject.name == l.currentSubject {
			subject = possibleSubject
			break
		}
	}

	topWidget := topBarSurface(func(gtx layout.Context) layout.Dimensions {
		if l.longFormatBool.Update(gtx) {
			gtx.Source.Execute(op.InvalidateCmd{})
		}

		return layout.Flex{
			Axis: layout.Vertical,
		}.Layout(
			gtx,
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return components.HorizontalWrap{
					Alignment:   layout.Middle,
					Spacing:     utils.CommonSpacing,
					LineSpacing: utils.CommonSpacing,
				}.Layout(
					gtx,
					defaultDropdownStyle(state, l.subjectDropdown).Layout,
					defaultCheckboxStyle(state, l.longFormatBool, "Use long numbers").Layout,
					defaultDropdownStyle(state, l.displayDropdown).Layout,
				)
			}),
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				if l.currentDisplay == DisplayGraphs {
					return layout.Flex{
						Axis: layout.Vertical,
					}.Layout(
						gtx,
						utils.FlexSpacerH(utils.CommonSpacing),
						layout.Rigid(components.StyleTimeController(state.Theme(), subject.timeController).Layout),
					)
				}

				return layout.Dimensions{}
			}),
		)
	})

	var widgets []layout.Widget

	switch l.currentDisplay {
	case DisplayBars:
		for _, skill := range subject.skills {
			widgets = append(widgets, l.drawBar(state, skill, subject.maxXP, 40))
		}
	case DisplayPie:
		widgets = append(widgets, func(gtx layout.Context) layout.Dimensions {
			var totalValue int
			pieItems := make([]components.PieChartItem, 0, max(1, len(subject.skills)))
			for _, skill := range subject.skills {
				pieItems = append(pieItems, components.PieChartItem{
					Name:    skill.name,
					Value:   skill.xp,
					SubText: XPValue(skill.xp).StringCL(l.longFormatBool.Value),
				})
				totalValue += skill.xp
			}

			style := components.StylePieChart(state.Theme())
			style.TextSize = 12

			return style.Layout(
				gtx,
				totalValue,
				pieItems...,
			)
		})
	case DisplayGraphs:
		controller := subject.timeController

		for _, skill := range subject.skills {
			skill.chart.DisplayTimeFrame = controller.CurrentTimeFrame
			skill.chart.DisplayValueRange = subject.maxRange

			chartStyle := components.StyleTimeBasedChart(state.Theme(), skill.chart)
			chartStyle.Color = components.StringToColor(skill.name)
			chartStyle.LongFormat = l.longFormatBool.Value

			widgets = append(widgets, l.drawWidget(state, skill, chartStyle.Layout, 100))
		}
	}

	return topWidget, widgets
}

func (l *LevelingCollector) exportWidget(styledFonts *drawing.StyledFontPack, skill skillXP, widget drawing.Widget) drawing.Widget {
	return exportUniversalStatsTextAsStack(
		styledFonts, XPValue(skill.xp),
		widget, skill.levels,
		skill.name, "leveled %v times",
		l.longFormatBool.Value,
	)
}

func (l *LevelingCollector) exportBar(styledFonts *drawing.StyledFontPack, skill skillXP, maxXP int) drawing.Widget {
	return exportUniversalBar(
		styledFonts, XPValue(skill.xp),
		skill.xp, maxXP, skill.levels,
		skill.name, "leveled %v times",
		l.longFormatBool.Value,
	)
}

func (l *LevelingCollector) Export(state abstract.LayeredState) image.Image {
	subject := l.all
	for _, possibleSubject := range l.subjects {
		if possibleSubject.name == l.currentSubject {
			subject = possibleSubject
			break
		}
	}

	styledFonts := drawing.StyleFontPack(state.FontPack(), state.Theme().Fg)

	body := drawing.Empty

	switch l.currentDisplay {
	case DisplayBars:
		items := make([]drawing.FlexChild, 0, len(subject.skills)*2-1)

		for i, skill := range subject.skills {
			if i != 0 {
				items = append(items, drawing.FlexVSpacer(drawing.CommonSpacing))
			}

			items = append(items, drawing.Rigid(l.exportBar(styledFonts, skill, subject.maxXP)))
		}

		body = drawing.Flex{
			ExpandW: true,
			Axis:    layout.Vertical,
		}.Layout(
			items...,
		)
	case DisplayPie:
		var totalValue int

		pieItems := make([]drawing.PieChartItem, 0, max(1, len(subject.skills)))
		for _, skill := range subject.skills {
			pieItems = append(pieItems, drawing.PieChartItem{
				Name:    skill.name,
				Value:   skill.xp,
				SubText: XPValue(skill.xp).StringCL(l.longFormatBool.Value),
			})
			totalValue += skill.xp
		}

		style := drawing.PieChart{
			OverflowLimit: 15,
			ColorBoxSize:  drawing.CommonSpacing * 4,
			TextStyle:     styledFonts.Body,
			SubTextStyle:  drawing.MakeTextStyle(styledFonts.Smaller.Face, utils.GrayText),
		}

		body = style.Layout(totalValue, pieItems...)
	case DisplayGraphs:
		items := make([]drawing.FlexChild, 0, len(subject.skills)*2-1+2)

		items = append(
			items,
			drawing.Rigid(exportTimeFrame(styledFonts, subject.timeController.CurrentTimeFrame)),
			drawing.FlexVSpacer(drawing.CommonSpacing),
		)

		for i, skill := range subject.skills {
			if i != 0 {
				items = append(items, drawing.FlexVSpacer(drawing.CommonSpacing))
			}

			style := drawing.StyleAreaChart(skill.chart, components.StringToColor(skill.name))
			style.MinHeight = 200

			items = append(items, drawing.Rigid(l.exportWidget(styledFonts, skill, style.Layout())))
		}

		body = drawing.Flex{
			ExpandW: true,
			Axis:    layout.Vertical,
		}.Layout(
			items...,
		)
	}

	base := layoutTitle(
		styledFonts,
		l.TabName(),
		drawing.HorizontalWrap{
			Alignment:   layout.Middle,
			Spacing:     drawing.CommonSpacing * 3,
			LineSpacing: drawing.CommonSpacing,
		}.Layout(
			func(ltx drawing.Context) drawing.Result {
				if l.currentSubject != "" {
					return styledFonts.Smaller.Layout(fmt.Sprintf("Subject: %v", l.currentSubject))(ltx)
				}

				return styledFonts.Smaller.Layout("Subject: All")(ltx)
			},
			styledFonts.Smaller.Layout(fmt.Sprintf("Display: %v", l.currentDisplay)),
		),
		drawing.RoundedSurface(
			utils.SecondBG,
			body,
		),
	)

	return drawing.ExportImage(state.Theme(), base, drawing.F64(800, 10000))
}
