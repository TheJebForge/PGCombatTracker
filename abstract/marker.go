package abstract

import (
	"fmt"
	"time"
)

type Marker struct {
	Time        time.Time
	Name        string
	User        string
	UserDefined bool `json:"-"`
}

func (m Marker) AsUserDefined() Marker {
	m.UserDefined = true
	return m
}

func (m Marker) EqualTo(other Marker) bool {
	return m.Time.Equal(other.Time) && m.Name == other.Name && m.User == other.User
}

func (m Marker) String() string {
	return fmt.Sprintf("%v by %v: %v", m.Time.Format(time.DateTime), m.User, m.Name)
}

func NewMarkers() *Markers {
	return &Markers{}
}

type MarkerFile struct {
	Path    string
	Markers []Marker
}

type Markers struct {
	Files []MarkerFile
}

type MarkerTimeFrame struct {
	User string
	From time.Time
	To   time.Time
}

func (t MarkerTimeFrame) String() string {
	return fmt.Sprintf("From '%v' to '%v' as '%v'", t.From.Format(time.DateTime), t.To.Format(time.DateTime), t.User)
}
func (t MarkerTimeFrame) Within(time time.Time) bool {
	return (time.After(t.From) || time.Equal(t.From)) && time.Before(t.To)
}
func WithinTimeFrames(timeFrames []MarkerTimeFrame, time time.Time) (bool, string) {
	for _, timeFrame := range timeFrames {
		if timeFrame.Within(time) {
			return true, timeFrame.User
		}
	}

	return false, ""
}
