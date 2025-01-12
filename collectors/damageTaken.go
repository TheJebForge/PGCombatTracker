package collectors

import (
	"PGCombatTracker/abstract"
	"PGCombatTracker/ui/components"
	"PGCombatTracker/utils"
	"PGCombatTracker/utils/drawing"
	"cmp"
	"fmt"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/unit"
	"gioui.org/widget"
	"image"
	"log"
	"slices"
	"strings"
	"time"
)

type enemyDamage struct {
	name   string
	amount int
	damage abstract.Vitals
	chart  *components.TimeBasedChart
}

type enemyDamageWithMax struct {
	enemies   []enemyDamage
	maxDamage abstract.Vitals
	maxRange  components.DataRange
}

type subjectiveDamageTaken struct {
	victim               string
	timeController       *components.TimeController
	totalChart           *components.TimeBasedChart
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

	displayDropdown, err := components.NewDropdown(
		"Display",
		DisplayBars,
		DisplayPie,
		DisplayGraphs,
	)
	if err != nil {
		log.Fatalln(err)
	}

	limitDropdown, err := components.NewDropdown(
		"Limit",
		LimitTop5,
		LimitTop10,
		LimitTop15,
		LimitTop25,
		LimitTop50,
		NoLimit,
	)

	return &DamageTakenCollector{
		groupByDropdown: groupByDropdown,
		victimDropdown:  victimDropdown,
		longFormatBool:  &widget.Bool{},
		displayDropdown: displayDropdown,
		limitDropdown:   limitDropdown,
		total: subjectiveDamageTaken{
			timeController: components.NewTimeControllerOrCrash(components.NewTimeBasedChart("Total")),
			totalChart:     components.NewTimeBasedChart("Total"),
		},
	}
}

type DamageTakenCollector struct {
	currentVictim  string
	total          subjectiveDamageTaken
	victims        []subjectiveDamageTaken
	registeredPets []string

	currentDisplay displayChoice
	currentLimit   limitChoice

	// UI stuff
	victimDropdown  *components.Dropdown
	groupByDropdown *components.Dropdown
	longFormatBool  *widget.Bool
	displayDropdown *components.Dropdown
	limitDropdown   *components.Dropdown
}

func (d *DamageTakenCollector) Reset(info abstract.StatisticsInformation) {
	d.victims = nil
	d.total = subjectiveDamageTaken{
		timeController: components.NewTimeControllerOrCrash(components.NewTimeBasedChart("Total")),
		totalChart:     components.NewTimeBasedChart("Total"),
	}
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

			chart := components.NewTimeBasedChart(subject)
			chart.Add(components.TimePoint{
				Time:    event.Time,
				Value:   skillUse.Damage.Total(),
				Details: *skillUse.Damage,
			})

			return enemyDamage{
				name:   subject,
				amount: 1,
				damage: *skillUse.Damage,
				chart:  chart,
			}
		}
	}
	updateEnemyDamage := func(enemy enemyDamage) enemyDamage {
		enemy.amount++
		enemy.damage = enemy.damage.Add(*skillUse.Damage)
		enemy.chart.Add(components.TimePoint{
			Time:    event.Time,
			Value:   enemy.damage.Total(),
			Details: enemy.damage,
		})

		return enemy
	}
	enemyDamageMax := func(a, b enemyDamage) int {
		return cmp.Compare(a.damage.Total(), b.damage.Total())
	}
	enemyDamageSort := func(a, b enemyDamage) int {
		return cmp.Compare(b.damage.Total(), a.damage.Total())
	}

	processEnemyDamageWithMax := func(enemies *enemyDamageWithMax, grouped bool) {
		enemies.enemies = utils.CreateUpdate(
			enemies.enemies,
			findEnemyDamage(grouped),
			createEnemyDamage(grouped),
			updateEnemyDamage,
		)
		slices.SortFunc(enemies.enemies, enemyDamageSort)
		enemies.maxDamage = slices.MaxFunc(enemies.enemies, enemyDamageMax).damage
		enemies.maxRange = enemies.maxRange.Expand(enemies.maxDamage.Total())
	}

	processSubjectiveDT := func(subject *subjectiveDamageTaken) {
		subject.totalDamage = subject.totalDamage.Add(*skillUse.Damage)
		point := components.TimePoint{
			Time:    event.Time,
			Value:   subject.totalDamage.Total(),
			Details: subject.totalDamage,
		}
		subject.timeController.Add(point)
		subject.totalChart.Add(point)
		processEnemyDamageWithMax(&subject.damageFromEnemies, false)
		processEnemyDamageWithMax(&subject.damageFromEnemyTypes, true)
	}

	// Ingest total stuff
	processSubjectiveDT(&d.total)

	// Ingest individual stuff
	d.victims = utils.CreateUpdate(
		d.victims,
		func(victim subjectiveDamageTaken) bool {
			return victim.victim == skillUse.Victim
		},
		func() subjectiveDamageTaken {
			totalDamage := *skillUse.Damage

			timeController := components.NewTimeControllerOrCrash(components.NewTimeBasedChart("Total"))
			totalChart := components.NewTimeBasedChart("Total")
			point := components.TimePoint{
				Time:    event.Time,
				Value:   totalDamage.Total(),
				Details: totalDamage,
			}
			timeController.Add(point)
			totalChart.Add(point)

			return subjectiveDamageTaken{
				victim:      skillUse.Victim,
				totalDamage: totalDamage,
				damageFromEnemies: enemyDamageWithMax{
					enemies: []enemyDamage{
						createEnemyDamage(false)(),
					},
					maxDamage: *skillUse.Damage,
					maxRange:  components.DataRange{Max: skillUse.Damage.Total()},
				},
				damageFromEnemyTypes: enemyDamageWithMax{
					enemies: []enemyDamage{
						createEnemyDamage(true)(),
					},
					maxDamage: *skillUse.Damage,
					maxRange:  components.DataRange{Max: skillUse.Damage.Total()},
				},
				timeController: timeController,
				totalChart:     totalChart,
			}
		},
		func(victim subjectiveDamageTaken) subjectiveDamageTaken {
			processSubjectiveDT(&victim)
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
			totalDamage := indirect.Damage

			timeController := components.NewTimeControllerOrCrash(components.NewTimeBasedChart("Total"))
			totalChart := components.NewTimeBasedChart("Total")
			point := components.TimePoint{
				Time:    event.Time,
				Value:   totalDamage.Total(),
				Details: totalDamage,
			}
			timeController.Add(point)
			totalChart.Add(point)

			return subjectiveDamageTaken{
				victim:         indirect.Subject,
				totalDamage:    totalDamage,
				indirectDamage: totalDamage,
				timeController: timeController,
				totalChart:     totalChart,
			}
		},
		func(victim subjectiveDamageTaken) subjectiveDamageTaken {
			victim.totalDamage = victim.totalDamage.Add(indirectDamage)
			victim.indirectDamage = victim.indirectDamage.Add(indirectDamage)

			point := components.TimePoint{
				Time:    event.Time,
				Value:   victim.totalDamage.Total(),
				Details: victim.totalDamage,
			}

			victim.timeController.Add(point)
			victim.totalChart.Add(point)

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

func (d *DamageTakenCollector) Tick(info abstract.StatisticsInformation, at time.Time) {

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

func (d *DamageTakenCollector) drawWidget(state abstract.LayeredState, enemy enemyDamage, widget layout.Widget, size unit.Dp) layout.Widget {
	return drawUniversalStatsText(
		state, enemy.damage,
		widget, enemy.amount,
		enemy.name, "attacked %v times",
		size, d.longFormatBool.Value,
	)
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

	if d.displayDropdown.Changed() {
		d.currentDisplay = d.displayDropdown.Value.(displayChoice)
	}

	if d.limitDropdown.Changed() {
		d.currentLimit = d.limitDropdown.Value.(limitChoice)
	}

	victim := d.total
	for _, possibleVictim := range d.victims {
		if possibleVictim.victim == d.currentVictim {
			victim = possibleVictim
			break
		}
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
				return components.HorizontalWrap{
					Alignment:   layout.Middle,
					Spacing:     utils.CommonSpacing,
					LineSpacing: utils.CommonSpacing,
				}.Layout(
					gtx,
					defaultDropdownStyle(state, d.victimDropdown).Layout,
					defaultCheckboxStyle(state, d.longFormatBool, "Use long numbers").Layout,
					defaultDropdownStyle(state, d.displayDropdown).Layout,
					func(gtx layout.Context) layout.Dimensions {
						switch d.currentDisplay {
						case DisplayGraphs:
							fallthrough
						case DisplayBars:
							return defaultDropdownStyle(state, d.groupByDropdown).Layout(gtx)
						case DisplayPie:
							return defaultDropdownStyle(state, d.limitDropdown).Layout(gtx)
						}

						return layout.Dimensions{}
					},
				)
			}),
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				if d.currentDisplay == DisplayGraphs {
					return layout.Flex{
						Axis: layout.Vertical,
					}.Layout(
						gtx,
						utils.FlexSpacerH(utils.CommonSpacing),
						layout.Rigid(components.StyleTimeController(state.Theme(), victim.timeController).Layout),
					)
				}

				return layout.Dimensions{}
			}),
		)
	})

	var widgets []layout.Widget

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

	switch d.currentDisplay {
	case DisplayBars:
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
	case DisplayPie:
		widgets = append(widgets, func(gtx layout.Context) layout.Dimensions {
			var totalValue int
			pieItems := make([]components.PieChartItem, 0, max(1, len(victim.damageFromEnemyTypes.enemies)))
			for i, enemy := range victim.damageFromEnemyTypes.enemies {
				if i >= d.currentLimit.Int() {
					break
				}

				pieItems = append(pieItems, components.PieChartItem{
					Name:    enemy.name,
					Value:   enemy.damage.Total(),
					SubText: enemy.damage.StringCL(d.longFormatBool.Value),
				})
				totalValue += enemy.damage.Total()
			}

			style := components.StylePieChart(state.Theme())
			style.TextSize = 12

			return style.Layout(
				gtx,
				totalValue,
				pieItems...,
			)
		})
	case DisplayGraphs:
		controller := victim.timeController

		totalChart := victim.totalChart

		totalChart.DisplayTimeFrame = controller.CurrentTimeFrame
		totalChart.DisplayValueRange = controller.FullValueRange

		totalChartStyle := components.StyleTimeBasedChart(state.Theme(), totalChart)
		totalChartStyle.Color = components.StringToColor("Total Damage")
		totalChartStyle.LongFormat = d.longFormatBool.Value

		widgets = append(widgets, d.drawWidget(state, enemyDamage{
			name:   "Total Damage",
			amount: 0,
			damage: victim.totalDamage,
		}, totalChartStyle.Layout, 100))

		for _, enemy := range enemies.enemies {
			enemy.chart.DisplayTimeFrame = controller.CurrentTimeFrame
			enemy.chart.DisplayValueRange = enemies.maxRange

			chartStyle := components.StyleTimeBasedChart(state.Theme(), enemy.chart)
			chartStyle.Color = components.StringToColor(enemy.name)
			chartStyle.LongFormat = d.longFormatBool.Value

			widgets = append(widgets, d.drawWidget(state, enemy, chartStyle.Layout, 100))
		}
	}

	return topWidget, widgets
}

func (d *DamageTakenCollector) exportWidget(styledFonts *drawing.StyledFontPack, enemy enemyDamage, widget drawing.Widget) drawing.Widget {
	return exportUniversalStatsTextAsStack(
		styledFonts, enemy.damage,
		widget, enemy.amount,
		enemy.name, "attacked %v times",
		d.longFormatBool.Value,
	)
}

func (d *DamageTakenCollector) exportBar(styledFonts *drawing.StyledFontPack, enemy enemyDamage, maxDamage int) drawing.Widget {
	return exportUniversalBar(
		styledFonts, enemy.damage,
		enemy.damage.Total(), maxDamage, enemy.amount,
		enemy.name, "attacked %v times",
		d.longFormatBool.Value,
	)
}

func (d *DamageTakenCollector) Export(state abstract.LayeredState) image.Image {
	victim := d.total
	for _, possibleVictim := range d.victims {
		if possibleVictim.victim == d.currentVictim {
			victim = possibleVictim
			break
		}
	}

	var enemies *enemyDamageWithMax
	switch d.groupByDropdown.Value.(GroupBy) {
	case DontGroup:
		enemies = &victim.damageFromEnemies
	case GroupByType:
		enemies = &victim.damageFromEnemyTypes
	default:
		log.Fatalln("wtf happened to the dropdown")
	}

	styledFonts := drawing.StyleFontPack(state.FontPack(), state.Theme().Fg)

	body := drawing.Empty

	switch d.currentDisplay {
	case DisplayBars:
		items := make([]drawing.FlexChild, 0, len(enemies.enemies)*2-1+4)

		maxDamage := enemies.maxDamage.Total()

		items = append(
			items,
			drawing.Rigid(d.exportBar(
				styledFonts,
				enemyDamage{
					name:   "Total Damage",
					amount: 0,
					damage: victim.totalDamage,
				},
				victim.totalDamage.Total(),
			)),
			drawing.FlexVSpacer(drawing.CommonSpacing),
		)

		for i, enemy := range enemies.enemies {
			if i != 0 {
				items = append(items, drawing.FlexVSpacer(drawing.CommonSpacing))
			}

			items = append(items, drawing.Rigid(d.exportBar(styledFonts, enemy, maxDamage)))
		}

		items = append(
			items,
			drawing.FlexVSpacer(drawing.CommonSpacing),
			drawing.Rigid(d.exportBar(
				styledFonts,
				enemyDamage{
					name:   "Indirect Damage",
					amount: 0,
					damage: victim.indirectDamage,
				},
				victim.indirectDamage.Total(),
			)),
		)

		body = drawing.Flex{
			ExpandW: true,
			Axis:    layout.Vertical,
		}.Layout(
			items...,
		)
	case DisplayPie:
		var totalValue int

		pieItems := make([]drawing.PieChartItem, 0, max(1, len(victim.damageFromEnemyTypes.enemies)))
		for i, enemy := range victim.damageFromEnemyTypes.enemies {
			if i >= d.currentLimit.Int() {
				break
			}

			pieItems = append(pieItems, drawing.PieChartItem{
				Name:    enemy.name,
				Value:   enemy.damage.Total(),
				SubText: enemy.damage.StringCL(d.longFormatBool.Value),
			})
			totalValue += enemy.damage.Total()
		}

		style := drawing.PieChart{
			OverflowLimit: 15,
			ColorBoxSize:  drawing.CommonSpacing * 4,
			TextStyle:     styledFonts.Body,
			SubTextStyle:  drawing.MakeTextStyle(styledFonts.Smaller.Face, utils.GrayText),
		}

		body = style.Layout(totalValue, pieItems...)
	case DisplayGraphs:
		items := make([]drawing.FlexChild, 0, len(enemies.enemies)*2-1+6)

		items = append(
			items,
			drawing.Rigid(exportTimeFrame(styledFonts, victim.timeController.CurrentTimeFrame)),
			drawing.FlexVSpacer(drawing.CommonSpacing),
		)

		style := drawing.StyleAreaChart(victim.totalChart, components.StringToColor("Total Damage"))
		style.MinHeight = 200

		items = append(
			items,
			drawing.Rigid(d.exportWidget(
				styledFonts,
				enemyDamage{
					name:   "Total Damage",
					amount: 0,
					damage: victim.totalDamage,
				},
				style.Layout(),
			)),
			drawing.FlexVSpacer(drawing.CommonSpacing),
		)

		for i, enemy := range enemies.enemies {
			if i != 0 {
				items = append(items, drawing.FlexVSpacer(drawing.CommonSpacing))
			}

			style := drawing.StyleAreaChart(enemy.chart, components.StringToColor(enemy.name))
			style.MinHeight = 200

			items = append(items, drawing.Rigid(d.exportWidget(styledFonts, enemy, style.Layout())))
		}

		body = drawing.Flex{
			ExpandW: true,
			Axis:    layout.Vertical,
		}.Layout(
			items...,
		)
	}

	base := layoutTitle(
		styledFonts,
		d.TabName(),
		drawing.HorizontalWrap{
			Alignment:   layout.Middle,
			Spacing:     drawing.CommonSpacing * 3,
			LineSpacing: drawing.CommonSpacing,
		}.Layout(
			func(ltx drawing.Context) drawing.Result {
				if d.currentVictim != "" {
					return styledFonts.Smaller.Layout(fmt.Sprintf("Victim: %v", d.currentVictim))(ltx)
				}

				return styledFonts.Smaller.Layout("Victim: All")(ltx)
			},
			styledFonts.Smaller.Layout(fmt.Sprintf("Display: %v", d.currentDisplay)),
			func(ltx drawing.Context) drawing.Result {
				if d.currentDisplay != DisplayPie {
					return styledFonts.Smaller.Layout(fmt.Sprintf("Group: %v", d.groupByDropdown.Value.(GroupBy)))(ltx)
				}

				return styledFonts.Smaller.Layout(fmt.Sprintf("Limit: %v", d.currentLimit))(ltx)
			},
		),
		drawing.RoundedSurface(
			utils.SecondBG,
			body,
		),
	)

	return drawing.ExportImage(state.Theme(), base, drawing.F64(800, 10000))
}
