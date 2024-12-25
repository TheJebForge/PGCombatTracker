package collectors

import (
	"PGCombatTracker/abstract"
	"PGCombatTracker/ui/components"
	"PGCombatTracker/utils"
	"cmp"
	"fmt"
	"gioui.org/layout"
	"gioui.org/widget/material"
	"image"
	"log"
	"slices"
)

type enemyDamage struct {
	name   string
	amount int
	damage *abstract.Vitals
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

	return &DamageTakenCollector{
		groupByDropdown: groupByDropdown,
	}
}

type DamageTakenCollector struct {
	damageFromEnemies    []*enemyDamage
	damageFromEnemyTypes []*enemyDamage

	// UI stuff
	groupByDropdown *components.Dropdown
}

func (d *DamageTakenCollector) Reset() {
	d.damageFromEnemies = nil
}

func (d *DamageTakenCollector) ingestDamage(event *abstract.ChatEvent) {
	skillUse := event.Contents.(*abstract.SkillUse)

	// Individual enemies
	d.damageFromEnemies = utils.CreateUpdate(
		d.damageFromEnemies,
		func(enemy *enemyDamage) bool {
			return enemy.name == skillUse.Subject
		},
		func() *enemyDamage {
			return &enemyDamage{
				name:   skillUse.Subject,
				amount: 1,
				damage: skillUse.Damage,
			}
		},
		func(counter *enemyDamage) *enemyDamage {
			counter.amount++
			counter.damage = counter.damage.Add(skillUse.Damage)

			return counter
		},
	)
	slices.SortFunc(d.damageFromEnemies, func(a, b *enemyDamage) int {
		return cmp.Compare(b.damage.Total(), a.damage.Total())
	})

	// EnemyTypes
	d.damageFromEnemyTypes = utils.CreateUpdate(
		d.damageFromEnemyTypes,
		func(enemy *enemyDamage) bool {
			return enemy.name == SplitOffId(skillUse.Subject)
		},
		func() *enemyDamage {
			return &enemyDamage{
				name:   SplitOffId(skillUse.Subject),
				amount: 1,
				damage: skillUse.Damage,
			}
		},
		func(counter *enemyDamage) *enemyDamage {
			counter.amount++
			counter.damage = counter.damage.Add(skillUse.Damage)

			return counter
		},
	)
	slices.SortFunc(d.damageFromEnemyTypes, func(a, b *enemyDamage) int {
		return cmp.Compare(b.damage.Total(), a.damage.Total())
	})

}

func (d *DamageTakenCollector) Collect(info abstract.StatisticsInformation, event *abstract.ChatEvent) error {
	if skillUse, ok := event.Contents.(*abstract.SkillUse); ok && skillUse.Damage != nil && skillUse.Victim == info.CurrentUsername() {
		d.ingestDamage(event)
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

func (d *DamageTakenCollector) topBar(state abstract.LayeredState) layout.Widget {
	return topBarSurface(func(gtx layout.Context) layout.Dimensions {
		return layout.Flex{
			Axis: layout.Horizontal,
		}.Layout(
			gtx,
			layout.Rigid(defaultDropdownStyle(state, d.groupByDropdown).Layout),
		)
	})
}

func (d *DamageTakenCollector) UI(state abstract.LayeredState) []layout.Widget {
	widgets := []layout.Widget{
		d.topBar(state),
	}

	// All the bars go here
	var damageArray []*enemyDamage
	switch d.groupByDropdown.Value.(GroupBy) {
	case DontGroup:
		damageArray = d.damageFromEnemies
	case GroupByType:
		damageArray = d.damageFromEnemyTypes
	}

	var maxDamage int
	for _, enemy := range damageArray {
		if totalDamage := enemy.damage.Total(); totalDamage > maxDamage {
			maxDamage = totalDamage
		}
	}

	for _, enemy := range damageArray {
		widgets = append(widgets,
			func(gtx layout.Context) layout.Dimensions {
				var progress = float64(enemy.damage.Total()) / float64(maxDamage)

				return components.Canvas{
					ExpandHorizontal: true,
					MinSize: image.Point{
						Y: gtx.Dp(40 + utils.CommonSpacing),
					},
				}.Layout(
					gtx,
					components.CanvasItem{
						Anchor: layout.N,
						Widget: components.BarWidget(components.StringToColor(enemy.name), 40, progress),
					},
					components.CanvasItem{
						Offset: image.Point{
							X: gtx.Dp(5),
							Y: gtx.Dp(5),
						},
						Widget: material.Label(state.Theme(), 12, enemy.name).Layout,
					},
					components.CanvasItem{
						Anchor: layout.SW,
						Offset: image.Point{
							X: gtx.Dp(5),
							Y: gtx.Dp(-5 - utils.CommonSpacing),
						},
						Widget: material.Label(state.Theme(), 12, fmt.Sprintf("attacked %v times",
							enemy.amount)).Layout,
					},
					components.CanvasItem{
						Anchor: layout.NE,
						Offset: image.Point{
							X: gtx.Dp(-5),
							Y: gtx.Dp(12.5),
						},
						Widget: material.Label(state.Theme(), 12, enemy.damage.String()).Layout,
					},
				)
			},
		)
	}

	return widgets
}
