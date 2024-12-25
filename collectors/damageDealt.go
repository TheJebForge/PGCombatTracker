package collectors

import (
	"PGCombatTracker/abstract"
	"PGCombatTracker/ui/components"
	"PGCombatTracker/utils"
	"cmp"
	"fmt"
	"gioui.org/layout"
	"gioui.org/unit"
	"gioui.org/widget/material"
	"image"
	"slices"
	"time"
)

type skillDamage struct {
	name     string
	amount   int
	damage   abstract.Vitals
	lastUsed time.Time
}

func NewDamageDealtCollector() *DamageDealtCollector {
	return &DamageDealtCollector{}
}

type DamageDealtCollector struct {
	totalDamage abstract.Vitals
	skillDamage []*skillDamage
}

func (col *DamageDealtCollector) ingestDamage(event *abstract.ChatEvent) {
	skillUse := event.Contents.(*abstract.SkillUse)

	// Skill damage
	col.skillDamage = utils.CreateUpdate(
		col.skillDamage,
		func(counter *skillDamage) bool {
			return counter.name == skillUse.Skill
		},
		func() *skillDamage {
			return &skillDamage{
				name:     skillUse.Skill,
				amount:   1,
				damage:   *skillUse.Damage,
				lastUsed: event.Time,
			}
		},
		func(counter *skillDamage) *skillDamage {
			if counter.lastUsed != event.Time {
				counter.amount++
			}
			counter.damage = counter.damage.Add(*skillUse.Damage)
			counter.lastUsed = event.Time

			return counter
		},
	)
	slices.SortFunc(col.skillDamage, func(a, b *skillDamage) int {
		return cmp.Compare(b.damage.Total(), a.damage.Total())
	})

	// Collect total damage
	col.totalDamage = col.totalDamage.Add(*skillUse.Damage)
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

func (col *DamageDealtCollector) drawBar(state abstract.LayeredState, skill *skillDamage, maxDamage int, size unit.Dp) layout.Widget {
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
				Widget: material.Label(state.Theme(), 12, skill.damage.String()).Layout,
			},
		)
	}
}

func (col *DamageDealtCollector) UI(state abstract.LayeredState) []layout.Widget {
	var widgets []layout.Widget

	var maxDamage int
	for _, skill := range col.skillDamage {
		if totalDamage := skill.damage.Total(); totalDamage > maxDamage {
			maxDamage = totalDamage
		}
	}

	widgets = append(widgets, col.drawBar(
		state,
		&skillDamage{
			name:   "Total Damage",
			amount: 0,
			damage: col.totalDamage,
		},
		col.totalDamage.Total(),
		25,
	))

	for _, skill := range col.skillDamage {
		widgets = append(widgets, col.drawBar(state, skill, maxDamage, 40))
	}

	return widgets
}
