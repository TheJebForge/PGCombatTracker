package collectors

import (
	"PGCombatTracker/abstract"
	"PGCombatTracker/parser"
	"bufio"
	"io"
	"log"
	"math"
	"os"
	"sync"
	"sync/atomic"
	"time"
)

type StatisticsCollector struct {
	settings   *abstract.Settings
	username   string
	timeFrames []abstract.MarkerTimeFrame
	dead       *atomic.Bool
	collectors []abstract.Collector
	quit       chan bool
	watch      bool
	fullPath   string
	file       *os.File
	reader     *bufio.Reader
	lock       *sync.RWMutex
	locked     bool
	notify     chan bool
}

func NewStatisticsCollector(state abstract.GlobalState, path string, watchFile bool, timeFrames []abstract.MarkerTimeFrame) (*StatisticsCollector, error) {
	file, err := os.Open(path)

	if err != nil {
		return nil, err
	}

	return &StatisticsCollector{
		settings: state.Settings(),
		collectors: []abstract.Collector{
			NewDamageDealtCollector(),
			NewDamageTakenCollector(),
			NewHealingCollector(),
			NewSkillsCollector(),
			NewLevelingCollector(),
			NewMiscCollector(),
		},
		timeFrames: timeFrames,
		dead:       &atomic.Bool{},
		quit:       make(chan bool, 1),
		watch:      watchFile,
		fullPath:   path,
		file:       file,
		reader:     bufio.NewReader(file),
		lock:       new(sync.RWMutex),
		notify:     make(chan bool),
	}, nil
}

func (stats *StatisticsCollector) SaveMarker(state abstract.GlobalState, name string) {
	stats.lock.Lock()
	state.SaveMarker(stats.fullPath, name, stats.username)
	stats.lock.Unlock()
}

func (stats *StatisticsCollector) Settings() *abstract.Settings {
	return stats.settings
}

func (stats *StatisticsCollector) CurrentUsername() string {
	return stats.username
}

func (stats *StatisticsCollector) Collectors() []abstract.Collector {
	return stats.collectors
}

func (stats *StatisticsCollector) Mutex() *sync.RWMutex {
	return stats.lock
}

func (stats *StatisticsCollector) Notify() chan bool {
	return stats.notify
}

func (stats *StatisticsCollector) lockTheLock() {
	if !stats.locked {
		stats.lock.Lock()
		stats.locked = true
	}
}

func (stats *StatisticsCollector) unlockTheLock() {
	if stats.locked {
		stats.locked = false
		stats.lock.Unlock()
		stats.notify <- true
	}
}

func (stats *StatisticsCollector) Reset() {
	stats.lock.Lock()

	for _, collector := range stats.collectors {
		collector.Reset()
	}

	stats.lock.Unlock()
}

func checkIfHasId(name string) bool {
	return name == SplitOffId(name)
}

func (stats *StatisticsCollector) FindUsername(event *abstract.ChatEvent) string {
	switch c := event.Contents.(type) {
	case *abstract.SkillUse:
		if checkIfHasId(c.Subject) {
			return c.Subject
		}

		if checkIfHasId(c.Victim) {
			return c.Victim
		}
	case *abstract.IndirectDamage:
		if checkIfHasId(c.Subject) {
			return c.Subject
		}
	case *abstract.Recovered:
		if checkIfHasId(c.Subject) {
			return c.Subject
		}
	}
	return ""
}

func (stats *StatisticsCollector) Run() {
	fileName := stats.file.Name()

	log.Printf("Starting to read file at '%v'\n", fileName)

	var nextTick time.Time
	tickIntervalDuration := time.Microsecond * time.Duration(math.Round(stats.settings.TickIntervalSeconds*1000000))

	tickIfNeeded := func(at time.Time) {
		for nextTick.Before(at) {
			stats.lockTheLock()

			for _, collector := range stats.collectors {
				collector.Tick(stats, nextTick)
			}

			nextTick = nextTick.Add(tickIntervalDuration)
		}
	}

	go func() {
		firstRead := true
		var lastWithin bool
	infinite:
		for {
			select {
			case <-stats.quit:
				break infinite
			default:
				line, err := stats.reader.ReadString('\n')

				// Don't act on the error, just log it
				if err != nil {
					stats.unlockTheLock()

					if err == io.EOF {
						if stats.watch {
							tickIfNeeded(time.Now())
							firstRead = false
							time.Sleep(100 * time.Millisecond)
							continue infinite
						} else {
							break infinite
						}
					} else {
						log.Printf("Encountered an error while reading line: %v\n", err)
						break infinite
					}
				}

				event := parser.ParseLine(line)

				if event == nil {
					continue infinite
				}

				if nextTick.IsZero() {
					nextTick = event.Time
				}

				// Check timeframe stuff
				within, timeFrameUser := abstract.WithinTimeFrames(stats.timeFrames, event.Time)

				if firstRead {
					if (within != lastWithin) && within {
						stats.lockTheLock()
						stats.username = timeFrameUser
						stats.unlockTheLock()
					}

					if !within {
						continue infinite
					}
				}

				lastWithin = within

				// Grab username from login if detected
				if login, ok := event.Contents.(*abstract.Login); ok && login != nil {
					log.Printf("Detected login as %v\n", login.Name)
					stats.username = login.Name
					continue infinite
				}

				// If username is still empty, try to find it
				if stats.username == "" {
					stats.username = stats.FindUsername(event)
				}

				//log.Println(event)

				stats.lockTheLock()
				tickIfNeeded(event.Time)

				for _, collector := range stats.collectors {
					err = collector.Collect(stats, event)

					if err != nil {
						log.Printf(
							"Collector '%v' encountered an error while ingesting line: %v\n",
							collector.TabName(),
							err,
						)
					}
				}
			}
		}

		stats.unlockTheLock()

		log.Printf("Closing file at '%v'\n", fileName)
		stats.dead.Store(true)
		close(stats.notify)

		err := stats.file.Close()

		if err != nil {
			log.Printf("Encountered an error while trying to close file: %v\n", err)
		}
	}()
}

func (stats *StatisticsCollector) IsAlive() bool {
	return !stats.dead.Load()
}

func (stats *StatisticsCollector) Close() {
	stats.quit <- true
}
