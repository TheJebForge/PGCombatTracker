package collectors

import (
	"PGCombatTracker/abstract"
	"PGCombatTracker/ui"
	"cmp"
	"fmt"
	"gioui.org/layout"
	"gioui.org/widget/material"
	"image"
	"slices"
)

type enemyDamage struct {
	name   string
	amount int
	damage *abstract.Vitals
}

func NewDamageTakenCollector() *DamageTakenCollector {
	return &DamageTakenCollector{}
}

type DamageTakenCollector struct {
	damageFromEnemies []*enemyDamage
}

func (d *DamageTakenCollector) Reset() {
	d.damageFromEnemies = nil
}

func (d *DamageTakenCollector) ingestDamage(event *abstract.ChatEvent) {
	var counter *enemyDamage
	skillUse := event.Contents.(*abstract.SkillUse)

	found := false
	for _, c := range d.damageFromEnemies {
		if c.name == skillUse.Skill {
			counter = c
			found = true
			break
		}
	}

	if !found {
		counter = &enemyDamage{
			name:   skillUse.Subject,
			amount: 1,
			damage: skillUse.Damage,
		}

		d.damageFromEnemies = append(d.damageFromEnemies, counter)
	} else {
		counter.amount++
		counter.damage = counter.damage.Add(skillUse.Damage)
	}

	slices.SortFunc(d.damageFromEnemies, func(a, b *enemyDamage) int {
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

func (d *DamageTakenCollector) UI(state abstract.GlobalState) []layout.Widget {
	var widgets []layout.Widget

	var maxDamage int
	for _, enemy := range d.damageFromEnemies {
		if totalDamage := enemy.damage.Total(); totalDamage > maxDamage {
			maxDamage = totalDamage
		}
	}

	for _, enemy := range d.damageFromEnemies {
		widgets = append(widgets,
			func(gtx layout.Context) layout.Dimensions {
				var progress = float64(enemy.damage.Total()) / float64(maxDamage)

				return ui.Canvas{
					ExpandHorizontal: true,
					MinSize: image.Point{
						Y: gtx.Dp(40 + abstract.CommonSpacing),
					},
				}.Layout(
					gtx,
					ui.CanvasItem{
						Anchor: layout.N,
						Widget: ui.BarWidget(ui.StringToColor(enemy.name), 40, progress),
					},
					ui.CanvasItem{
						Offset: image.Point{
							X: gtx.Dp(5),
							Y: gtx.Dp(5),
						},
						Widget: material.Label(state.Theme(), 12, enemy.name).Layout,
					},
					ui.CanvasItem{
						Anchor: layout.SW,
						Offset: image.Point{
							X: gtx.Dp(5),
							Y: gtx.Dp(-5 - abstract.CommonSpacing),
						},
						Widget: material.Label(state.Theme(), 12, fmt.Sprintf("attacked %v times",
							enemy.amount)).Layout,
					},
					ui.CanvasItem{
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
