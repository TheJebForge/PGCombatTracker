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

type healing struct {
	name      string
	amount    int
	recovered abstract.Vitals
	chart     *components.TimeBasedChart
}

type healingWithMax struct {
	subjects       []healing
	timeController *components.TimeController
	total          abstract.Vitals
	max            abstract.Vitals
	maxRange       components.DataRange
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

func freshHealingWithMax() healingWithMax {
	return healingWithMax{
		timeController: components.NewTimeControllerOrCrash(components.NewTimeBasedChart("Total")),
	}
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

	displayDropdown, err := components.NewDropdown(
		"Display",
		DisplayBars,
		DisplayPie,
		DisplayGraphs,
	)
	if err != nil {
		log.Fatalln(err)
	}

	return &HealingCollector{
		allies:            freshHealingWithMax(),
		enemies:           freshHealingWithMax(),
		enemyTypes:        freshHealingWithMax(),
		allWithEnemies:    freshHealingWithMax(),
		allWithEnemyTypes: freshHealingWithMax(),

		subjectDropdown:    subjectDropdown,
		displayDropdown:    displayDropdown,
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
	currentDisplay     displayChoice
	subjectDropdown    *components.Dropdown
	displayDropdown    *components.Dropdown
	enemyTypesCheckbox *widget.Bool
	longFormatBool     *widget.Bool
}

func (h *HealingCollector) Reset(info abstract.StatisticsInformation) {
	h.allies = freshHealingWithMax()
	h.enemies = freshHealingWithMax()
	h.enemyTypes = freshHealingWithMax()
	h.allWithEnemies = freshHealingWithMax()
	h.allWithEnemyTypes = freshHealingWithMax()
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
			chart := components.NewTimeBasedChart(subject)
			chart.Add(components.TimePoint{
				Time:    event.Time,
				Value:   recovered.Healed.Total(),
				Details: recovered.Healed,
			})

			return healing{
				name:      subject,
				amount:    1,
				recovered: recovered.Healed,
				chart:     chart,
			}
		}
	}
	updateHeal := func(heal healing) healing {
		heal.amount++
		heal.recovered = heal.recovered.Add(recovered.Healed)
		heal.chart.Add(components.TimePoint{
			Time:    event.Time,
			Value:   heal.recovered.Total(),
			Details: recovered.Healed,
		})

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
		stat.total = stat.total.Add(recovered.Healed)
		stat.timeController.Add(components.TimePoint{
			Time:  event.Time,
			Value: stat.total.Total(),
		})
		slices.SortFunc(stat.subjects, healSort)
		stat.max = slices.MaxFunc(stat.subjects, healMax).recovered
		stat.maxRange = stat.maxRange.Expand(stat.max.Total())
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

func (h *HealingCollector) drawWidget(state abstract.LayeredState, healed healing, widget layout.Widget, size unit.Dp) layout.Widget {
	return drawUniversalStatsText(
		state, healed.recovered,
		widget, healed.amount,
		healed.name, "recovered %v times",
		size, h.longFormatBool.Value,
	)
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

	if h.displayDropdown.Changed() {
		h.currentDisplay = h.displayDropdown.Value.(displayChoice)
	}

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
				return components.HorizontalWrap{
					Alignment:   layout.Middle,
					Spacing:     utils.CommonSpacing,
					LineSpacing: utils.CommonSpacing,
				}.Layout(
					gtx,
					defaultDropdownStyle(state, h.subjectDropdown).Layout,
					defaultCheckboxStyle(state, h.enemyTypesCheckbox, "Group enemy types").Layout,
					defaultCheckboxStyle(state, h.longFormatBool, "Use long numbers").Layout,
					defaultDropdownStyle(state, h.displayDropdown).Layout,
				)
			}),
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				if h.currentDisplay == DisplayGraphs {
					return layout.Flex{
						Axis: layout.Vertical,
					}.Layout(
						gtx,
						utils.FlexSpacerH(utils.CommonSpacing),
						layout.Rigid(components.StyleTimeController(state.Theme(), stats.timeController).Layout),
					)
				}

				return layout.Dimensions{}
			}),
		)
	})

	var widgets []layout.Widget

	switch h.currentDisplay {
	case DisplayBars:
		maxHealed := stats.max.Total()

		for _, healed := range stats.subjects {
			widgets = append(widgets, h.drawBar(state, healed, maxHealed, 40))
		}
	case DisplayPie:
		widgets = append(widgets, func(gtx layout.Context) layout.Dimensions {
			var totalValue int
			pieItems := make([]components.PieChartItem, 0, max(1, len(stats.subjects)))
			for _, subject := range stats.subjects {
				pieItems = append(pieItems, components.PieChartItem{
					Name:    subject.name,
					Value:   subject.recovered.Total(),
					SubText: subject.recovered.StringCL(h.longFormatBool.Value),
				})
				totalValue += subject.recovered.Total()
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
		controller := stats.timeController

		for _, subject := range stats.subjects {
			subject.chart.DisplayTimeFrame = controller.CurrentTimeFrame
			subject.chart.DisplayValueRange = stats.maxRange

			chartStyle := components.StyleTimeBasedChart(state.Theme(), subject.chart)
			chartStyle.Color = components.StringToColor(subject.name)
			chartStyle.LongFormat = h.longFormatBool.Value

			widgets = append(widgets, h.drawWidget(state, subject, chartStyle.Layout, 100))
		}
	}

	return topWidget, widgets
}

func (h *HealingCollector) exportWidget(styledFonts *drawing.StyledFontPack, healed healing, widget drawing.Widget) drawing.Widget {
	return exportUniversalStatsTextAsStack(
		styledFonts, healed.recovered,
		widget, healed.amount,
		healed.name, "recovered %v times",
		h.longFormatBool.Value,
	)
}

func (h *HealingCollector) exportBar(styledFonts *drawing.StyledFontPack, healed healing, maxDamage int) drawing.Widget {
	return exportUniversalBar(
		styledFonts, healed.recovered,
		healed.recovered.Total(), maxDamage, healed.amount,
		healed.name, "recovered %v times",
		h.longFormatBool.Value,
	)
}

func (h *HealingCollector) Export(state abstract.LayeredState) image.Image {
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

	styledFonts := drawing.StyleFontPack(state.FontPack(), state.Theme().Fg)

	body := drawing.Empty

	switch h.currentDisplay {
	case DisplayBars:
		items := make([]drawing.FlexChild, 0, len(stats.subjects)*2-1)

		maxRecovered := stats.max.Total()

		for i, healed := range stats.subjects {
			if i != 0 {
				items = append(items, drawing.FlexVSpacer(drawing.CommonSpacing))
			}

			items = append(items, drawing.Rigid(h.exportBar(styledFonts, healed, maxRecovered)))
		}

		body = drawing.Flex{
			ExpandW: true,
			Axis:    layout.Vertical,
		}.Layout(
			items...,
		)
	case DisplayPie:
		var totalValue int

		pieItems := make([]drawing.PieChartItem, 0, max(1, len(stats.subjects)))
		for _, healed := range stats.subjects {
			pieItems = append(pieItems, drawing.PieChartItem{
				Name:    healed.name,
				Value:   healed.recovered.Total(),
				SubText: healed.recovered.StringCL(h.longFormatBool.Value),
			})
			totalValue += healed.recovered.Total()
		}

		style := drawing.PieChart{
			OverflowLimit: 15,
			ColorBoxSize:  drawing.CommonSpacing * 4,
			TextStyle:     styledFonts.Body,
			SubTextStyle:  drawing.MakeTextStyle(styledFonts.Smaller.Face, utils.GrayText),
		}

		body = style.Layout(totalValue, pieItems...)
	case DisplayGraphs:
		items := make([]drawing.FlexChild, 0, len(stats.subjects)*2-1+2)

		items = append(
			items,
			drawing.Rigid(exportTimeFrame(styledFonts, stats.timeController.CurrentTimeFrame)),
			drawing.FlexVSpacer(drawing.CommonSpacing),
		)

		for i, healed := range stats.subjects {
			if i != 0 {
				items = append(items, drawing.FlexVSpacer(drawing.CommonSpacing))
			}

			style := drawing.StyleAreaChart(healed.chart, components.StringToColor(healed.name))
			style.MinHeight = 200

			items = append(items, drawing.Rigid(h.exportWidget(styledFonts, healed, style.Layout())))
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
		h.TabName(),
		drawing.HorizontalWrap{
			Alignment:   layout.Middle,
			Spacing:     drawing.CommonSpacing * 3,
			LineSpacing: drawing.CommonSpacing,
		}.Layout(
			styledFonts.Smaller.Layout(fmt.Sprintf("Subject: %v", h.currentSubject)),
			styledFonts.Smaller.Layout(fmt.Sprintf("Display: %v", h.currentDisplay)),
			func(ltx drawing.Context) drawing.Result {
				if h.enemyTypesCheckbox.Value {
					return styledFonts.Smaller.Layout("Grouping by enemy type")(ltx)
				}

				return drawing.Empty(ltx)
			},
		),
		drawing.RoundedSurface(
			utils.SecondBG,
			body,
		),
	)

	return drawing.ExportImage(state.Theme(), base, drawing.F64(800, 10000))
}
