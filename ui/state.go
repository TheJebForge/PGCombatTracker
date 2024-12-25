package ui

import (
	"PGCombatTracker/abstract"
	"PGCombatTracker/parser"
	"PGCombatTracker/ui/components"
	"PGCombatTracker/utils"
	"bytes"
	"encoding/json"
	"gioui.org/app"
	"gioui.org/widget/material"
	"github.com/spf13/viper"
	"image/color"
	"log"
)

type GlobalState struct {
	settings            *abstract.Settings
	gorgonFolder        string
	statisticsFactory   abstract.StatisticsFactory
	statisticsCollector abstract.StatisticsCollector
	page                abstract.Page
	storage             map[string]any
	window              *app.Window
	theme               *material.Theme
	draggable           bool
}

const settingsLocation = "settings.json"

func loadSettings(settings *abstract.Settings) error {
	viper.SetConfigFile(settingsLocation)
	err := viper.ReadInConfig()
	if err != nil {
		return err
	}

	err = viper.Unmarshal(settings)
	if err != nil {
		return err
	}

	return nil
}

func saveSettings(settings *abstract.Settings) error {
	bs, err := json.Marshal(settings)
	if err != nil {
		return err
	}

	err = viper.ReadConfig(bytes.NewReader(bs))
	if err != nil {
		return err
	}

	err = viper.WriteConfig()
	if err != nil {
		return err
	}

	return nil
}

func NewGlobalState(window *app.Window, factory abstract.StatisticsFactory) (abstract.GlobalState, error) {
	gorgonFolder, err := parser.GetGorgonFolder()

	if err != nil {
		return nil, err
	}

	theme := material.NewTheme()

	theme.Bg = utils.BG
	theme.Fg = color.NRGBA{R: 255, G: 255, B: 255, A: 255}
	theme.ContrastBg = color.NRGBA{R: 100, G: 100, B: 100, A: 255}

	sett := abstract.NewSettings()
	if err := loadSettings(sett); err != nil {
		log.Printf("Failed to load %v, continuing from defaults. Reason: %v\n", settingsLocation, err)

		err := saveSettings(sett)
		if err != nil {
			log.Printf("Failed to save %v: %v\n", settingsLocation, err)
		}
	}

	return &GlobalState{
		settings:          sett,
		gorgonFolder:      gorgonFolder,
		statisticsFactory: factory,
		storage:           make(map[string]any),
		window:            window,
		theme:             theme,
		draggable:         true,
	}, nil
}

func (g *GlobalState) ReloadSettings() {
	err := loadSettings(g.settings)
	if err != nil {
		log.Printf("Failed to load %v: %v\n", settingsLocation, err)
	}
}

func (g *GlobalState) SaveSettings() {
	err := saveSettings(g.settings)
	if err != nil {
		log.Printf("Failed to save %v: %v\n", settingsLocation, err)
	}
}

func (g *GlobalState) CanBeDragged() bool {
	return g.draggable
}

func (g *GlobalState) SetWindowDrag(value bool) {
	g.draggable = value
}

func (g *GlobalState) OpenFile(path string, watch bool) bool {
	if g.statisticsCollector != nil && g.statisticsCollector.IsAlive() {
		g.statisticsCollector.Close()
		g.statisticsCollector = nil
	}

	stats, err := g.statisticsFactory(path, watch)

	if err != nil {
		log.Printf("Encountered an error while trying to start collecting: %v\n", err)
		return false
	}

	stats.Run()
	g.statisticsCollector = stats

	return true
}

func (g *GlobalState) Settings() *abstract.Settings {
	return g.settings
}

func (g *GlobalState) GordonFolder() string {
	return g.gorgonFolder
}

func (g *GlobalState) StatisticsCollector() abstract.StatisticsCollector {
	return g.statisticsCollector
}

func (g *GlobalState) SetStatisticsCollector(collector abstract.StatisticsCollector) {
	g.statisticsCollector = collector
}

func (g *GlobalState) Page() abstract.Page {
	return g.page
}

func (g *GlobalState) Storage() map[string]any {
	return g.storage
}

func (g *GlobalState) Window() *app.Window {
	return g.window
}

func (g *GlobalState) Theme() *material.Theme {
	return g.theme
}

func (g *GlobalState) SwitchPage(page abstract.Page) {
	g.page = page
	page.SetupWindow(g)
}

func NewLayeredState(state abstract.GlobalState, modalLayer *components.ModalLayer) *LayeredState {
	// Straight up casting to global state because nothing else will be implementing this anyway
	globalState := state.(*GlobalState)

	return &LayeredState{
		GlobalState: globalState,
		modalLayer:  modalLayer,
	}
}

type LayeredState struct {
	*GlobalState
	modalLayer *components.ModalLayer
}

func (l LayeredState) ModalLayer() *components.ModalLayer {
	return l.modalLayer
}
