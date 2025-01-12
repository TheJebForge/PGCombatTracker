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

type skillDamage struct {
	name          string
	amount        int
	damage        abstract.Vitals
	lastUsed      time.Time
	totalChart    *components.TimeBasedChart
	dpsChart      *components.TimeBasedChart
	dpsCalculator *DPSCalculator
}

type subjectiveDamageDealt struct {
	subject string

	// Total damage versions of charts
	totalChart        *components.TimeController
	totalMaxRange     components.DataRange
	stackedTotalChart *components.StackedTimeBasedChart

	// DPS versions of charts
	dpsChart        *components.TimeController
	stackedDpsChart *components.StackedTimeBasedChart
	dpsCalculator   *DPSCalculator

	totalDamage    abstract.Vitals
	maxDamage      abstract.Vitals
	indirectDamage abstract.Vitals
	skillDamage    []skillDamage
}

func NewDamageDealtCollector(settings *abstract.Settings) *DamageDealtCollector {
	subjectDropdown, err := components.NewDropdown("Subject", subjectChoice(""))
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

	chartDropdown, err := components.NewDropdown(
		"View",
		DamageChart,
		DPSChart,
	)

	dpsTimeController := components.NewTimeControllerOrCrash(components.NewTimeBasedChart("Total DPS"))

	return &DamageDealtCollector{
		subjectDropdown: subjectDropdown,
		displayDropdown: displayDropdown,
		chartDropdown:   chartDropdown,
		longFormatBool:  &widget.Bool{},
		total: subjectiveDamageDealt{
			totalChart:        components.NewTimeControllerOrCrash(components.NewTimeBasedChart("Total")),
			stackedTotalChart: components.NewStackedTimeBasedChart(),
			dpsChart:          dpsTimeController,
			stackedDpsChart:   components.NewStackedTimeBasedChart(),
			dpsCalculator:     NewDPSCalculatorForController(dpsTimeController, settings),
		},
	}
}

type DamageDealtCollector struct {
	currentSubject string
	total          subjectiveDamageDealt
	subjects       []subjectiveDamageDealt

	currentDisplay   displayChoice
	currentChartView damageChartChoice

	subjectDropdown *components.Dropdown
	displayDropdown *components.Dropdown
	chartDropdown   *components.Dropdown
	longFormatBool  *widget.Bool
}

func (d *DamageDealtCollector) ingestSkillDamage(info abstract.StatisticsInformation, event *abstract.ChatEvent) {
	skillUse := event.Contents.(*abstract.SkillUse)
	skillName := skillUse.Skill

	if info.Settings().RemoveLevelsFromSkills {
		skillName = SplitOffId(skillName)
	}

	// Functions for dealing with skillDamage
	findSkillDamage := func(skill skillDamage) bool {
		return skill.name == skillName
	}
	createSkillDamage := func(subject *subjectiveDamageDealt) func() skillDamage {
		return func() skillDamage {
			totalChart := components.NewTimeBasedChart(skillName)
			totalChart.Add(components.TimePoint{
				Time:    event.Time,
				Value:   skillUse.Damage.Total(),
				Details: *skillUse.Damage,
			})
			subject.stackedTotalChart.Add(totalChart, skillName)

			dpsChart := components.NewTimeBasedChart(skillName)
			dpsCalculator := NewDPSCalculatorForChart(dpsChart, info.Settings())
			dpsCalculator.Add(event.Time, skillUse.Damage.Total())
			subject.stackedDpsChart.Add(dpsChart, skillName)

			return skillDamage{
				name:          skillName,
				amount:        1,
				damage:        *skillUse.Damage,
				lastUsed:      event.Time,
				totalChart:    totalChart,
				dpsChart:      dpsChart,
				dpsCalculator: dpsCalculator,
			}
		}
	}
	updateSkillDamage := func(skill skillDamage) skillDamage {
		if skill.lastUsed != event.Time {
			skill.amount++
		}
		skill.damage = skill.damage.Add(*skillUse.Damage)
		skill.lastUsed = event.Time
		skill.totalChart.Add(components.TimePoint{
			Time:    event.Time,
			Value:   skill.damage.Total(),
			Details: skill.damage,
		})
		skill.dpsCalculator.Add(event.Time, skillUse.Damage.Total())

		return skill
	}
	skillDamageMax := func(a, b skillDamage) int {
		return cmp.Compare(a.damage.Total(), b.damage.Total())
	}
	skillDamageSort := func(a, b skillDamage) int {
		return cmp.Compare(b.damage.Total(), a.damage.Total())
	}

	processSubjectiveDD := func(subject *subjectiveDamageDealt) {
		subject.totalDamage = subject.totalDamage.Add(*skillUse.Damage)
		subject.totalChart.Add(components.TimePoint{
			Time:  event.Time,
			Value: subject.totalDamage.Total(),
		})
		subject.dpsCalculator.Add(event.Time, skillUse.Damage.Total())
		subject.skillDamage = utils.CreateUpdate(
			subject.skillDamage,
			findSkillDamage,
			createSkillDamage(subject),
			updateSkillDamage,
		)
		subject.maxDamage = slices.MaxFunc(subject.skillDamage, skillDamageMax).damage
		subject.totalMaxRange = subject.totalMaxRange.Expand(subject.maxDamage.Total())

		slices.SortFunc(subject.skillDamage, skillDamageSort)
	}

	// Ingest total stuff
	processSubjectiveDD(&d.total)

	// Ingest individual stuff
	d.subjects = utils.CreateUpdate(
		d.subjects,
		func(subject subjectiveDamageDealt) bool {
			return subject.subject == skillUse.Subject
		},
		func() subjectiveDamageDealt {
			totalChart, err := components.NewTimeController(components.NewTimeBasedChart("Total"))
			if err != nil {
				log.Fatalln(err)
			}

			totalChart.Add(components.TimePoint{
				Time:  event.Time,
				Value: skillUse.Damage.Total(),
			})

			dpsChart, err := components.NewTimeController(components.NewTimeBasedChart("Total DPS"))
			if err != nil {
				log.Fatalln(err)
			}
			dpsCalculator := NewDPSCalculatorForController(dpsChart, info.Settings())

			subject := subjectiveDamageDealt{
				subject:     skillUse.Subject,
				totalDamage: *skillUse.Damage,
				maxDamage:   *skillUse.Damage,
				totalMaxRange: components.DataRange{
					Max: skillUse.Damage.Total(),
				},
				totalChart:        totalChart,
				stackedTotalChart: components.NewStackedTimeBasedChart(),
				dpsChart:          dpsChart,
				stackedDpsChart:   components.NewStackedTimeBasedChart(),
				dpsCalculator:     dpsCalculator,
			}

			subject.skillDamage = []skillDamage{
				createSkillDamage(&subject)(),
			}

			return subject
		},
		func(subject subjectiveDamageDealt) subjectiveDamageDealt {
			processSubjectiveDD(&subject)
			return subject
		},
	)

	// Add victims to dropdown
	d.subjectDropdown.SetOptions(utils.CreateUpdate(
		d.subjectDropdown.Options(),
		func(item fmt.Stringer) bool {
			casted := item.(subjectChoice)
			return string(casted) == skillUse.Subject
		},
		func() fmt.Stringer {
			return subjectChoice(skillUse.Subject)
		},
		func(stringer fmt.Stringer) fmt.Stringer {
			return stringer
		},
	))
}

func (d *DamageDealtCollector) ingestIndirect(event *abstract.ChatEvent) {
	indirect := event.Contents.(*abstract.IndirectDamage)
	d.total.indirectDamage = d.total.indirectDamage.Add(indirect.Damage.Abs())
}

func (d *DamageDealtCollector) Reset(info abstract.StatisticsInformation) {
	dpsTimeController := components.NewTimeControllerOrCrash(components.NewTimeBasedChart("Total DPS"))

	d.subjects = nil
	d.total = subjectiveDamageDealt{
		totalChart:        components.NewTimeControllerOrCrash(components.NewTimeBasedChart("Total")),
		stackedTotalChart: components.NewStackedTimeBasedChart(),
		dpsChart:          dpsTimeController,
		stackedDpsChart:   components.NewStackedTimeBasedChart(),
		dpsCalculator:     NewDPSCalculatorForController(dpsTimeController, info.Settings()),
	}
	d.subjectDropdown.SetOptions([]fmt.Stringer{subjectChoice("")})
	d.currentSubject = ""
}

func isSubjectValuable(info abstract.StatisticsInformation, subject, skill string) bool {
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

func (d *DamageDealtCollector) Tick(info abstract.StatisticsInformation, at time.Time) {
	d.total.dpsCalculator.Tick(at)
	for _, skill := range d.total.skillDamage {
		skill.dpsCalculator.Tick(at)
	}

	for _, subject := range d.subjects {
		subject.dpsCalculator.Tick(at)
		for _, skill := range subject.skillDamage {
			skill.dpsCalculator.Tick(at)
		}
	}
}

func (d *DamageDealtCollector) Collect(info abstract.StatisticsInformation, event *abstract.ChatEvent) error {
	if skillUse, ok := event.Contents.(*abstract.SkillUse); ok && skillUse.Damage != nil && isSubjectValuable(info, skillUse.Subject, skillUse.Skill) {
		d.ingestSkillDamage(info, event)
	}

	if indirect, ok := event.Contents.(*abstract.IndirectDamage); ok && !isSubjectValuable(info, indirect.Subject, "") {
		d.ingestIndirect(event)
	}

	return nil
}

func (d *DamageDealtCollector) TabName() string {
	return "Damage Dealt"
}

func (d *DamageDealtCollector) drawWidget(state abstract.LayeredState, skill skillDamage, widget layout.Widget, size unit.Dp) layout.Widget {
	return drawUniversalStatsText(
		state, skill.damage,
		widget, skill.amount,
		skill.name, "used %v times",
		size, d.longFormatBool.Value,
	)
}

func (d *DamageDealtCollector) drawBar(state abstract.LayeredState, skill skillDamage, maxDamage int, size unit.Dp) layout.Widget {
	return drawUniversalBar(
		state, skill.damage,
		skill.damage.Total(), maxDamage, skill.amount,
		skill.name, "used %v times",
		size, d.longFormatBool.Value,
	)
}

var nowLocation = time.Now().Location()

func (d *DamageDealtCollector) UI(state abstract.LayeredState) (layout.Widget, []layout.Widget) {
	if d.subjectDropdown.Changed() {
		d.currentSubject = string(d.subjectDropdown.Value.(subjectChoice))
	}

	if d.displayDropdown.Changed() {
		d.currentDisplay = d.displayDropdown.Value.(displayChoice)
	}

	if d.chartDropdown.Changed() {
		d.currentChartView = d.chartDropdown.Value.(damageChartChoice)
	}

	subject := d.total
	for _, possibleSubject := range d.subjects {
		if possibleSubject.subject == d.currentSubject {
			subject = possibleSubject
			break
		}
	}

	topWidget := topBarSurface(func(gtx layout.Context) layout.Dimensions {
		if d.longFormatBool.Update(gtx) {
			gtx.Source.Execute(op.InvalidateCmd{})
		}

		return layout.Flex{
			Axis: layout.Vertical,
		}.Layout(
			gtx,
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return components.HorizontalWrap{
					Alignment:   layout.Middle,
					Spacing:     utils.CommonSpacing * 2,
					LineSpacing: utils.CommonSpacing,
				}.Layout(
					gtx,
					defaultDropdownStyle(state, d.subjectDropdown).Layout,
					defaultCheckboxStyle(state, d.longFormatBool, "Use long numbers").Layout,
					defaultDropdownStyle(state, d.displayDropdown).Layout,
					func(gtx layout.Context) layout.Dimensions {
						if d.currentDisplay == DisplayGraphs {
							return defaultDropdownStyle(state, d.chartDropdown).Layout(gtx)
						}

						return layout.Dimensions{}
					},
				)
			}),
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				if d.currentDisplay == DisplayGraphs {
					return layout.Flex{
						Axis: layout.Vertical,
					}.Layout(
						gtx,
						utils.FlexSpacerH(utils.CommonSpacing),
						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							if d.currentDisplay == DisplayGraphs {
								if d.currentChartView == DamageChart {
									return components.StyleTimeController(state.Theme(), subject.totalChart).Layout(gtx)
								} else {
									return components.StyleTimeController(state.Theme(), subject.dpsChart).Layout(gtx)
								}
							}

							return layout.Dimensions{}
						}),
					)
				}

				return layout.Dimensions{}
			}),
		)

	})

	var widgets []layout.Widget

	switch d.currentDisplay {
	case DisplayBars:
		var maxDamage = subject.maxDamage.Total()

		widgets = append(widgets, d.drawBar(
			state,
			skillDamage{
				name:   "Total Damage",
				amount: 0,
				damage: subject.totalDamage,
			},
			maxDamage,
			25,
		))

		for _, skill := range subject.skillDamage {
			widgets = append(widgets, d.drawBar(state, skill, maxDamage, 40))
		}

		widgets = append(widgets, d.drawBar(
			state,
			skillDamage{
				name:   "Indirect Damage",
				amount: 0,
				damage: subject.indirectDamage,
			},
			subject.indirectDamage.Total(),
			25,
		))
	case DisplayPie:
		widgets = append(widgets, func(gtx layout.Context) layout.Dimensions {
			totalValue := subject.totalDamage.Total()

			pieItems := make([]components.PieChartItem, len(subject.skillDamage))
			for i, skill := range subject.skillDamage {
				pieItems[i] = components.PieChartItem{
					Name:    skill.name,
					Value:   skill.damage.Total(),
					SubText: skill.damage.StringCL(d.longFormatBool.Value),
				}
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
		stackedChart := subject.stackedTotalChart
		if d.currentChartView == DPSChart {
			stackedChart = subject.stackedDpsChart
		}

		controller := subject.totalChart
		if d.currentChartView == DPSChart {
			controller = subject.dpsChart
		}

		stackedChart.DisplayTimeFrame = controller.CurrentTimeFrame
		stackedChart.DisplayValueRange = controller.FullValueRange

		stackedStyle := components.StyleStackedTimeBasedChart(state.Theme(), stackedChart)
		stackedStyle.Alpha = 255
		stackedStyle.MinHeight = 150
		stackedStyle.TextSize = 12
		stackedStyle.LongFormat = d.longFormatBool.Value
		widgets = append(widgets, func(gtx layout.Context) layout.Dimensions {
			return layout.Flex{
				Axis: layout.Vertical,
			}.Layout(
				gtx,
				layout.Rigid(stackedStyle.Layout),
				utils.FlexSpacerH(utils.CommonSpacing),
			)
		})

		for _, skill := range subject.skillDamage {
			skillChart := skill.totalChart
			if d.currentChartView == DPSChart {
				skillChart = skill.dpsChart
			}

			skillChart.DisplayTimeFrame = controller.CurrentTimeFrame

			skillChart.DisplayValueRange = subject.totalMaxRange
			if d.currentChartView == DPSChart {
				skillChart.DisplayValueRange = controller.FullValueRange
			}

			chartStyle := components.StyleTimeBasedChart(state.Theme(), skillChart)
			chartStyle.Color = components.StringToColor(skill.name)
			chartStyle.LongFormat = d.longFormatBool.Value

			widgets = append(widgets, d.drawWidget(state, skill, chartStyle.Layout, 100))
		}
	}

	return topWidget, widgets
}

func (d *DamageDealtCollector) exportWidget(styledFonts *drawing.StyledFontPack, skill skillDamage, widget drawing.Widget) drawing.Widget {
	return exportUniversalStatsTextAsStack(
		styledFonts, skill.damage,
		widget, skill.amount,
		skill.name, "used %v times",
		d.longFormatBool.Value,
	)
}

func (d *DamageDealtCollector) exportBar(styledFonts *drawing.StyledFontPack, skill skillDamage, maxDamage int) drawing.Widget {
	return exportUniversalBar(
		styledFonts, skill.damage,
		skill.damage.Total(), maxDamage, skill.amount,
		skill.name, "used %v times",
		d.longFormatBool.Value,
	)
}

func (d *DamageDealtCollector) Export(state abstract.LayeredState) image.Image {
	subject := d.total
	for _, possibleSubject := range d.subjects {
		if possibleSubject.subject == d.currentSubject {
			subject = possibleSubject
			break
		}
	}

	styledFonts := drawing.StyleFontPack(state.FontPack(), state.Theme().Fg)

	body := drawing.Empty

	switch d.currentDisplay {
	case DisplayBars:
		items := make([]drawing.FlexChild, 0, len(subject.skillDamage)*2-1+4)

		maxDamage := subject.maxDamage.Total()

		items = append(
			items,
			drawing.Rigid(d.exportBar(
				styledFonts,
				skillDamage{
					name:   "Total Damage",
					amount: 0,
					damage: subject.totalDamage,
				},
				subject.totalDamage.Total(),
			)),
			drawing.FlexVSpacer(drawing.CommonSpacing),
		)

		for i, skill := range subject.skillDamage {
			if i != 0 {
				items = append(items, drawing.FlexVSpacer(drawing.CommonSpacing))
			}

			items = append(items, drawing.Rigid(d.exportBar(styledFonts, skill, maxDamage)))
		}

		items = append(
			items,
			drawing.FlexVSpacer(drawing.CommonSpacing),
			drawing.Rigid(d.exportBar(
				styledFonts,
				skillDamage{
					name:   "Indirect Damage",
					amount: 0,
					damage: subject.indirectDamage,
				},
				subject.indirectDamage.Total(),
			)),
		)

		body = drawing.Flex{
			ExpandW: true,
			Axis:    layout.Vertical,
		}.Layout(
			items...,
		)
	case DisplayPie:
		totalValue := subject.totalDamage.Total()

		pieItems := make([]drawing.PieChartItem, len(subject.skillDamage))
		for i, skill := range subject.skillDamage {
			pieItems[i] = drawing.PieChartItem{
				Name:    skill.name,
				Value:   skill.damage.Total(),
				SubText: skill.damage.StringCL(d.longFormatBool.Value),
			}
		}

		style := drawing.PieChart{
			OverflowLimit: 15,
			ColorBoxSize:  drawing.CommonSpacing * 4,
			TextStyle:     styledFonts.Body,
			SubTextStyle:  drawing.MakeTextStyle(styledFonts.Smaller.Face, utils.GrayText),
		}

		body = style.Layout(totalValue, pieItems...)
	case DisplayGraphs:
		items := make([]drawing.FlexChild, 0, len(subject.skillDamage)*2-1+6)

		items = append(
			items,
			drawing.Rigid(exportTimeFrame(styledFonts, subject.totalChart.CurrentTimeFrame)),
			drawing.FlexVSpacer(drawing.CommonSpacing),
		)

		items = append(
			items,
			drawing.Rigid(d.exportBar(
				styledFonts,
				skillDamage{
					name:   "Total Damage",
					amount: 0,
					damage: subject.totalDamage,
				},
				subject.totalDamage.Total(),
			)),
			drawing.FlexVSpacer(drawing.CommonSpacing),
		)

		stackedChart := subject.stackedTotalChart
		if d.currentChartView == DPSChart {
			stackedChart = subject.stackedDpsChart
		}

		items = append(
			items,
			drawing.Rigid(
				drawing.StyleStackedAreaChart(styledFonts, stackedChart).Layout(),
			),
			drawing.FlexVSpacer(drawing.CommonSpacing),
		)

		for i, skill := range subject.skillDamage {
			if i != 0 {
				items = append(items, drawing.FlexVSpacer(drawing.CommonSpacing))
			}

			skillChart := skill.totalChart
			if d.currentChartView == DPSChart {
				skillChart = skill.dpsChart
			}

			style := drawing.StyleAreaChart(skillChart, components.StringToColor(skill.name))
			style.MinHeight = 200

			items = append(items, drawing.Rigid(d.exportWidget(styledFonts, skill, style.Layout())))
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
		d.TabName(),
		drawing.HorizontalWrap{
			Alignment:   layout.Middle,
			Spacing:     drawing.CommonSpacing * 3,
			LineSpacing: drawing.CommonSpacing,
		}.Layout(
			func(ltx drawing.Context) drawing.Result {
				if d.currentSubject != "" {
					return styledFonts.Smaller.Layout(fmt.Sprintf("Subject: %v", d.currentSubject))(ltx)
				}

				return styledFonts.Smaller.Layout("Subject: All")(ltx)
			},
			styledFonts.Smaller.Layout(fmt.Sprintf("Display: %v", d.currentDisplay)),
			func(ltx drawing.Context) drawing.Result {
				if d.currentDisplay == DisplayGraphs {
					return styledFonts.Smaller.Layout(fmt.Sprintf("%v Chart", d.currentChartView))(ltx)
				}

				return drawing.Empty(ltx)
			},
		),
		drawing.RoundedSurface(
			utils.SecondBG,
			body,
		),
	)

	return drawing.ExportImage(state.Theme(), base, drawing.F64(800, 10000))
}
