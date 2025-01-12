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
	"strings"
	"time"
)

type skillUse struct {
	name     string
	amount   int
	damage   abstract.Vitals
	lastUsed time.Time
	chart    *components.TimeBasedChart
}

type subjectiveSkillUses struct {
	name           string
	skills         []skillUse
	timeController *components.TimeController
	totalUsed      int
	maxUsed        int
	maxRange       components.DataRange
}

type skillUseType int

const (
	UseAllies skillUseType = iota
	UseEnemies
	UseAll
	UseCustom
)

type skillUseSubject struct {
	ty   skillUseType
	name string
}

func (s skillUseSubject) String() string {
	switch s.ty {
	case UseAllies:
		return "Allies"
	case UseEnemies:
		return "Enemies"
	case UseAll:
		return "All"
	case UseCustom:
		return s.name
	}
	return ""
}

func (s skillUseSubject) EqualTo(other skillUseSubject) bool {
	return s.ty == other.ty && s.name == other.name
}

func freshSubjectiveSkillUse() subjectiveSkillUses {
	return subjectiveSkillUses{
		timeController: components.NewTimeControllerOrCrash(components.NewTimeBasedChart("Total")),
	}
}

func NewSkillsCollector() *SkillsCollector {
	subjectDropdown, err := components.NewDropdown(
		"Subject",
		skillUseSubject{
			ty: UseAllies,
		},
		skillUseSubject{
			ty: UseEnemies,
		},
		skillUseSubject{
			ty: UseAll,
		},
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

	return &SkillsCollector{
		allies:  freshSubjectiveSkillUse(),
		enemies: freshSubjectiveSkillUse(),
		all:     freshSubjectiveSkillUse(),

		subjectDropdown: subjectDropdown,
		displayDropdown: displayDropdown,
		longFormatBool:  &widget.Bool{},
	}
}

type SkillsCollector struct {
	allies   subjectiveSkillUses
	enemies  subjectiveSkillUses
	all      subjectiveSkillUses
	subjects []subjectiveSkillUses

	currentSubject  skillUseSubject
	currentDisplay  displayChoice
	subjectDropdown *components.Dropdown
	displayDropdown *components.Dropdown
	longFormatBool  *widget.Bool
}

func (s *SkillsCollector) Reset(info abstract.StatisticsInformation) {
	s.allies = freshSubjectiveSkillUse()
	s.enemies = freshSubjectiveSkillUse()
	s.all = freshSubjectiveSkillUse()
	s.subjects = nil

	s.currentSubject = skillUseSubject{}
	s.subjectDropdown.SetOptions([]fmt.Stringer{
		skillUseSubject{
			ty: UseAllies,
		},
		skillUseSubject{
			ty: UseEnemies,
		},
		skillUseSubject{
			ty: UseAll,
		},
	})
}

func (s *SkillsCollector) isAlly(info abstract.StatisticsInformation, subject, skill string) bool {
	if subject == info.CurrentUsername() {
		return true
	}

	if strings.Contains(skill, "(Pet)") {
		return true
	}

	lowTrimmedSubject := strings.ToLower(strings.TrimSpace(SplitOffId(subject)))

	for _, expectedName := range info.Settings().EntitiesThatCountAsPets {
		if lowTrimmedSubject == strings.ToLower(strings.TrimSpace(expectedName)) {
			return true
		}
	}

	return false
}

type skillUseCounter int

func (counter skillUseCounter) StringCL(long bool) string {
	if long {
		return fmt.Sprintf("%d use(s)", counter)
	} else {
		return fmt.Sprintf("%v use(s)", utils.FormatNumber(int(counter)))
	}
}
func (counter skillUseCounter) Interpolate(other utils.Interpolatable, t float64) utils.Interpolatable {
	otherCounter, ok := other.(skillUseCounter)
	if !ok {
		return other
	}

	return skillUseCounter(utils.LerpInt(
		int(counter),
		int(otherCounter),
		t,
	))
}
func (counter skillUseCounter) InterpolateILF(other utils.InterpolatableLongFormatable, t float64) utils.InterpolatableLongFormatable {
	return counter.Interpolate(other, t).(utils.InterpolatableLongFormatable)
}

func (s *SkillsCollector) ingestSkillUse(info abstract.StatisticsInformation, event *abstract.ChatEvent) {
	skill := event.Contents.(*abstract.SkillUse)
	skillName := skill.Skill

	if info.Settings().RemoveLevelsFromSkills {
		skillName = SplitOffId(skillName)
	}

	damage := abstract.Vitals{}
	if skill.Damage != nil {
		damage = *skill.Damage
	}

	findSkillUse := func(use skillUse) bool {
		return use.name == skillName
	}
	createSkillUse := func() skillUse {
		chart := components.NewTimeBasedChart(skillName)
		chart.Add(components.TimePoint{
			Time:    event.Time,
			Value:   1,
			Details: skillUseCounter(1),
		})

		return skillUse{
			name:     skillName,
			amount:   1,
			damage:   damage,
			lastUsed: event.Time,
			chart:    chart,
		}
	}
	updateSkillUse := func(use skillUse) skillUse {
		if use.lastUsed != event.Time {
			use.amount++
		}
		use.damage = use.damage.Add(damage)
		use.lastUsed = event.Time
		use.chart.Add(components.TimePoint{
			Time:    event.Time,
			Value:   use.amount,
			Details: skillUseCounter(use.amount),
		})

		return use
	}
	skillUseSort := func(a, b skillUse) int {
		return cmp.Compare(b.amount, a.amount)
	}
	skillUseMax := func(a, b skillUse) int {
		return cmp.Compare(a.amount, b.amount)
	}
	processSubjectiveSkillUses := func(stats *subjectiveSkillUses) {
		stats.skills = utils.CreateUpdate(
			stats.skills,
			findSkillUse,
			createSkillUse,
			updateSkillUse,
		)
		stats.totalUsed += 1
		stats.timeController.Add(components.TimePoint{
			Time:  event.Time,
			Value: stats.totalUsed,
		})
		slices.SortFunc(stats.skills, skillUseSort)
		stats.maxUsed = slices.MaxFunc(stats.skills, skillUseMax).amount
		stats.maxRange = stats.maxRange.Expand(stats.maxUsed)
	}

	processSubjectiveSkillUses(&s.all)

	isAlly := s.isAlly(info, skill.Subject, skill.Skill)
	if isAlly {
		processSubjectiveSkillUses(&s.allies)
	} else {
		processSubjectiveSkillUses(&s.enemies)
	}

	subject := skill.Subject
	if !isAlly {
		subject = SplitOffId(subject)
	}

	s.subjects = utils.CreateUpdate(
		s.subjects,
		func(uses subjectiveSkillUses) bool {
			return uses.name == subject
		},
		func() subjectiveSkillUses {
			timeController := components.NewTimeControllerOrCrash(components.NewTimeBasedChart("Total"))
			timeController.Add(components.TimePoint{
				Time:  event.Time,
				Value: 1,
			})
			return subjectiveSkillUses{
				name:      subject,
				maxUsed:   1,
				totalUsed: 1,
				skills: []skillUse{
					createSkillUse(),
				},
				timeController: timeController,
				maxRange:       components.DataRange{Max: 1},
			}
		},
		func(uses subjectiveSkillUses) subjectiveSkillUses {
			processSubjectiveSkillUses(&uses)
			return uses
		},
	)

	option := skillUseSubject{
		ty:   UseCustom,
		name: subject,
	}

	s.subjectDropdown.SetOptions(utils.CreateUpdate(
		s.subjectDropdown.Options(),
		func(item fmt.Stringer) bool {
			casted := item.(skillUseSubject)
			return casted.EqualTo(option)
		},
		func() fmt.Stringer {
			return option
		},
		func(stringer fmt.Stringer) fmt.Stringer {
			return stringer
		},
	))
}

func (s *SkillsCollector) Tick(info abstract.StatisticsInformation, at time.Time) {

}

func (s *SkillsCollector) Collect(info abstract.StatisticsInformation, event *abstract.ChatEvent) error {
	if _, ok := event.Contents.(*abstract.SkillUse); ok {
		s.ingestSkillUse(info, event)
	}

	return nil
}

func (s *SkillsCollector) TabName() string {
	return "Skill Uses"
}

func (s *SkillsCollector) drawWidget(state abstract.LayeredState, skill skillUse, widget layout.Widget, size unit.Dp) layout.Widget {
	return drawUniversalStatsText(
		state, skill.damage,
		widget, skill.amount,
		skill.name, "used %v times",
		size, s.longFormatBool.Value,
	)
}

func (s *SkillsCollector) drawBar(state abstract.LayeredState, skill skillUse, maxUsed int, size unit.Dp) layout.Widget {
	return drawUniversalBar(
		state, skill.damage,
		skill.amount, maxUsed, skill.amount,
		skill.name, "used %v times",
		size, s.longFormatBool.Value,
	)
}

func (s *SkillsCollector) UI(state abstract.LayeredState) (layout.Widget, []layout.Widget) {
	if s.subjectDropdown.Changed() {
		s.currentSubject = s.subjectDropdown.Value.(skillUseSubject)
	}

	if s.displayDropdown.Changed() {
		s.currentDisplay = s.displayDropdown.Value.(displayChoice)
	}

	var uses subjectiveSkillUses
	switch s.currentSubject.ty {
	case UseAllies:
		uses = s.allies
	case UseEnemies:
		uses = s.enemies
	case UseAll:
		uses = s.all
	case UseCustom:
		for _, potentialUses := range s.subjects {
			if potentialUses.name == s.currentSubject.name {
				uses = potentialUses
				break
			}
		}
	}

	topWidget := topBarSurface(func(gtx layout.Context) layout.Dimensions {
		if s.longFormatBool.Update(gtx) {
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
					defaultDropdownStyle(state, s.subjectDropdown).Layout,
					defaultCheckboxStyle(state, s.longFormatBool, "Use long numbers").Layout,
					defaultDropdownStyle(state, s.displayDropdown).Layout,
				)
			}),
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				if s.currentDisplay == DisplayGraphs {
					return layout.Flex{
						Axis: layout.Vertical,
					}.Layout(
						gtx,
						utils.FlexSpacerH(utils.CommonSpacing),
						layout.Rigid(components.StyleTimeController(state.Theme(), uses.timeController).Layout),
					)
				}

				return layout.Dimensions{}
			}),
		)
	})

	var widgets []layout.Widget

	switch s.currentDisplay {
	case DisplayBars:
		for _, skill := range uses.skills {
			widgets = append(widgets, s.drawBar(state, skill, uses.maxUsed, 40))
		}
	case DisplayPie:
		widgets = append(widgets, func(gtx layout.Context) layout.Dimensions {
			var totalValue int
			pieItems := make([]components.PieChartItem, 0, max(1, len(uses.skills)))
			for _, skill := range uses.skills {
				pieItems = append(pieItems, components.PieChartItem{
					Name:    skill.name,
					Value:   skill.amount,
					SubText: skillUseCounter(skill.amount).StringCL(s.longFormatBool.Value),
				})
				totalValue += skill.amount
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
		controller := uses.timeController

		for _, skill := range uses.skills {
			skill.chart.DisplayTimeFrame = controller.CurrentTimeFrame
			skill.chart.DisplayValueRange = uses.maxRange

			chartStyle := components.StyleTimeBasedChart(state.Theme(), skill.chart)
			chartStyle.Color = components.StringToColor(skill.name)
			chartStyle.LongFormat = s.longFormatBool.Value

			widgets = append(widgets, s.drawWidget(state, skill, chartStyle.Layout, 100))
		}
	}

	return topWidget, widgets
}

func (s *SkillsCollector) exportWidget(styledFonts *drawing.StyledFontPack, skill skillUse, widget drawing.Widget) drawing.Widget {
	return exportUniversalStatsTextAsStack(
		styledFonts, skill.damage,
		widget, skill.amount,
		skill.name, "used %v times",
		s.longFormatBool.Value,
	)
}

func (s *SkillsCollector) exportBar(styledFonts *drawing.StyledFontPack, skill skillUse, maxUsed int) drawing.Widget {
	return exportUniversalBar(
		styledFonts, skill.damage,
		skill.amount, maxUsed, skill.amount,
		skill.name, "used %v times",
		s.longFormatBool.Value,
	)
}

func (s *SkillsCollector) Export(state abstract.LayeredState) image.Image {
	var uses subjectiveSkillUses
	switch s.currentSubject.ty {
	case UseAllies:
		uses = s.allies
	case UseEnemies:
		uses = s.enemies
	case UseAll:
		uses = s.all
	case UseCustom:
		for _, potentialUses := range s.subjects {
			if potentialUses.name == s.currentSubject.name {
				uses = potentialUses
				break
			}
		}
	}

	styledFonts := drawing.StyleFontPack(state.FontPack(), state.Theme().Fg)

	body := drawing.Empty

	switch s.currentDisplay {
	case DisplayBars:
		items := make([]drawing.FlexChild, 0, len(uses.skills)*2-1)

		for i, skill := range uses.skills {
			if i != 0 {
				items = append(items, drawing.FlexVSpacer(drawing.CommonSpacing))
			}

			items = append(items, drawing.Rigid(s.exportBar(styledFonts, skill, uses.maxUsed)))
		}

		body = drawing.Flex{
			ExpandW: true,
			Axis:    layout.Vertical,
		}.Layout(
			items...,
		)
	case DisplayPie:
		var totalValue int

		pieItems := make([]drawing.PieChartItem, 0, max(1, len(uses.skills)))
		for _, skill := range uses.skills {
			pieItems = append(pieItems, drawing.PieChartItem{
				Name:    skill.name,
				Value:   skill.amount,
				SubText: skillUseCounter(skill.amount).StringCL(s.longFormatBool.Value),
			})
			totalValue += skill.amount
		}

		style := drawing.PieChart{
			OverflowLimit: 15,
			ColorBoxSize:  drawing.CommonSpacing * 4,
			TextStyle:     styledFonts.Body,
			SubTextStyle:  drawing.MakeTextStyle(styledFonts.Smaller.Face, utils.GrayText),
		}

		body = style.Layout(totalValue, pieItems...)
	case DisplayGraphs:
		items := make([]drawing.FlexChild, 0, len(uses.skills)*2-1+2)

		items = append(
			items,
			drawing.Rigid(exportTimeFrame(styledFonts, uses.timeController.CurrentTimeFrame)),
			drawing.FlexVSpacer(drawing.CommonSpacing),
		)

		for i, skill := range uses.skills {
			if i != 0 {
				items = append(items, drawing.FlexVSpacer(drawing.CommonSpacing))
			}

			style := drawing.StyleAreaChart(skill.chart, components.StringToColor(skill.name))
			style.MinHeight = 200

			items = append(items, drawing.Rigid(s.exportWidget(styledFonts, skill, style.Layout())))
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
		s.TabName(),
		drawing.HorizontalWrap{
			Alignment:   layout.Middle,
			Spacing:     drawing.CommonSpacing * 3,
			LineSpacing: drawing.CommonSpacing,
		}.Layout(
			styledFonts.Smaller.Layout(fmt.Sprintf("Subject: %v", s.currentSubject)),
			styledFonts.Smaller.Layout(fmt.Sprintf("Display: %v", s.currentDisplay)),
		),
		drawing.RoundedSurface(
			utils.SecondBG,
			body,
		),
	)

	return drawing.ExportImage(state.Theme(), base, drawing.F64(800, 10000))
}
