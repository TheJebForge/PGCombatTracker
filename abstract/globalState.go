package abstract

import (
	"PGCombatTracker/ui/components"
	"gioui.org/app"
	"gioui.org/widget/material"
)

type GlobalState interface {
	SettingsBearer
	MarkersBearer
	StatisticsBearer
	PageSwitcher

	GorgonFolder() string

	Storage() map[string]any
	Window() *app.Window
	Theme() *material.Theme

	CanBeDragged() bool
	SetWindowDrag(value bool)
}

type SettingsBearer interface {
	Settings() *Settings
	ReloadSettings()
	SaveSettings()
}

type MarkersBearer interface {
	FindMarkers(path string) ([]Marker, error)
	DeleteMarker(path string, marker Marker)
	SaveMarker(path, name, user string)
}

type StatisticsBearer interface {
	StatisticsCollector() StatisticsCollector
	OpenFile(path string, watch bool, timeFrames []MarkerTimeFrame) bool
}

type PageSwitcher interface {
	Page() Page
	SwitchPage(page Page)
}

type LayeredState interface {
	GlobalState
	ModalLayer() *components.ModalLayer
}
