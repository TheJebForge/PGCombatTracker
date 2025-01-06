package collectors

import (
	"PGCombatTracker/abstract"
	"PGCombatTracker/ui/components"
	"PGCombatTracker/utils"
	"cmp"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/unit"
	"gioui.org/widget"
	"log"
	"slices"
	"strings"
	"time"
)

type healing struct {
	name      string
	amount    int
	recovered abstract.Vitals
}

type healingWithMax struct {
	subjects []healing
	max      abstract.Vitals
}

type healingSubject int

const (
	RecAllies healingSubject = iota
	RecEnemies
	RecAll
)

func (h healingSubject) String() string {
	switch h {
	case RecAllies:
		return "Allies"
	case RecEnemies:
		return "Enemies"
	case RecAll:
		return "All"
	}

	return ""
}

func NewHealingCollector() *HealingCollector {
	subjectDropdown, err := components.NewDropdown(
		"Subject",
		RecAllies,
		RecEnemies,
		RecAll,
	)
	if err != nil {
		log.Fatalln(err)
	}

	return &HealingCollector{
		subjectDropdown:    subjectDropdown,
		enemyTypesCheckbox: &widget.Bool{},
		longFormatBool:     &widget.Bool{},
	}
}

type HealingCollector struct {
	allies            healingWithMax
	enemies           healingWithMax
	enemyTypes        healingWithMax
	allWithEnemies    healingWithMax
	allWithEnemyTypes healingWithMax
	registeredPets    []string

	currentSubject     healingSubject
	subjectDropdown    *components.Dropdown
	enemyTypesCheckbox *widget.Bool
	longFormatBool     *widget.Bool
}

func (h *HealingCollector) Reset() {
	h.allies = healingWithMax{}
	h.enemies = healingWithMax{}
	h.enemyTypes = healingWithMax{}
	h.allWithEnemies = healingWithMax{}
	h.allWithEnemyTypes = healingWithMax{}
	h.registeredPets = nil
}
func (h *HealingCollector) lookForPet(skillUse *abstract.SkillUse) {
	if strings.Contains(skillUse.Skill, "(Pet)") {
		h.registeredPets = append(h.registeredPets, skillUse.Subject)
	}
}

func (h *HealingCollector) isAlly(info abstract.StatisticsInformation, subject string) bool {
	if subject == info.CurrentUsername() {
		return true
	}

	for _, pet := range h.registeredPets {
		if pet == subject {
			return true
		}
	}

	lowTrimmedSubject := strings.ToLower(strings.TrimSpace(SplitOffId(subject)))

	for _, expectedName := range info.Settings().EntitiesThatCountAsPets {
		if lowTrimmedSubject == strings.ToLower(strings.TrimSpace(expectedName)) {
			return true
		}
	}

	return false
}

func (h *HealingCollector) ingestRecovered(info abstract.StatisticsInformation, event *abstract.ChatEvent) {
	recovered := event.Contents.(*abstract.Recovered)

	findHeal := func(subject string) func(heal healing) bool {
		return func(heal healing) bool {
			return heal.name == subject
		}
	}
	createHeal := func(subject string) func() healing {
		return func() healing {
			return healing{
				name:      subject,
				amount:    1,
				recovered: recovered.Healed,
			}
		}
	}
	updateHeal := func(heal healing) healing {
		heal.amount++
		heal.recovered = heal.recovered.Add(recovered.Healed)

		return heal
	}
	healMax := func(a, b healing) int {
		return cmp.Compare(a.recovered.Total(), b.recovered.Total())
	}
	healSort := func(a, b healing) int {
		return cmp.Compare(b.recovered.Total(), a.recovered.Total())
	}
	processHealingWithMax := func(stat *healingWithMax, subject string) {
		stat.subjects = utils.CreateUpdate(
			stat.subjects,
			findHeal(subject),
			createHeal(subject),
			updateHeal,
		)
		slices.SortFunc(stat.subjects, healSort)
		stat.max = slices.MaxFunc(stat.subjects, healMax).recovered
	}

	if h.isAlly(info, recovered.Subject) {
		processHealingWithMax(&h.allies, recovered.Subject)
		processHealingWithMax(&h.allWithEnemies, recovered.Subject)
		processHealingWithMax(&h.allWithEnemyTypes, recovered.Subject)
	} else {
		processHealingWithMax(&h.enemies, recovered.Subject)
		processHealingWithMax(&h.allWithEnemies, recovered.Subject)

		group := SplitOffId(recovered.Subject)
		processHealingWithMax(&h.enemyTypes, group)
		processHealingWithMax(&h.allWithEnemyTypes, group)
	}
}

func (h *HealingCollector) Tick(info abstract.StatisticsInformation, at time.Time) {

}

func (h *HealingCollector) Collect(info abstract.StatisticsInformation, event *abstract.ChatEvent) error {
	if skillUse, ok := event.Contents.(*abstract.SkillUse); ok && skillUse.Damage != nil {
		h.lookForPet(skillUse)
	}

	if _, ok := event.Contents.(*abstract.Recovered); ok {
		h.ingestRecovered(info, event)
	}

	return nil
}

func (h *HealingCollector) TabName() string {
	return "Recovered"
}

func (h *HealingCollector) drawBar(state abstract.LayeredState, healed healing, maxDamage int, size unit.Dp) layout.Widget {
	return drawUniversalBar(
		state, healed.recovered,
		healed.recovered.Total(), maxDamage, healed.amount,
		healed.name, "recovered %v times",
		size, h.longFormatBool.Value,
	)
}

func (h *HealingCollector) UI(state abstract.LayeredState) (layout.Widget, []layout.Widget) {
	if h.subjectDropdown.Changed() {
		h.currentSubject = h.subjectDropdown.Value.(healingSubject)
	}

	topWidget := topBarSurface(func(gtx layout.Context) layout.Dimensions {
		if h.enemyTypesCheckbox.Update(gtx) {
			gtx.Source.Execute(op.InvalidateCmd{})
		}

		if h.longFormatBool.Update(gtx) {
			gtx.Source.Execute(op.InvalidateCmd{})
		}

		return layout.Flex{
			Axis: layout.Vertical,
		}.Layout(
			gtx,
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return layout.Flex{
					Axis:      layout.Horizontal,
					Alignment: layout.Middle,
				}.Layout(
					gtx,
					layout.Rigid(defaultDropdownStyle(state, h.subjectDropdown).Layout),
				)
			}),
			utils.FlexSpacerH(utils.CommonSpacing),
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return layout.Flex{
					Axis:      layout.Horizontal,
					Alignment: layout.Middle,
				}.Layout(
					gtx,
					layout.Rigid(defaultCheckboxStyle(state, h.enemyTypesCheckbox, "Group enemy types").Layout),
					utils.FlexSpacerW(utils.CommonSpacing),
					layout.Rigid(defaultCheckboxStyle(state, h.longFormatBool, "Use long numbers").Layout),
				)
			}),
		)
	})

	var widgets []layout.Widget

	var stats healingWithMax
	switch h.currentSubject {
	case RecAllies:
		stats = h.allies
	case RecEnemies:
		if h.enemyTypesCheckbox.Value {
			stats = h.enemyTypes
		} else {
			stats = h.enemies
		}
	case RecAll:
		if h.enemyTypesCheckbox.Value {
			stats = h.allWithEnemyTypes
		} else {
			stats = h.allWithEnemies
		}
	}

	maxHealed := stats.max.Total()

	for _, healed := range stats.subjects {
		widgets = append(widgets, h.drawBar(state, healed, maxHealed, 40))
	}

	return topWidget, widgets
}
