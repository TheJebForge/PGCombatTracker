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
	"time"
)

type skillDamage struct {
	name     string
	amount   int
	damage   *abstract.Vitals
	lastUsed time.Time
}

func NewDamageDealtCollector() *DamageDealtCollector {
	return &DamageDealtCollector{}
}

type DamageDealtCollector struct {
	skillDamage []*skillDamage
}

func (col *DamageDealtCollector) ingestDamage(event *abstract.ChatEvent) {
	var counter *skillDamage
	skillUse := event.Contents.(*abstract.SkillUse)

	found := false
	for _, c := range col.skillDamage {
		if c.name == skillUse.Skill {
			counter = c
			found = true
			break
		}
	}

	if !found {
		counter = &skillDamage{
			name:     skillUse.Skill,
			amount:   1,
			damage:   skillUse.Damage,
			lastUsed: event.Time,
		}

		col.skillDamage = append(col.skillDamage, counter)
	} else {
		if counter.lastUsed != event.Time {
			counter.amount++
		}
		counter.damage = counter.damage.Add(skillUse.Damage)
		counter.lastUsed = event.Time
	}

	slices.SortFunc(col.skillDamage, func(a, b *skillDamage) int {
		return cmp.Compare(b.damage.Total(), a.damage.Total())
	})
}

func (col *DamageDealtCollector) Reset() {
	col.skillDamage = nil
}

func (col *DamageDealtCollector) Collect(info abstract.StatisticsInformation, event *abstract.ChatEvent) error {
	if skillUse, ok := event.Contents.(*abstract.SkillUse); ok && skillUse.Damage != nil && skillUse.Subject == info.CurrentUsername() {
		col.ingestDamage(event)
	}

	return nil
}

func (col *DamageDealtCollector) TabName() string {
	return "Damage Dealt"
}

func (col *DamageDealtCollector) UI(state abstract.GlobalState) []layout.Widget {
	var widgets []layout.Widget

	var maxDamage int
	for _, skill := range col.skillDamage {
		if totalDamage := skill.damage.Total(); totalDamage > maxDamage {
			maxDamage = totalDamage
		}
	}

	for _, skill := range col.skillDamage {
		widgets = append(widgets,
			func(gtx layout.Context) layout.Dimensions {
				var progress = float64(skill.damage.Total()) / float64(maxDamage)

				return ui.Canvas{
					ExpandHorizontal: true,
					MinSize: image.Point{
						Y: gtx.Dp(40 + abstract.CommonSpacing),
					},
				}.Layout(
					gtx,
					ui.CanvasItem{
						Anchor: layout.N,
						Widget: ui.BarWidget(ui.StringToColor(skill.name), 40, progress),
					},
					ui.CanvasItem{
						Offset: image.Point{
							X: gtx.Dp(5),
							Y: gtx.Dp(5),
						},
						Widget: material.Label(state.Theme(), 12, skill.name).Layout,
					},
					ui.CanvasItem{
						Anchor: layout.SW,
						Offset: image.Point{
							X: gtx.Dp(5),
							Y: gtx.Dp(-5 - abstract.CommonSpacing),
						},
						Widget: material.Label(state.Theme(), 12, fmt.Sprintf("used %v times",
							skill.amount)).Layout,
					},
					ui.CanvasItem{
						Anchor: layout.NE,
						Offset: image.Point{
							X: gtx.Dp(-5),
							Y: gtx.Dp(12.5),
						},
						Widget: material.Label(state.Theme(), 12, skill.damage.String()).Layout,
					},
				)
			},
		)
	}

	return widgets
}
