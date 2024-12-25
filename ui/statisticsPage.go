package ui

import (
	"PGCombatTracker/abstract"
	"PGCombatTracker/ui/components"
	"PGCombatTracker/utils"
	"gioui.org/app"
	"gioui.org/io/system"
	"gioui.org/layout"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"golang.org/x/exp/shiny/materialdesign/icons"
	"image"
)

type StatisticsPage struct {
	// Actual properties
	currentCollector int

	// UI garbage
	modalLayer          *components.ModalLayer
	backIcon            *widget.Icon
	backButton          *widget.Clickable
	resetIcon           *widget.Icon
	resetButton         *widget.Clickable
	lockButton          *widget.Clickable
	lockIcon            *widget.Icon
	unlockIcon          *widget.Icon
	collectorTabs       *widget.List
	collectorTabButtons []*widget.Clickable
	collectorBody       *widget.List
}

func initializeCollectorTabButtons(state abstract.GlobalState) []*widget.Clickable {
	amount := len(state.StatisticsCollector().Collectors())
	buttons := make([]*widget.Clickable, amount)

	for i := 0; i < amount; i++ {
		buttons[i] = &widget.Clickable{}
	}

	return buttons
}

func getFreshCollectorBody() *widget.List {
	return &widget.List{
		List: layout.List{
			Axis: layout.Vertical,
		},
	}
}

func NewStatisticsPage(state abstract.GlobalState) (*StatisticsPage, error) {
	backIcon, err := widget.NewIcon(icons.NavigationArrowBack)

	if err != nil {
		return nil, err
	}

	resetIcon, err := widget.NewIcon(icons.ActionDelete)

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

	return &StatisticsPage{
		modalLayer:  components.NewModalLayer(),
		backIcon:    backIcon,
		backButton:  &widget.Clickable{},
		resetIcon:   resetIcon,
		resetButton: &widget.Clickable{},
		lockButton:  &widget.Clickable{},
		lockIcon:    lockIcon,
		unlockIcon:  unlockIcon,
		collectorTabs: &widget.List{
			List: layout.List{
				Axis: layout.Horizontal,
			},
		},
		collectorTabButtons: initializeCollectorTabButtons(state),
		collectorBody:       getFreshCollectorBody(),
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

func windowDragArea(draggable bool) layout.Widget {
	return func(gtx layout.Context) layout.Dimensions {
		width := gtx.Dp(60)
		size := image.Point{
			X: width,
			Y: gtx.Constraints.Max.Y - gtx.Dp(10),
		}

		defer clip.UniformRRect(image.Rectangle{Max: size}, 5).Push(gtx.Ops).Pop()

		if draggable {
			system.ActionInputOp(system.ActionMove).Add(gtx.Ops)
		}

		paint.NewImageOp(utils.CheckerImage{Size: size}).Add(gtx.Ops)
		paint.PaintOp{}.Add(gtx.Ops)

		return layout.Dimensions{
			Size: size,
		}
	}
}

func goBack(state abstract.GlobalState) {
	state.StatisticsCollector().Close()
	state.SwitchPage(NewFileSelectionPage())
}

func (s *StatisticsPage) switchCollectorTab(newIndex int) {
	s.currentCollector = newIndex
	s.collectorBody = getFreshCollectorBody()
}

func (s *StatisticsPage) navBar(state abstract.GlobalState) layout.Widget {
	return func(gtx layout.Context) layout.Dimensions {
		if s.lockButton.Clicked(gtx) {
			state.SetWindowDrag(!state.CanBeDragged())
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
						layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
							stats := state.StatisticsCollector()
							collectors := stats.Collectors()
							amount := len(collectors)

							return material.List(state.Theme(), s.collectorTabs).Layout(
								gtx,
								amount,
								func(gtx layout.Context, index int) layout.Dimensions {
									collector := collectors[index]
									button := s.collectorTabButtons[index]

									if button.Clicked(gtx) {
										s.switchCollectorTab(index)
									}

									style := navButton(state, button, collector.TabName())

									if s.currentCollector != index {
										style.Background = utils.LesserContrastBg
									}

									if index+1 == amount {
										return style.Layout(gtx)
									} else {
										return layout.Flex{
											Axis: layout.Horizontal,
										}.Layout(
											gtx,
											layout.Rigid(style.Layout),
											utils.FlexSpacerW(utils.CommonSpacing),
										)
									}
								},
							)
						}),
						utils.FlexSpacerW(utils.CommonSpacing),
						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							return layout.Stack{
								Alignment: layout.Center,
							}.Layout(
								gtx,
								layout.Expanded(windowDragArea(state.CanBeDragged())),
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

func (s *StatisticsPage) body(state abstract.GlobalState) layout.Widget {
	return s.modalLayer.Overlay(func(gtx layout.Context) layout.Dimensions {
		stats := state.StatisticsCollector()
		lock := stats.Mutex()

		lock.RLock()
		defer lock.RUnlock()

		collectors := stats.Collectors()

		if s.currentCollector >= len(collectors) {
			return layout.Dimensions{}
		}

		body := collectors[s.currentCollector].UI(
			NewLayeredState(state, s.modalLayer),
		)

		return material.List(state.Theme(), s.collectorBody).Layout(
			gtx,
			len(body),
			func(gtx layout.Context, index int) layout.Dimensions {
				return body[index](gtx)
			},
		)
	})
}

func (s *StatisticsPage) Layout(ctx layout.Context, state abstract.GlobalState) error {
	if s.backButton.Clicked(ctx) {
		goBack(state)
	}

	if s.resetButton.Clicked(ctx) {
		if stats := state.StatisticsCollector(); stats != nil && stats.IsAlive() {
			stats.Reset()
		}
	}

	layout.Flex{
		Axis: layout.Vertical,
	}.Layout(
		ctx,
		layout.Rigid(s.navBar(state)),
		layout.Flexed(1, s.body(state)),
	)

	return nil
}

func (s *StatisticsPage) SetupWindow(state abstract.GlobalState) {
	state.Window().Option(
		app.Decorated(false),
		app.MinSize(350, 350),
		app.Size(350, 350),
	)
}
