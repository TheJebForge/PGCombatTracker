package abstract

import (
	"gioui.org/app"
	"gioui.org/widget/material"
)

type GlobalState interface {
	Settings() *Settings
	ReloadSettings()
	SaveSettings()

	GordonFolder() string
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
