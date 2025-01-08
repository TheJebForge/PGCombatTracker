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
	"time"
)

type skillXP struct {
	name   string
	xp     int
	levels int
}

type subjectiveSkillsXP struct {
	name   string
	skills []skillXP
	maxXP  int
}

func NewLevelingCollector() *LevelingCollector {
	subjectDropdown, err := components.NewDropdown(
		"Subject",
		subjectChoice(""),
	)
	if err != nil {
		log.Fatalln(err)
	}

	return &LevelingCollector{
		subjectDropdown: subjectDropdown,
		longFormatBool:  &widget.Bool{},
	}
}

type LevelingCollector struct {
	all      subjectiveSkillsXP
	subjects []subjectiveSkillsXP

	currentSubject  string
	subjectDropdown *components.Dropdown
	longFormatBool  *widget.Bool
}

func (l *LevelingCollector) Reset(info abstract.StatisticsInformation) {
	l.all = subjectiveSkillsXP{}
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

			return skillXP{
				name:   skillName,
				xp:     gainedXP,
				levels: level,
			}
		}
	}
	updateSkillXp := func(leveled bool) func(skillXP) skillXP {
		return func(skill skillXP) skillXP {
			skill.xp += gainedXP
			if leveled {
				skill.levels++
			}

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
		slices.SortFunc(stats.skills, skillXpSort)
		stats.maxXP = slices.MaxFunc(stats.skills, skillXpMax).xp
	}
	processSubjects := func(leveled bool) {
		l.subjects = utils.CreateUpdate(
			l.subjects,
			func(subject subjectiveSkillsXP) bool {
				return subject.name == info.CurrentUsername()
			},
			func() subjectiveSkillsXP {
				return subjectiveSkillsXP{
					name: info.CurrentUsername(),
					skills: []skillXP{
						createSkillXp(leveled)(),
					},
					maxXP: gainedXP,
				}
			},
			func(subject subjectiveSkillsXP) subjectiveSkillsXP {
				processSubjectiveSkillXp(&subject, leveled)
				return subject
			},
		)
	}

	switch {
	case xpOk:
		processSubjectiveSkillXp(&l.all, false)
		processSubjects(false)
	case levelOk:
		processSubjectiveSkillXp(&l.all, true)
		processSubjects(true)
	default:
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

	topWidget := topBarSurface(func(gtx layout.Context) layout.Dimensions {
		if l.longFormatBool.Update(gtx) {
			gtx.Source.Execute(op.InvalidateCmd{})
		}

		return layout.Flex{
			Axis:      layout.Horizontal,
			Alignment: layout.Middle,
		}.Layout(
			gtx,
			layout.Rigid(defaultDropdownStyle(state, l.subjectDropdown).Layout),
			utils.FlexSpacerW(utils.CommonSpacing),
			layout.Rigid(defaultCheckboxStyle(state, l.longFormatBool, "Use long numbers").Layout),
		)
	})

	var widgets []layout.Widget

	subject := l.all
	for _, possibleSubject := range l.subjects {
		if possibleSubject.name == l.currentSubject {
			subject = possibleSubject
			break
		}
	}

	for _, skill := range subject.skills {
		widgets = append(widgets, l.drawBar(state, skill, subject.maxXP, 40))
	}

	return topWidget, widgets
}
