package abstract

import (
	"PGCombatTracker/ui/components"
	"gioui.org/app"
	"gioui.org/widget/material"
)

type GlobalState interface {
	Settings() *Settings
	ReloadSettings()
	SaveSettings()

	GorgonFolder() string
	StatisticsCollector() StatisticsCollector
	Page() Page
	Storage() map[string]any
	Window() *app.Window
	Theme() *material.Theme

	CanBeDragged() bool
	SetWindowDrag(value bool)

	OpenFile(path string, watch bool) bool
	SwitchPage(page Page)
}

type LayeredState interface {
	GlobalState
	ModalLayer() *components.ModalLayer
}
