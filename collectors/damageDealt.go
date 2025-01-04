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
	totalChart     *components.TimeBasedChart
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

	return &DamageDealtCollector{
		subjectDropdown: subjectDropdown,
		longFormatBool:  &widget.Bool{},
		total: subjectiveDamageDealt{
			totalChart: components.NewTimeBasedChart(),
		},
	}
}

type DamageDealtCollector struct {
	currentSubject string
	total          subjectiveDamageDealt
	subjects       []subjectiveDamageDealt

	subjectDropdown *components.Dropdown
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
	createSkillDamage := func() skillDamage {
		chart := components.NewTimeBasedChart()
		chart.Add(components.TimePoint{
			Time:  event.Time,
			Value: skillUse.Damage.Total(),
		})
		return skillDamage{
			name:     skillName,
			amount:   1,
			damage:   *skillUse.Damage,
			lastUsed: event.Time,
			chart:    chart,
		}
	}
	updateSkillDamage := func(skill skillDamage) skillDamage {
		if skill.lastUsed != event.Time {
			skill.amount++
		}
		skill.damage = skill.damage.Add(*skillUse.Damage)
		skill.lastUsed = event.Time
		skill.chart.Add(components.TimePoint{
			Time:  event.Time,
			Value: skill.damage.Total(),
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
	d.total.totalChart.Add(components.TimePoint{
		Time:  event.Time,
		Value: d.total.totalDamage.Total(),
	})
	d.total.skillDamage = utils.CreateUpdate(
		d.total.skillDamage,
		findSkillDamage,
		createSkillDamage,
		updateSkillDamage,
	)
	d.total.maxDamage = slices.MaxFunc(d.total.skillDamage, skillDamageMax).damage

	slices.SortFunc(d.total.skillDamage, skillDamageSort)

	// Ingest individual stuff
	d.subjects = utils.CreateUpdate(
		d.subjects,
		func(subject subjectiveDamageDealt) bool {
			return subject.subject == skillUse.Subject
		},
		func() subjectiveDamageDealt {
			chart := components.NewTimeBasedChart()
			chart.Add(components.TimePoint{
				Time:  event.Time,
				Value: skillUse.Damage.Total(),
			})
			return subjectiveDamageDealt{
				subject:     skillUse.Subject,
				totalDamage: *skillUse.Damage,
				skillDamage: []skillDamage{
					createSkillDamage(),
				},
				totalChart: chart,
			}
		},
		func(subject subjectiveDamageDealt) subjectiveDamageDealt {
			subject.totalDamage = subject.totalDamage.Add(*skillUse.Damage)
			subject.totalChart.Add(components.TimePoint{
				Time:  event.Time,
				Value: subject.totalDamage.Total(),
			})

			subject.skillDamage = utils.CreateUpdate(
				subject.skillDamage,
				findSkillDamage,
				createSkillDamage,
				updateSkillDamage,
			)
			subject.maxDamage = slices.MaxFunc(subject.skillDamage, skillDamageMax).damage

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
	d.subjects = nil
	d.total = subjectiveDamageDealt{
		totalChart: components.NewTimeBasedChart(),
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

	topWidget := topBarSurface(func(gtx layout.Context) layout.Dimensions {
		if d.longFormatBool.Update(gtx) {
			gtx.Source.Execute(op.InvalidateCmd{})
		}

		return layout.Flex{
			Axis:      layout.Horizontal,
			Alignment: layout.Middle,
		}.Layout(
			gtx,
			layout.Rigid(defaultDropdownStyle(state, d.subjectDropdown).Layout),
			utils.FlexSpacerW(utils.CommonSpacing),
			layout.Rigid(defaultCheckboxStyle(state, d.longFormatBool, "Use long numbers").Layout),
		)
	})

	var widgets []layout.Widget

	subject := d.total
	for _, possibleSubject := range d.subjects {
		if possibleSubject.subject == d.currentSubject {
			subject = possibleSubject
			break
		}
	}

	//var maxDamage = subject.maxDamage.Total()

	//widgets = append(widgets, func(gtx layout.Context) layout.Dimensions {
	//	return layout.Flex{
	//		Axis: layout.Vertical,
	//	}.Layout(
	//		gtx,
	//		layout.Rigid(components.StyleTimeBasedChart(subject.totalChart).Layout),
	//		utils.FlexSpacerH(utils.CommonSpacing),
	//		layout.Rigid(d.drawBar(
	//			state,
	//			skillDamage{
	//				name:   "Total Damage",
	//				amount: 0,
	//				damage: subject.totalDamage,
	//			},
	//			maxDamage,
	//			25,
	//		)),
	//	)
	//})

	log.Println("screen update")

	totalChartStyle := components.StyleTimeBasedChart(subject.totalChart)
	totalChartStyle.Color = components.StringToColor("Total Damage")

	widgets = append(widgets, d.drawWidget(
		state,
		skillDamage{
			name:   "Total Damage",
			amount: 0,
			damage: subject.totalDamage,
		},
		totalChartStyle.Layout,
		100,
	))

	for _, skill := range subject.skillDamage {
		skill.chart.DataTimeFrame = subject.totalChart.DataTimeFrame
		chartStyle := components.StyleTimeBasedChart(skill.chart)
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

	return topWidget, widgets
}
