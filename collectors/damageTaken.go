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
)

type enemyDamage struct {
	name   string
	amount int
	damage abstract.Vitals
}

type enemyDamageWithMax struct {
	enemies   []enemyDamage
	maxDamage abstract.Vitals
}

type subjectiveDamageTaken struct {
	victim               string
	totalDamage          abstract.Vitals
	indirectDamage       abstract.Vitals
	damageFromEnemies    enemyDamageWithMax
	damageFromEnemyTypes enemyDamageWithMax
}

func NewDamageTakenCollector() *DamageTakenCollector {
	groupByDropdown, err := components.NewDropdown(
		"Group By",
		DontGroup,
		GroupByType,
	)
	if err != nil {
		log.Panicln(err)
	}

	victimDropdown, err := components.NewDropdown("Victim", subjectChoice(""))
	if err != nil {
		log.Fatalln(err)
	}

	return &DamageTakenCollector{
		groupByDropdown: groupByDropdown,
		victimDropdown:  victimDropdown,
		longFormatBool:  &widget.Bool{},
	}
}

type DamageTakenCollector struct {
	currentVictim  string
	total          subjectiveDamageTaken
	victims        []subjectiveDamageTaken
	registeredPets []string

	// UI stuff
	victimDropdown  *components.Dropdown
	groupByDropdown *components.Dropdown
	longFormatBool  *widget.Bool
}

func (d *DamageTakenCollector) Reset() {
	d.victims = nil
	d.total = subjectiveDamageTaken{}
	d.registeredPets = nil
	d.victimDropdown.SetOptions([]fmt.Stringer{subjectChoice("")})
	d.currentVictim = ""
}

func (d *DamageTakenCollector) ingestDamage(event *abstract.ChatEvent) {
	skillUse := event.Contents.(*abstract.SkillUse)

	findEnemyDamage := func(grouped bool) func(enemy enemyDamage) bool {
		return func(enemy enemyDamage) bool {
			if grouped {
				return enemy.name == SplitOffId(skillUse.Subject)
			} else {
				return enemy.name == skillUse.Subject
			}
		}
	}
	createEnemyDamage := func(grouped bool) func() enemyDamage {
		return func() enemyDamage {
			subject := skillUse.Subject
			if grouped {
				subject = SplitOffId(subject)
			}

			return enemyDamage{
				name:   subject,
				amount: 1,
				damage: *skillUse.Damage,
			}
		}
	}
	updateEnemyDamage := func(enemy enemyDamage) enemyDamage {
		enemy.amount++
		enemy.damage = enemy.damage.Add(*skillUse.Damage)

		return enemy
	}
	enemyDamageMax := func(a, b enemyDamage) int {
		return cmp.Compare(a.damage.Total(), b.damage.Total())
	}
	enemyDamageSort := func(a, b enemyDamage) int {
		return cmp.Compare(b.damage.Total(), a.damage.Total())
	}

	// Ingest total stuff
	d.total.totalDamage = d.total.totalDamage.Add(*skillUse.Damage)
	d.total.damageFromEnemies.enemies = utils.CreateUpdate(
		d.total.damageFromEnemies.enemies,
		findEnemyDamage(false),
		createEnemyDamage(false),
		updateEnemyDamage,
	)
	slices.SortFunc(d.total.damageFromEnemies.enemies, enemyDamageSort)
	d.total.damageFromEnemies.maxDamage = slices.MaxFunc(d.total.damageFromEnemies.enemies, enemyDamageMax).damage
	d.total.damageFromEnemyTypes.enemies = utils.CreateUpdate(
		d.total.damageFromEnemyTypes.enemies,
		findEnemyDamage(true),
		createEnemyDamage(true),
		updateEnemyDamage,
	)
	slices.SortFunc(d.total.damageFromEnemyTypes.enemies, enemyDamageSort)
	d.total.damageFromEnemyTypes.maxDamage = slices.MaxFunc(d.total.damageFromEnemyTypes.enemies, enemyDamageMax).damage

	// Ingest individual stuff
	d.victims = utils.CreateUpdate(
		d.victims,
		func(victim subjectiveDamageTaken) bool {
			return victim.victim == skillUse.Victim
		},
		func() subjectiveDamageTaken {
			return subjectiveDamageTaken{
				victim:      skillUse.Victim,
				totalDamage: *skillUse.Damage,
				damageFromEnemies: enemyDamageWithMax{
					enemies: []enemyDamage{
						createEnemyDamage(false)(),
					},
					maxDamage: *skillUse.Damage,
				},
				damageFromEnemyTypes: enemyDamageWithMax{
					enemies: []enemyDamage{
						createEnemyDamage(true)(),
					},
					maxDamage: *skillUse.Damage,
				},
			}
		},
		func(victim subjectiveDamageTaken) subjectiveDamageTaken {
			victim.totalDamage = victim.totalDamage.Add(*skillUse.Damage)
			victim.damageFromEnemies.enemies = utils.CreateUpdate(
				victim.damageFromEnemies.enemies,
				findEnemyDamage(false),
				createEnemyDamage(false),
				updateEnemyDamage,
			)
			slices.SortFunc(victim.damageFromEnemies.enemies, enemyDamageSort)
			victim.damageFromEnemies.maxDamage = slices.MaxFunc(victim.damageFromEnemies.enemies, enemyDamageMax).damage
			victim.damageFromEnemyTypes.enemies = utils.CreateUpdate(
				victim.damageFromEnemyTypes.enemies,
				findEnemyDamage(true),
				createEnemyDamage(true),
				updateEnemyDamage,
			)
			slices.SortFunc(victim.damageFromEnemyTypes.enemies, enemyDamageSort)
			victim.damageFromEnemyTypes.maxDamage = slices.MaxFunc(victim.damageFromEnemyTypes.enemies, enemyDamageMax).damage

			return victim
		},
	)

	d.victimDropdown.SetOptions(utils.CreateUpdate(
		d.victimDropdown.Options(),
		func(item fmt.Stringer) bool {
			casted := item.(subjectChoice)
			return string(casted) == skillUse.Victim
		},
		func() fmt.Stringer {
			return subjectChoice(skillUse.Victim)
		},
		func(stringer fmt.Stringer) fmt.Stringer {
			return stringer
		},
	))
}

func (d *DamageTakenCollector) ingestIndirectDamage(event *abstract.ChatEvent) {
	indirect := event.Contents.(*abstract.IndirectDamage)
	indirectDamage := indirect.Damage.Abs()
	d.total.totalDamage = d.total.totalDamage.Add(indirectDamage)
	d.total.indirectDamage = d.total.indirectDamage.Add(indirectDamage)

	d.victims = utils.CreateUpdate(
		d.victims,
		func(victim subjectiveDamageTaken) bool {
			return victim.victim == indirect.Subject
		},
		func() subjectiveDamageTaken {
			return subjectiveDamageTaken{
				victim:         indirect.Subject,
				totalDamage:    indirect.Damage,
				indirectDamage: indirect.Damage,
			}
		},
		func(victim subjectiveDamageTaken) subjectiveDamageTaken {
			victim.totalDamage = victim.totalDamage.Add(indirectDamage)
			victim.indirectDamage = victim.indirectDamage.Add(indirectDamage)

			return victim
		},
	)
}

func (d *DamageTakenCollector) lookForPet(skillUse *abstract.SkillUse) {
	if strings.Contains(skillUse.Skill, "(Pet)") {
		d.registeredPets = append(d.registeredPets, skillUse.Subject)
	}
}

func (d *DamageTakenCollector) isVictimValuable(info abstract.StatisticsInformation, victim string) bool {
	if victim == info.CurrentUsername() {
		return true
	}

	for _, pet := range d.registeredPets {
		if pet == victim {
			return true
		}
	}

	lowTrimmedSubject := strings.ToLower(strings.TrimSpace(SplitOffId(victim)))

	for _, expectedName := range info.Settings().EntitiesThatCountAsPets {
		if lowTrimmedSubject == strings.ToLower(strings.TrimSpace(expectedName)) {
			return true
		}
	}

	return false
}

func (d *DamageTakenCollector) Collect(info abstract.StatisticsInformation, event *abstract.ChatEvent) error {
	if skillUse, ok := event.Contents.(*abstract.SkillUse); ok && skillUse.Damage != nil {
		d.lookForPet(skillUse)

		if d.isVictimValuable(info, skillUse.Victim) {
			d.ingestDamage(event)
		}
	}

	if indirect, ok := event.Contents.(*abstract.IndirectDamage); ok && d.isVictimValuable(info, indirect.Subject) {
		d.ingestIndirectDamage(event)
	}

	return nil
}

func (d *DamageTakenCollector) TabName() string {
	return "Damage Taken"
}

type GroupBy int

const (
	DontGroup GroupBy = iota
	GroupByType
)

func (g GroupBy) String() string {
	switch g {
	case DontGroup:
		return "Don't Group"
	case GroupByType:
		return "Group By Enemy Type"
	}

	return "Unknown"
}

func (d *DamageTakenCollector) drawBar(state abstract.LayeredState, enemy enemyDamage, maxDamage int, size unit.Dp) layout.Widget {
	return drawUniversalBar(
		state, enemy.damage,
		enemy.damage.Total(), maxDamage, enemy.amount,
		enemy.name, "attacked %v times",
		size, d.longFormatBool.Value,
	)
}

func (d *DamageTakenCollector) UI(state abstract.LayeredState) (layout.Widget, []layout.Widget) {
	if d.victimDropdown.Changed() {
		d.currentVictim = string(d.victimDropdown.Value.(subjectChoice))
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
				return layout.Flex{
					Axis:      layout.Horizontal,
					Alignment: layout.Middle,
				}.Layout(
					gtx,
					layout.Rigid(defaultDropdownStyle(state, d.victimDropdown).Layout),
					utils.FlexSpacerW(utils.CommonSpacing),
					layout.Rigid(defaultCheckboxStyle(state, d.longFormatBool, "Use long numbers").Layout),
				)
			}),
			utils.FlexSpacerH(utils.CommonSpacing),
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return layout.Flex{
					Axis:      layout.Horizontal,
					Alignment: layout.Middle,
				}.Layout(
					gtx,
					layout.Rigid(defaultDropdownStyle(state, d.groupByDropdown).Layout),
				)
			}),
		)
	})

	var widgets []layout.Widget

	victim := d.total
	for _, possibleVictim := range d.victims {
		if possibleVictim.victim == d.currentVictim {
			victim = possibleVictim
			break
		}
	}

	widgets = append(widgets, d.drawBar(
		state,
		enemyDamage{
			name:   "Total Damage",
			amount: 0,
			damage: victim.totalDamage,
		},
		victim.totalDamage.Total(),
		25,
	))

	// All the bars go here
	var enemies *enemyDamageWithMax
	switch d.groupByDropdown.Value.(GroupBy) {
	case DontGroup:
		enemies = &victim.damageFromEnemies
	case GroupByType:
		enemies = &victim.damageFromEnemyTypes
	default:
		log.Fatalln("wtf happened to the dropdown")
	}

	maxDamage := enemies.maxDamage.Total()

	for _, enemy := range enemies.enemies {
		widgets = append(widgets, d.drawBar(state, enemy, maxDamage, 40))
	}

	widgets = append(widgets, d.drawBar(
		state,
		enemyDamage{
			name:   "Indirect Damage",
			amount: 0,
			damage: victim.indirectDamage,
		},
		victim.indirectDamage.Total(),
		25,
	))

	return topWidget, widgets
}
