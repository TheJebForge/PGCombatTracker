package abstract

import (
	"gioui.org/layout"
	"image"
	"time"
)

type Collector interface {
	Reset(info StatisticsInformation)
	Tick(info StatisticsInformation, at time.Time)
	Collect(info StatisticsInformation, event *ChatEvent) error
	TabName() string
	UI(state LayeredState) (layout.Widget, []layout.Widget)
	Export(state LayeredState) image.Image
}
