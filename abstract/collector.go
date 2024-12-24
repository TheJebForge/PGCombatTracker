package abstract

import (
	"gioui.org/layout"
)

type Collector interface {
	Reset()
	Collect(info StatisticsInformation, event *ChatEvent) error
	TabName() string
	UI(state GlobalState) []layout.Widget
}
