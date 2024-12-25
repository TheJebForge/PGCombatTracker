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
	skillUse := event.Contents.(*abstract.SkillUse)

	col.skillDamage = utils.CreateUpdate(
		col.skillDamage,
		func(counter *skillDamage) bool {
			return counter.name == skillUse.Skill
		},
		func() *skillDamage {
			return &skillDamage{
				name:     skillUse.Skill,
				amount:   1,
				damage:   skillUse.Damage,
				lastUsed: event.Time,
			}
		},
		func(counter *skillDamage) *skillDamage {
			if counter.lastUsed != event.Time {
				counter.amount++
			}
			counter.damage = counter.damage.Add(skillUse.Damage)
			counter.lastUsed = event.Time

			return counter
		},
	)

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

func (col *DamageDealtCollector) UI(state abstract.LayeredState) []layout.Widget {
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

				return components.Canvas{
					ExpandHorizontal: true,
					MinSize: image.Point{
						Y: gtx.Dp(40 + utils.CommonSpacing),
					},
				}.Layout(
					gtx,
					components.CanvasItem{
						Anchor: layout.N,
						Widget: components.BarWidget(components.StringToColor(skill.name), 40, progress),
					},
					components.CanvasItem{
						Offset: image.Point{
							X: gtx.Dp(5),
							Y: gtx.Dp(5),
						},
						Widget: material.Label(state.Theme(), 12, skill.name).Layout,
					},
					components.CanvasItem{
						Anchor: layout.SW,
						Offset: image.Point{
							X: gtx.Dp(5),
							Y: gtx.Dp(-5 - utils.CommonSpacing),
						},
						Widget: material.Label(state.Theme(), 12, fmt.Sprintf("used %v times",
							skill.amount)).Layout,
					},
					components.CanvasItem{
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
