package main

import (
	"PGCombatTracker/abstract"
	"PGCombatTracker/collectors"
	"PGCombatTracker/ui"
	"PGCombatTracker/utils"
	"gioui.org/app"
	"gioui.org/layout"
	"gioui.org/op"
	"log"
	"os"
	"time"
)

func main() {
	go func() {
		window := new(app.Window)

		window.Option(
			app.Title("PGCombatTracker"),
		)

		err := run(window)
		if err != nil {
			log.Fatal(err)
		}
		os.Exit(0)
	}()
	app.Main()
}

func run(window *app.Window) error {
	state, err := ui.NewGlobalState(
		window,
		func(state abstract.GlobalState, path string, watch bool, timeFrames []abstract.MarkerTimeFrame) (abstract.StatisticsCollector, error) {
			return collectors.NewStatisticsCollector(state, path, watch, timeFrames)
		},
	)

	if err != nil {
		return err
	}

	state.SwitchPage(ui.NewFileSelectionPage())

	go func() {
		for {
			stats := state.StatisticsCollector()

			if stats != nil && stats.IsAlive() {
				<-stats.Notify()
				window.Invalidate()
				continue
			}

			time.Sleep(100 * time.Millisecond)
		}
	}()

	var ops op.Ops
	for {
		switch e := window.Event().(type) {
		case app.DestroyEvent:
			return e.Err
		case app.FrameEvent:
			// This graphics context is used for managing the rendering state.
			gtx := app.NewContext(&ops, e)

			layout.Background{}.Layout(
				gtx,
				utils.MakeColoredAndOptionalDragBG(state.Theme().Bg, state.CanBeDragged()),
				func(gtx layout.Context) layout.Dimensions {
					if state.Page != nil {
						err := state.Page().Layout(gtx, state)
						if err != nil {
							log.Printf("Error updating UI: %v\n", err)
						}
					}

					return layout.Dimensions{Size: gtx.Constraints.Min}
				},
			)

			// Pass the drawing operations to the GPU.
			e.Frame(gtx.Ops)
		}
	}
}
