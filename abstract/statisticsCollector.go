package abstract

import "sync"

type StatisticsCollector interface {
	SaveMarker(state GlobalState, name string)
	Mutex() *sync.RWMutex
	Reset()
	Collectors() []Collector
	Notify() chan bool
	Run()
	IsAlive() bool
	Close()
}

type StatisticsInformation interface {
	CurrentUsername() string
	Settings() *Settings
}

type StatisticsFactory func(state GlobalState, path string, watch bool, timeFrames []MarkerTimeFrame) (StatisticsCollector, error)
