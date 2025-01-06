package collectors

import (
	"PGCombatTracker/abstract"
	"PGCombatTracker/ui/components"
	"PGCombatTracker/utils"
	"cmp"
	"fmt"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/unit"
	"gioui.org/widget"
	"log"
	"slices"
	"strings"
	"time"
)

type skillDamage struct {
	name     string
	amount   int
	damage   abstract.Vitals
	lastUsed time.Time
	chart    *components.TimeBasedChart
}

type subjectiveDamageDealt struct {
	subject        string
	timeController *components.TimeController
	maxRange       components.DataRange
	stackedChart   *components.StackedTimeBasedChart
	totalDamage    abstract.Vitals
	maxDamage      abstract.Vitals
	indirectDamage abstract.Vitals
	skillDamage    []skillDamage
}

func NewDamageDealtCollector() *DamageDealtCollector {
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

	timeController, err := components.NewTimeController(components.NewTimeBasedChart("Total"))
	if err != nil {
		log.Fatalln(err)
	}

	return &DamageDealtCollector{
		subjectDropdown: subjectDropdown,
		displayDropdown: displayDropdown,
		longFormatBool:  &widget.Bool{},
		total: subjectiveDamageDealt{
			timeController: timeController,
			stackedChart:   components.NewStackedTimeBasedChart(),
		},
	}
}

type DamageDealtCollector struct {
	currentSubject string
	total          subjectiveDamageDealt
	subjects       []subjectiveDamageDealt

	currentDisplay displayChoice

	subjectDropdown *components.Dropdown
	displayDropdown *components.Dropdown
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
	createSkillDamage := func(subject subjectiveDamageDealt) func() skillDamage {
		return func() skillDamage {
			chart := components.NewTimeBasedChart(skillName)
			chart.Add(components.TimePoint{
				Time:    event.Time,
				Value:   skillUse.Damage.Total(),
				Details: *skillUse.Damage,
			})
			subject.stackedChart.Add(chart, skillName)
			return skillDamage{
				name:     skillName,
				amount:   1,
				damage:   *skillUse.Damage,
				lastUsed: event.Time,
				chart:    chart,
			}
		}
	}
	updateSkillDamage := func(skill skillDamage) skillDamage {
		if skill.lastUsed != event.Time {
			skill.amount++
		}
		skill.damage = skill.damage.Add(*skillUse.Damage)
		skill.lastUsed = event.Time
		skill.chart.Add(components.TimePoint{
			Time:    event.Time,
			Value:   skill.damage.Total(),
			Details: skill.damage,
		})

		return skill
	}
	skillDamageMax := func(a, b skillDamage) int {
		return cmp.Compare(a.damage.Total(), b.damage.Total())
	}
	skillDamageSort := func(a, b skillDamage) int {
		return cmp.Compare(b.damage.Total(), a.damage.Total())
	}

	// Ingest total stuff
	d.total.totalDamage = d.total.totalDamage.Add(*skillUse.Damage)
	d.total.timeController.Add(components.TimePoint{
		Time:  event.Time,
		Value: d.total.totalDamage.Total(),
	})
	d.total.skillDamage = utils.CreateUpdate(
		d.total.skillDamage,
		findSkillDamage,
		createSkillDamage(d.total),
		updateSkillDamage,
	)
	d.total.maxDamage = slices.MaxFunc(d.total.skillDamage, skillDamageMax).damage
	d.total.maxRange = d.total.maxRange.Expand(d.total.maxDamage.Total())

	slices.SortFunc(d.total.skillDamage, skillDamageSort)

	// Ingest individual stuff
	d.subjects = utils.CreateUpdate(
		d.subjects,
		func(subject subjectiveDamageDealt) bool {
			return subject.subject == skillUse.Subject
		},
		func() subjectiveDamageDealt {
			controller, err := components.NewTimeController(components.NewTimeBasedChart("Total"))
			if err != nil {
				log.Fatalln(err)
			}

			controller.Add(components.TimePoint{
				Time:  event.Time,
				Value: skillUse.Damage.Total(),
			})

			subject := subjectiveDamageDealt{
				subject:     skillUse.Subject,
				totalDamage: *skillUse.Damage,
				maxDamage:   *skillUse.Damage,
				maxRange: components.DataRange{
					Max: skillUse.Damage.Total(),
				},
				timeController: controller,
				stackedChart:   components.NewStackedTimeBasedChart(),
			}

			subject.skillDamage = []skillDamage{
				createSkillDamage(subject)(),
			}

			return subject
		},
		func(subject subjectiveDamageDealt) subjectiveDamageDealt {
			subject.totalDamage = subject.totalDamage.Add(*skillUse.Damage)
			subject.timeController.Add(components.TimePoint{
				Time:  event.Time,
				Value: subject.totalDamage.Total(),
			})

			subject.skillDamage = utils.CreateUpdate(
				subject.skillDamage,
				findSkillDamage,
				createSkillDamage(subject),
				updateSkillDamage,
			)
			subject.maxDamage = slices.MaxFunc(subject.skillDamage, skillDamageMax).damage
			subject.maxRange = subject.maxRange.Expand(subject.maxDamage.Total())

			slices.SortFunc(subject.skillDamage, skillDamageSort)

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

func (d *DamageDealtCollector) Reset() {
	timeController, err := components.NewTimeController(components.NewTimeBasedChart("Total"))
	if err != nil {
		log.Fatalln(err)
	}

	d.subjects = nil
	d.total = subjectiveDamageDealt{
		timeController: timeController,
		stackedChart:   components.NewStackedTimeBasedChart(),
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
				)
			}),
			utils.FlexSpacerH(utils.CommonSpacing),
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				if d.currentDisplay == DisplayGraphs {
					return components.StyleTimeController(state.Theme(), subject.timeController).Layout(gtx)
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
		subject.stackedChart.DisplayTimeFrame = subject.timeController.CurrentTimeFrame
		subject.stackedChart.DisplayValueRange = subject.timeController.FullValueRange

		stackedStyle := components.StyleStackedTimeBasedChart(state.Theme(), subject.stackedChart)
		stackedStyle.Alpha = 255
		stackedStyle.MinHeight = 150
		stackedStyle.TextSize = 12
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
			skill.chart.DisplayTimeFrame = subject.timeController.CurrentTimeFrame
			skill.chart.DisplayValueRange = subject.maxRange
			chartStyle := components.StyleTimeBasedChart(state.Theme(), skill.chart)
			chartStyle.Color = components.StringToColor(skill.name)

			//widgets = append(widgets, func(gtx layout.Context) layout.Dimensions {
			//	return layout.Flex{
			//		Axis: layout.Vertical,
			//	}.Layout(
			//		gtx,
			//		layout.Rigid(chartStyle.Layout),
			//		utils.FlexSpacerH(utils.CommonSpacing),
			//		layout.Rigid(d.drawBar(state, skill, maxDamage, 40)),
			//	)
			//})

			widgets = append(widgets, d.drawWidget(state, skill, chartStyle.Layout, 100))
		}
	}

	return topWidget, widgets
}
