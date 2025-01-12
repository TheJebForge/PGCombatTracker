package collectors

import (
	"PGCombatTracker/abstract"
	"PGCombatTracker/ui/components"
	"PGCombatTracker/utils"
	"PGCombatTracker/utils/drawing"
	"fmt"
	"gioui.org/layout"
	"image"
	"log"
	"strings"
	"time"
)

type subjectiveMisc struct {
	name string

	coinsFound         int
	coinsReceived      int
	errorCount         int
	killedCount        int
	critCount          int
	enemyCrits         int
	enemyEvasions      int
	evadedCount        int
	noDamageCount      int
	enemyNoDamageCount int
	deathCount         int
}

func NewMiscCollector() *MiscCollector {
	subjectDropdown, err := components.NewDropdown(
		"Subject",
		subjectChoice(""),
	)
	if err != nil {
		log.Fatalln(err)
	}

	return &MiscCollector{
		subjectDropdown: subjectDropdown,
	}
}

type MiscCollector struct {
	all      subjectiveMisc
	subjects []subjectiveMisc

	currentSubject  string
	subjectDropdown *components.Dropdown
}

func (m *MiscCollector) Reset(info abstract.StatisticsInformation) {
	m.all = subjectiveMisc{}
	m.subjects = nil
	m.subjectDropdown.SetOptions([]fmt.Stringer{subjectChoice("")})
	m.currentSubject = ""
}

func (m *MiscCollector) updateData(subject string, updateFunc func(misc subjectiveMisc) subjectiveMisc) {
	m.all = updateFunc(m.all)
	m.subjects = utils.CreateUpdate(
		m.subjects,
		func(misc subjectiveMisc) bool {
			return misc.name == subject
		},
		func() subjectiveMisc {
			misc := subjectiveMisc{
				name: subject,
			}
			return updateFunc(misc)
		},
		updateFunc,
	)
}

func (m *MiscCollector) isAlly(info abstract.StatisticsInformation, subject, skill string) bool {
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

func (m *MiscCollector) Tick(info abstract.StatisticsInformation, at time.Time) {

}

func (m *MiscCollector) Collect(info abstract.StatisticsInformation, event *abstract.ChatEvent) error {
	if collected, ok := event.Contents.(*abstract.FoundCoins); ok {
		m.updateData(info.CurrentUsername(), func(misc subjectiveMisc) subjectiveMisc {
			misc.coinsFound += collected.Coins
			return misc
		})
	} else if received, ok := event.Contents.(*abstract.ReceivedCoins); ok {
		m.updateData(info.CurrentUsername(), func(misc subjectiveMisc) subjectiveMisc {
			misc.coinsReceived += received.Coins
			return misc
		})
	} else if _, ok := event.Contents.(*abstract.ErrorLine); ok {
		m.updateData(info.CurrentUsername(), func(misc subjectiveMisc) subjectiveMisc {
			misc.errorCount++
			return misc
		})
	} else if skill, ok := event.Contents.(*abstract.SkillUse); ok && m.isAlly(info, skill.Subject, skill.Skill) {
		m.updateData(info.CurrentUsername(), func(misc subjectiveMisc) subjectiveMisc {
			switch {
			case skill.Fatality:
				misc.killedCount++
			case skill.Crit:
				misc.critCount++
			case skill.Evaded:
				misc.enemyEvasions++
			case skill.Damage != nil && skill.Damage.Total() == 0:
				misc.noDamageCount++
			}

			return misc
		})
	} else if skill, ok := event.Contents.(*abstract.SkillUse); ok && skill.Victim == info.CurrentUsername() {
		m.updateData(info.CurrentUsername(), func(misc subjectiveMisc) subjectiveMisc {
			switch {
			case skill.Fatality:
				misc.deathCount++
			case skill.Evaded:
				misc.evadedCount++
			case skill.Crit:
				misc.enemyCrits++
			case skill.Damage != nil && skill.Damage.Total() == 0:
				misc.enemyNoDamageCount++
			}

			return misc
		})
	} else {
		return nil
	}

	m.subjectDropdown.SetOptions(utils.CreateUpdate(
		m.subjectDropdown.Options(),
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

func (m *MiscCollector) TabName() string {
	return "Misc"
}

func addSubjectLabels[T any](subject *subjectiveMisc, label func(format string, args ...any) T) []T {
	return []T{
		label("Found %d coins", subject.coinsFound),
		label("Received %d coins", subject.coinsReceived),
		label("%d errors noticed", subject.errorCount),
		label("%d enemies killed", subject.killedCount),
		label("%d critical attacks", subject.critCount),
		label("%d times attacks did no damage", subject.noDamageCount),
		label("%d enemy attacks evaded", subject.evadedCount),
		label("%d enemy crits on subject", subject.enemyCrits),
		label("%d attacks enemies evaded", subject.enemyEvasions),
		label("%d times enemy attacks did no damage", subject.enemyNoDamageCount),
		label("%d times died", subject.deathCount),
	}
}

func (m *MiscCollector) UI(state abstract.LayeredState) (layout.Widget, []layout.Widget) {
	if m.subjectDropdown.Changed() {
		m.currentSubject = string(m.subjectDropdown.Value.(subjectChoice))
	}

	topWidget := topBarSurface(func(gtx layout.Context) layout.Dimensions {
		return layout.Flex{
			Axis:      layout.Horizontal,
			Alignment: layout.Middle,
		}.Layout(
			gtx,
			layout.Rigid(defaultDropdownStyle(state, m.subjectDropdown).Layout),
		)
	})

	var widgets []layout.Widget

	subject := m.all
	for _, possibleSubject := range m.subjects {
		if possibleSubject.name == m.currentSubject {
			subject = possibleSubject
			break
		}
	}

	label := func(format string, args ...any) layout.Widget {
		text := fmt.Sprintf(format, args...)

		return func(gtx layout.Context) layout.Dimensions {
			return layout.UniformInset(utils.CommonSpacing).Layout(
				gtx,
				func(gtx layout.Context) layout.Dimensions {
					style := defaultLabelStyle(state, text)
					style.TextSize = 14
					return style.Layout(gtx)
				},
			)
		}
	}

	widgets = append(
		widgets,
		addSubjectLabels(&subject, label)...,
	)

	return topWidget, widgets
}

func (m *MiscCollector) Export(state abstract.LayeredState) image.Image {
	subject := m.all
	for _, possibleSubject := range m.subjects {
		if possibleSubject.name == m.currentSubject {
			subject = possibleSubject
			break
		}
	}

	styledFonts := drawing.StyleFontPack(state.FontPack(), state.Theme().Fg)

	var flexItems []drawing.FlexChild

	label := func(format string, args ...any) drawing.FlexChild {
		text := fmt.Sprintf(format, args...)

		return drawing.Rigid(drawing.UniformInset(drawing.CommonSpacing).Layout(
			styledFonts.Body.Layout(text),
		))
	}

	flexItems = append(
		flexItems,
		addSubjectLabels(&subject, label)...,
	)

	body := drawing.Flex{
		Axis:    layout.Vertical,
		ExpandW: true,
	}.Layout(
		flexItems...,
	)

	base := layoutTitle(
		styledFonts,
		m.TabName(),
		drawing.HorizontalWrap{
			Alignment:   layout.Middle,
			Spacing:     drawing.CommonSpacing * 3,
			LineSpacing: drawing.CommonSpacing,
		}.Layout(
			func(ltx drawing.Context) drawing.Result {
				if m.currentSubject != "" {
					return styledFonts.Smaller.Layout(fmt.Sprintf("Subject: %v", m.currentSubject))(ltx)
				}

				return styledFonts.Smaller.Layout("Subject: All")(ltx)
			},
		),
		drawing.RoundedSurface(
			utils.SecondBG,
			body,
		),
	)

	return drawing.ExportImage(state.Theme(), base, drawing.F64(800, 10000))
}
