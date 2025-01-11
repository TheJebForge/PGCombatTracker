package ui

import (
	"PGCombatTracker/abstract"
	"PGCombatTracker/parser"
	"PGCombatTracker/ui/components"
	"PGCombatTracker/utils"
	"gioui.org/app"
	"gioui.org/widget/material"
	"github.com/samber/lo"
	"image/color"
	"log"
	"slices"
	"time"
)

type GlobalState struct {
	settings            *abstract.Settings
	markers             *abstract.Markers
	gorgonFolder        string
	statisticsFactory   abstract.StatisticsFactory
	statisticsCollector abstract.StatisticsCollector
	page                abstract.Page
	storage             map[string]any
	window              *app.Window
	theme               *material.Theme
	fonts               *abstract.FontPack
	draggable           bool
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
	if err := LoadSettings(sett); err != nil {
		log.Printf("Failed to load %v, continuing from defaults. Reason: %v\n", SettingsLocation, err)
	}

	markers := abstract.NewMarkers()
	if err := LoadMarkersFile(markers); err != nil {
		log.Printf("Failed to load %v, starting fresh. Reason: %v\n", MarkersLocation, err)
	}

	fonts, err := abstract.LoadFontPack()
	if err != nil {
		log.Fatalln(err)
	}

	return &GlobalState{
		settings:          sett,
		markers:           markers,
		gorgonFolder:      gorgonFolder,
		statisticsFactory: factory,
		storage:           make(map[string]any),
		window:            window,
		theme:             theme,
		fonts:             fonts,
		draggable:         true,
	}, nil
}

func (g *GlobalState) FindMarkers(fullPath string) ([]abstract.Marker, error) {
	allMarkers, err := PrereadLogsFile(fullPath)
	if err != nil {
		return nil, err
	}

	for _, userFile := range g.markers.Files {
		if userFile.Path == fullPath {
			for _, userMarker := range userFile.Markers {
				allMarkers = append(allMarkers, userMarker.AsUserDefined())
			}
			break
		}
	}

	slices.SortFunc(allMarkers, func(a, b abstract.Marker) int {
		return a.Time.Compare(b.Time)
	})

	return allMarkers, nil
}

func (g *GlobalState) DeleteMarker(path string, marker abstract.Marker) {
	newFiles := make([]abstract.MarkerFile, 0, len(g.markers.Files))

	for _, file := range g.markers.Files {
		if file.Path == path {
			file.Markers = lo.Filter(file.Markers, func(item abstract.Marker, index int) bool {
				return !item.EqualTo(marker)
			})
		}

		newFiles = append(newFiles, file)
	}

	g.markers.Files = newFiles

	err := SaveMarkersFile(g.markers)
	if err != nil {
		log.Println(err)
	}
}

func (g *GlobalState) SaveMarker(path, name, user string) {
	now := time.Now()
	newMarker := abstract.Marker{
		Time: now,
		Name: name,
		User: user,
	}

	g.markers.Files = utils.CreateUpdate(
		g.markers.Files,
		func(file abstract.MarkerFile) bool {
			return file.Path == path
		},
		func() abstract.MarkerFile {
			return abstract.MarkerFile{
				Path: path,
				Markers: []abstract.Marker{
					newMarker,
				},
			}
		},
		func(file abstract.MarkerFile) abstract.MarkerFile {
			file.Markers = append(file.Markers, newMarker)
			return file
		},
	)

	err := SaveMarkersFile(g.markers)
	if err != nil {
		log.Println(err)
	}
}

func (g *GlobalState) Settings() *abstract.Settings {
	return g.settings
}

func (g *GlobalState) ReloadSettings() {
	err := LoadSettings(g.settings)
	if err != nil {
		log.Printf("Failed to load %v: %v\n", SettingsLocation, err)
	}
}

func (g *GlobalState) SaveSettings() {
	err := SaveSettings(g.settings)
	if err != nil {
		log.Printf("Failed to save %v: %v\n", SettingsLocation, err)
	}
}

func (g *GlobalState) CanBeDragged() bool {
	return g.draggable
}

func (g *GlobalState) SetWindowDrag(value bool) {
	g.draggable = value
}

func (g *GlobalState) OpenFile(path string, watch bool, timeFrames []abstract.MarkerTimeFrame) bool {
	if g.statisticsCollector != nil && g.statisticsCollector.IsAlive() {
		g.statisticsCollector.Close()
		g.statisticsCollector = nil
	}

	stats, err := g.statisticsFactory(g, path, watch, timeFrames)

	if err != nil {
		log.Printf("Encountered an error while trying to start collecting: %v\n", err)
		return false
	}

	stats.Run()
	g.statisticsCollector = stats

	return true
}

func (g *GlobalState) GorgonFolder() string {
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

func (g *GlobalState) FontPack() *abstract.FontPack {
	return g.fonts
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
