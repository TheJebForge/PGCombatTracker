package ui

import (
	"PGCombatTracker/abstract"
	"PGCombatTracker/ui/components"
	"PGCombatTracker/utils"
	"bytes"
	"fmt"
	"gioui.org/app"
	"gioui.org/io/system"
	"gioui.org/layout"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"github.com/samber/lo"
	"golang.design/x/clipboard"
	"golang.org/x/exp/shiny/materialdesign/icons"
	"image"
	"image/png"
	"log"
)

type StatisticsPage struct {
	// Actual properties
	currentCollector int

	// UI garbage
	filePath          string
	modalLayer        *components.ModalLayer
	backIcon          *widget.Icon
	backButton        *widget.Clickable
	resetIcon         *widget.Icon
	resetButton       *widget.Clickable
	addIcon           *widget.Icon
	addButton         *widget.Clickable
	lockButton        *widget.Clickable
	lockIcon          *widget.Icon
	unlockIcon        *widget.Icon
	copyIcon          *widget.Icon
	copyButton        *widget.Clickable
	collectorDropdown *components.Dropdown
	collectorBody     *widget.List
	windowedButton    *widget.Clickable
	windowedIcon      *widget.Icon
}

type CollectorPageIndex struct {
	index int
	name  string
}

func (c CollectorPageIndex) String() string {
	return c.name
}

func getFreshCollectorBody() *widget.List {
	return &widget.List{
		List: layout.List{
			Axis: layout.Vertical,
		},
	}
}

func NewStatisticsPage(state abstract.GlobalState, filePath string) (*StatisticsPage, error) {
	backIcon, err := widget.NewIcon(icons.NavigationArrowBack)

	if err != nil {
		return nil, err
	}

	resetIcon, err := widget.NewIcon(icons.ActionDelete)

	if err != nil {
		return nil, err
	}

	addIcon, err := widget.NewIcon(icons.ContentFlag)

	if err != nil {
		return nil, err
	}

	lockIcon, err := widget.NewIcon(icons.ActionLockOutline)

	if err != nil {
		return nil, err
	}

	unlockIcon, err := widget.NewIcon(icons.ActionLockOpen)

	if err != nil {
		return nil, err
	}

	copyIcon, err := widget.NewIcon(icons.ContentContentCopy)

	if err != nil {
		return nil, err
	}

	windowedIcon, err := widget.NewIcon(icons.ActionFlipToFront)

	if err != nil {
		return nil, err
	}

	collectorDropdown, err := components.NewDropdown("Page", CollectorPageIndex{})

	if err != nil {
		return nil, err
	}

	options := lo.Map(
		state.StatisticsCollector().Collectors(),
		func(item abstract.Collector, index int) fmt.Stringer {
			return CollectorPageIndex{
				index: index,
				name:  item.TabName(),
			}
		},
	)

	collectorDropdown.Value = options[0]
	collectorDropdown.SetOptions(options)

	return &StatisticsPage{
		filePath:          filePath,
		modalLayer:        components.NewModalLayer(),
		backIcon:          backIcon,
		backButton:        &widget.Clickable{},
		addIcon:           addIcon,
		addButton:         &widget.Clickable{},
		resetIcon:         resetIcon,
		resetButton:       &widget.Clickable{},
		lockButton:        &widget.Clickable{},
		lockIcon:          lockIcon,
		unlockIcon:        unlockIcon,
		copyIcon:          copyIcon,
		copyButton:        &widget.Clickable{},
		collectorDropdown: collectorDropdown,
		collectorBody:     getFreshCollectorBody(),
		windowedButton:    &widget.Clickable{},
		windowedIcon:      windowedIcon,
	}, nil
}

func navButton(state abstract.GlobalState, button *widget.Clickable, txt string) material.ButtonStyle {
	style := material.Button(state.Theme(), button, txt)
	style.Inset = layout.UniformInset(7)
	style.TextSize = 12
	return style
}

func navIconButton(state abstract.GlobalState, button *widget.Clickable, icon *widget.Icon, desc string) material.IconButtonStyle {
	style := material.IconButton(state.Theme(), button, icon, desc)
	style.Inset = layout.UniformInset(5)
	style.Size = 20
	return style
}

func (s *StatisticsPage) windowDragArea(state abstract.LayeredState) layout.Widget {
	return func(gtx layout.Context) layout.Dimensions {
		width := gtx.Dp(60)
		size := image.Point{
			X: width,
			Y: gtx.Constraints.Max.Y - gtx.Dp(10),
		}

		defer clip.UniformRRect(image.Rectangle{Max: size}, 5).Push(gtx.Ops).Pop()

		if state.CanBeDragged() {
			system.ActionInputOp(system.ActionMove).Add(gtx.Ops)
		}

		paint.NewImageOp(utils.CheckerImage{Size: size}).Add(gtx.Ops)
		paint.PaintOp{}.Add(gtx.Ops)

		return layout.Dimensions{
			Size: size,
		}
	}
}

func (s *StatisticsPage) goBack(state abstract.GlobalState) {
	markers, err := state.FindMarkers(s.filePath)

	if err != nil {
		log.Printf("Failed to open markers page: %v\n", err)
		return
	}

	state.StatisticsCollector().Close()
	state.SwitchPage(NewMarkersPage(s.filePath, markers))
}

func (s *StatisticsPage) switchCollectorTab(newIndex int) {
	s.currentCollector = newIndex
	s.collectorBody = getFreshCollectorBody()
}

func (s *StatisticsPage) navBar(state abstract.LayeredState) layout.Widget {
	return func(gtx layout.Context) layout.Dimensions {
		if s.lockButton.Clicked(gtx) {
			state.SetWindowDrag(!state.CanBeDragged())
		}

		if s.windowedButton.Clicked(gtx) {
			state.Window().Option(app.Windowed.Option())
		}

		return layout.Background{}.Layout(
			gtx,
			utils.MakeColoredAndOptionalDragBG(utils.SecondBG, state.CanBeDragged()),
			func(gtx layout.Context) layout.Dimensions {
				return utils.LayoutDefinedHeight(gtx, gtx.Dp(40), func(gtx layout.Context) layout.Dimensions {
					return layout.Flex{
						Axis: layout.Horizontal,
					}.Layout(
						gtx,
						layout.Rigid(navIconButton(state, s.backButton, s.backIcon, "Back").Layout),
						utils.FlexSpacerW(utils.CommonSpacing),
						layout.Rigid(navIconButton(state, s.resetButton, s.resetIcon, "Reset").Layout),
						utils.FlexSpacerW(utils.CommonSpacing),
						layout.Rigid(navIconButton(state, s.addButton, s.addIcon, "Add").Layout),
						utils.FlexSpacerW(utils.CommonSpacing),
						layout.Rigid(navIconButton(state, s.copyButton, s.copyIcon, "Copy").Layout),
						utils.FlexSpacerW(utils.CommonSpacing),
						layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
							if s.collectorDropdown.Changed() {
								value := s.collectorDropdown.Value.(CollectorPageIndex)
								s.switchCollectorTab(value.index)
							}

							style := components.StyleDropdown(state.Theme(), state.ModalLayer(), s.collectorDropdown)
							style.NoLabel = true
							style.Inset = layout.UniformInset(7)
							style.TextSize = 12
							style.DialogTextSize = 12

							return style.Layout(gtx)
						}),
						utils.FlexSpacerW(utils.CommonSpacing),
						layout.Rigid(navIconButton(state, s.windowedButton, s.windowedIcon, "Windowed").Layout),
						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							return layout.Stack{
								Alignment: layout.Center,
							}.Layout(
								gtx,
								layout.Expanded(s.windowDragArea(state)),
								layout.Stacked(func(gtx layout.Context) layout.Dimensions {
									if !state.CanBeDragged() {
										return material.Label(state.Theme(), 12, "Locked").Layout(gtx)
									} else {
										return layout.Dimensions{}
									}
								}),
							)
						}),
						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							if state.CanBeDragged() {
								return navIconButton(state, s.lockButton, s.unlockIcon, "Disable Drag").Layout(gtx)
							} else {
								return navIconButton(state, s.lockButton, s.lockIcon, "Enable Drag").Layout(gtx)
							}
						}),
					)
				})
			},
		)

	}
}

func (s *StatisticsPage) exportToClipboard(state abstract.LayeredState, collector abstract.Collector) {
	buf := &bytes.Buffer{}

	img := collector.Export(state)

	err := png.Encode(buf, img)
	if err != nil {
		log.Println("Failed to export image", err)
		return
	}

	clipboard.Write(clipboard.FmtImage, buf.Bytes())
}

func (s *StatisticsPage) body(state abstract.LayeredState, currentCollector abstract.Collector) layout.Widget {
	return func(gtx layout.Context) layout.Dimensions {
		top, body := currentCollector.UI(state)

		return layout.Flex{
			Axis: layout.Vertical,
		}.Layout(
			gtx,
			layout.Rigid(top),
			utils.FlexSpacerH(utils.CommonSpacing),
			layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
				return layout.Inset{
					Left: utils.CommonSpacing, Right: utils.CommonSpacing,
				}.Layout(
					gtx,
					func(gtx layout.Context) layout.Dimensions {
						return material.List(state.Theme(), s.collectorBody).Layout(
							gtx,
							len(body),
							func(gtx layout.Context, index int) layout.Dimensions {
								return body[index](gtx)
							},
						)
					},
				)
			}),
		)
	}
}

func (s *StatisticsPage) Layout(ctx layout.Context, state abstract.GlobalState) error {
	if s.backButton.Clicked(ctx) {
		s.goBack(state)
	}

	if s.addButton.Clicked(ctx) {
		dialog := components.NewInputDialog(
			state.Theme(),
			"Create New Marker",
			"Input name for the marker",
			"Unnamed",
			func(marker string) {
				if stats := state.StatisticsCollector(); stats != nil {
					if marker == "" {
						marker = "Unnamed"
					}

					stats.SaveMarker(state, marker)
				}
			},
		)
		dialog.TextSize = 12
		dialog.Open(s.modalLayer)
	}

	if s.resetButton.Clicked(ctx) {
		if stats := state.StatisticsCollector(); stats != nil && stats.IsAlive() {
			stats.Reset()
		}
	}

	layeredState := NewLayeredState(state, s.modalLayer)

	s.modalLayer.Overlay(func(gtx layout.Context) layout.Dimensions {
		stats := state.StatisticsCollector()
		lock := stats.Mutex()

		lock.RLock()
		defer lock.RUnlock()

		collectors := stats.Collectors()

		if s.currentCollector < len(collectors) {
			currentCollector := collectors[s.currentCollector]

			if s.copyButton.Clicked(gtx) {
				log.Println("trying to copy to clipboard")
				s.exportToClipboard(layeredState, currentCollector)
			}
		}

		return layout.Flex{
			Axis: layout.Vertical,
		}.Layout(
			ctx,
			layout.Rigid(s.navBar(layeredState)),
			layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
				if s.currentCollector >= len(collectors) {
					return layout.Dimensions{}
				}

				return s.body(layeredState, collectors[s.currentCollector])(gtx)
			}),
		)
	})(ctx)

	return nil
}

func (s *StatisticsPage) SetupWindow(state abstract.GlobalState) {
	state.Window().Option(
		app.Decorated(false),
		app.MinSize(400, 350),
		app.Size(400, 350),
	)
}
