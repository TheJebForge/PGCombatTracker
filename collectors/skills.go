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

type skillUse struct {
	name     string
	amount   int
	damage   abstract.Vitals
	lastUsed time.Time
}

type subjectiveSkillUses struct {
	name    string
	skills  []skillUse
	maxUsed int
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

	return &SkillsCollector{
		subjectDropdown: subjectDropdown,
		longFormatBool:  &widget.Bool{},
	}
}

type SkillsCollector struct {
	allies   subjectiveSkillUses
	enemies  subjectiveSkillUses
	all      subjectiveSkillUses
	subjects []subjectiveSkillUses

	currentSubject  skillUseSubject
	subjectDropdown *components.Dropdown
	longFormatBool  *widget.Bool
}

func (s *SkillsCollector) Reset() {
	s.allies = subjectiveSkillUses{}
	s.enemies = subjectiveSkillUses{}
	s.all = subjectiveSkillUses{}
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
		return skillUse{
			name:     skillName,
			amount:   1,
			damage:   damage,
			lastUsed: event.Time,
		}
	}
	updateSkillUse := func(use skillUse) skillUse {
		if use.lastUsed != event.Time {
			use.amount++
		}
		use.damage = use.damage.Add(damage)
		use.lastUsed = event.Time

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
		slices.SortFunc(stats.skills, skillUseSort)
		stats.maxUsed = slices.MaxFunc(stats.skills, skillUseMax).amount
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
			return subjectiveSkillUses{
				name:    subject,
				maxUsed: 1,
				skills: []skillUse{
					createSkillUse(),
				},
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

	topWidget := topBarSurface(func(gtx layout.Context) layout.Dimensions {
		if s.longFormatBool.Update(gtx) {
			gtx.Source.Execute(op.InvalidateCmd{})
		}

		return layout.Flex{
			Axis:      layout.Horizontal,
			Alignment: layout.Middle,
		}.Layout(
			gtx,
			layout.Rigid(defaultDropdownStyle(state, s.subjectDropdown).Layout),
			utils.FlexSpacerW(utils.CommonSpacing),
			layout.Rigid(defaultCheckboxStyle(state, s.longFormatBool, "Use long numbers").Layout),
		)
	})

	var widgets []layout.Widget

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

	for _, skill := range uses.skills {
		widgets = append(widgets, s.drawBar(state, skill, uses.maxUsed, 40))
	}

	return topWidget, widgets
}
