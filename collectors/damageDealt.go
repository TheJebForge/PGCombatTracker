package collectors

import (
	"PGCombatTracker/abstract"
	"PGCombatTracker/ui/components"
	"PGCombatTracker/utils"
	"cmp"
	"fmt"
	"gioui.org/layout"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"image"
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
}

type subjectiveDamageDealt struct {
	subject        string
	totalDamage    abstract.Vitals
	maxDamage      abstract.Vitals
	indirectDamage abstract.Vitals
	skillDamage    []skillDamage
}

type SubjectChoice string

func (s SubjectChoice) String() string {
	if s == "" {
		return "All"
	}
	return string(s)
}

func NewDamageDealtCollector() *DamageDealtCollector {
	subjectDropdown, err := components.NewDropdown("Subject", SubjectChoice(""))
	if err != nil {
		log.Fatalln(err)
	}

	return &DamageDealtCollector{
		subjectDropdown: subjectDropdown,
		longFormatBool:  &widget.Bool{},
	}
}

type DamageDealtCollector struct {
	currentSubject string
	total          subjectiveDamageDealt
	subjects       []subjectiveDamageDealt

	subjectDropdown *components.Dropdown
	longFormatBool  *widget.Bool
}

func (col *DamageDealtCollector) ingestSkillDamage(info abstract.StatisticsInformation, event *abstract.ChatEvent) {
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
		return skillDamage{
			name:     skillName,
			amount:   1,
			damage:   *skillUse.Damage,
			lastUsed: event.Time,
		}
	}
	updateSkillDamage := func(skill skillDamage) skillDamage {
		if skill.lastUsed != event.Time {
			skill.amount++
		}
		skill.damage = skill.damage.Add(*skillUse.Damage)
		skill.lastUsed = event.Time

		return skill
	}
	skillDamageMax := func(a, b skillDamage) int {
		return cmp.Compare(a.damage.Total(), b.damage.Total())
	}
	skillDamageSort := func(a, b skillDamage) int {
		return cmp.Compare(b.damage.Total(), a.damage.Total())
	}

	// Ingest total stuff
	col.total.totalDamage = col.total.totalDamage.Add(*skillUse.Damage)
	col.total.skillDamage = utils.CreateUpdate(
		col.total.skillDamage,
		findSkillDamage,
		createSkillDamage,
		updateSkillDamage,
	)
	col.total.maxDamage = slices.MaxFunc(col.total.skillDamage, skillDamageMax).damage

	slices.SortFunc(col.total.skillDamage, skillDamageSort)

	// Ingest individual stuff
	col.subjects = utils.CreateUpdate(
		col.subjects,
		func(subject subjectiveDamageDealt) bool {
			return subject.subject == skillUse.Subject
		},
		func() subjectiveDamageDealt {
			return subjectiveDamageDealt{
				subject:     skillUse.Subject,
				totalDamage: *skillUse.Damage,
				skillDamage: []skillDamage{
					{
						name:     skillName,
						amount:   1,
						damage:   *skillUse.Damage,
						lastUsed: event.Time,
					},
				},
			}
		},
		func(subject subjectiveDamageDealt) subjectiveDamageDealt {
			subject.totalDamage = subject.totalDamage.Add(*skillUse.Damage)
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

	// Add subjects to dropdown
	col.subjectDropdown.SetOptions(utils.CreateUpdate(
		col.subjectDropdown.Options(),
		func(item fmt.Stringer) bool {
			casted := item.(SubjectChoice)
			return string(casted) == skillUse.Subject
		},
		func() fmt.Stringer {
			return SubjectChoice(skillUse.Subject)
		},
		func(stringer fmt.Stringer) fmt.Stringer {
			return stringer
		},
	))
}

func (col *DamageDealtCollector) ingestIndirect(event *abstract.ChatEvent) {
	indirect := event.Contents.(*abstract.IndirectDamage)
	col.total.indirectDamage = col.total.indirectDamage.Add(indirect.Damage.Abs())
}

func (col *DamageDealtCollector) Reset() {
	col.subjects = nil
	col.total = subjectiveDamageDealt{}
	col.subjectDropdown.SetOptions([]fmt.Stringer{SubjectChoice("")})
	col.currentSubject = ""
}

func isSkillUseSubjectValuable(info abstract.StatisticsInformation, subject, skill string) bool {
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

func (col *DamageDealtCollector) Collect(info abstract.StatisticsInformation, event *abstract.ChatEvent) error {
	if skillUse, ok := event.Contents.(*abstract.SkillUse); ok && skillUse.Damage != nil && isSkillUseSubjectValuable(info, skillUse.Subject, skillUse.Skill) {
		col.ingestSkillDamage(info, event)
	}

	if indirect, ok := event.Contents.(*abstract.IndirectDamage); ok && !isSkillUseSubjectValuable(info, indirect.Subject, "") {
		col.ingestIndirect(event)
	}

	return nil
}

func (col *DamageDealtCollector) TabName() string {
	return "Damage Dealt"
}

func (col *DamageDealtCollector) drawBar(state abstract.LayeredState, skill skillDamage, maxDamage int, size unit.Dp) layout.Widget {
	return func(gtx layout.Context) layout.Dimensions {
		var progress = float64(skill.damage.Total()) / float64(maxDamage)

		return components.Canvas{
			ExpandHorizontal: true,
			MinSize: image.Point{
				Y: gtx.Dp(size + utils.CommonSpacing),
			},
		}.Layout(
			gtx,
			components.CanvasItem{
				Anchor: layout.N,
				Widget: components.BarWidget(components.StringToColor(skill.name), size, progress),
			},
			components.CanvasItem{
				Anchor: layout.W,
				Offset: image.Point{
					X: gtx.Dp(utils.CommonSpacing),
					Y: gtx.Dp(-2.5),
				},
				Widget: func(gtx layout.Context) layout.Dimensions {
					if skill.amount == 0 {
						return material.Label(state.Theme(), 12, skill.name).Layout(gtx)
					} else {
						return layout.Flex{
							Axis: layout.Vertical,
						}.Layout(
							gtx,
							layout.Rigid(material.Label(state.Theme(), 12, skill.name).Layout),
							layout.Rigid(material.Label(state.Theme(), 12, fmt.Sprintf("used %v times",
								skill.amount)).Layout),
						)
					}
				},
			},
			components.CanvasItem{
				Anchor: layout.E,
				Offset: image.Point{
					X: gtx.Dp(-utils.CommonSpacing),
					Y: gtx.Dp(-2.5),
				},
				Widget: material.Label(state.Theme(), 12, skill.damage.StringCL(col.longFormatBool.Value)).Layout,
			},
		)
	}
}

func (col *DamageDealtCollector) UI(state abstract.LayeredState) (layout.Widget, []layout.Widget) {
	if col.subjectDropdown.Changed() {
		col.currentSubject = string(col.subjectDropdown.Value.(SubjectChoice))
	}

	topWidget := topBarSurface(func(gtx layout.Context) layout.Dimensions {
		return layout.Flex{
			Axis:      layout.Horizontal,
			Alignment: layout.Middle,
		}.Layout(
			gtx,
			layout.Rigid(defaultDropdownStyle(state, col.subjectDropdown).Layout),
			utils.FlexSpacerW(utils.CommonSpacing),
			layout.Rigid(defaultCheckboxStyle(state, col.longFormatBool, "Use long numbers").Layout),
		)
	})

	var widgets []layout.Widget

	subject := col.total
	for _, possibleSubject := range col.subjects {
		if possibleSubject.subject == col.currentSubject {
			subject = possibleSubject
			break
		}
	}

	var maxDamage = subject.maxDamage.Total()

	widgets = append(widgets, col.drawBar(
		state,
		skillDamage{
			name:   "Total Damage",
			amount: 0,
			damage: col.total.totalDamage,
		},
		maxDamage,
		25,
	))

	for _, skill := range subject.skillDamage {
		widgets = append(widgets, col.drawBar(state, skill, maxDamage, 40))
	}

	widgets = append(widgets, col.drawBar(
		state,
		skillDamage{
			name:   "Indirect Damage",
			amount: 0,
			damage: col.total.indirectDamage,
		},
		col.total.indirectDamage.Total(),
		25,
	))

	return topWidget, widgets
}
