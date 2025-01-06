package abstract

import (
	"gioui.org/layout"
	"time"
)

type Collector interface {
	Reset()
	Tick(info StatisticsInformation, at time.Time)
	Collect(info StatisticsInformation, event *ChatEvent) error
	TabName() string
	UI(state LayeredState) (layout.Widget, []layout.Widget)
}
