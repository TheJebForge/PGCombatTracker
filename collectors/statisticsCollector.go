package collectors

import (
	"PGCombatTracker/abstract"
	"PGCombatTracker/parser"
	"bufio"
	"io"
	"log"
	"os"
	"sync"
	"sync/atomic"
	"time"
)

type StatisticsCollector struct {
	username   string
	dead       *atomic.Bool
	collectors []abstract.Collector
	quit       chan bool
	watch      bool
	file       *os.File
	reader     *bufio.Reader
	lock       *sync.RWMutex
	locked     bool
	notify     chan bool
}

func NewStatisticsCollector(path string, watchFile bool) (*StatisticsCollector, error) {
	file, err := os.Open(path)

	if err != nil {
		return nil, err
	}

	return &StatisticsCollector{
		collectors: []abstract.Collector{
			NewDamageDealtCollector(),
			NewDamageTakenCollector(),
		},
		dead:   &atomic.Bool{},
		quit:   make(chan bool, 1),
		watch:  watchFile,
		file:   file,
		reader: bufio.NewReader(file),
		lock:   new(sync.RWMutex),
		notify: make(chan bool),
	}, nil
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

func (stats *StatisticsCollector) Run() {
	fileName := stats.file.Name()

	log.Printf("Starting to read file at '%v'\n", fileName)

	go func() {
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

				event, err := parser.ParseLine(line)

				if err != nil {
					log.Printf("Encountered an error while parsing line: %v\n", err)
					continue infinite
				}

				if event == nil {
					continue infinite
				}

				if login, ok := event.Contents.(*abstract.Login); ok && login != nil {
					log.Printf("Detected login as %v\n", login.Name)
					stats.username = login.Name
					continue infinite
				}

				log.Println(event)

				stats.lockTheLock()

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
