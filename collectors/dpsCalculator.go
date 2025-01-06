package collectors

import (
	"PGCombatTracker/abstract"
	"PGCombatTracker/ui/components"
	"PGCombatTracker/utils"
	"fmt"
	"math"
	"time"
)

func NewDPSCalculator(chart *components.TimeBasedChart, settings *abstract.Settings) *DPSCalculator {
	return &DPSCalculator{
		Chart:             chart,
		SecondsUntilReset: settings.SecondsUntilDPSReset,
	}
}

type DPSCalculator struct {
	Chart             *components.TimeBasedChart
	SecondsUntilReset int

	startTime   time.Time
	lastTime    time.Time
	totalDamage int
}

type DPSDetail int

func (D DPSDetail) StringCL(long bool) string {
	if long {
		return fmt.Sprintf("%d DPS", int(D))
	} else {
		return fmt.Sprintf("%v DPS", utils.FormatNumber(int(D)))
	}
}

func (D DPSDetail) Interpolate(other utils.Interpolatable, t float64) utils.Interpolatable {
	otherInt, ok := other.(DPSDetail)
	if !ok {
		return other
	}

	return DPSDetail(utils.LerpInt(int(D), int(otherInt), t))
}

func (D DPSDetail) InterpolateILF(other utils.InterpolatableLongFormatable, t float64) utils.InterpolatableLongFormatable {
	return D.Interpolate(other, t).(utils.InterpolatableLongFormatable)
}

func (dps *DPSCalculator) politeAdd(at time.Time, value int) {
	dp := dps.Chart.DataPoints
	last := len(dp) - 1

	if len(dp) < 2 {
		dps.Chart.Add(components.TimePoint{
			Time:    at,
			Value:   value,
			Details: DPSDetail(value),
		})
	} else {
		if dp[last].Value == dp[last-1].Value {
			dp[last].Time = at
		} else {
			dps.Chart.Add(components.TimePoint{
				Time:    at,
				Value:   value,
				Details: DPSDetail(value),
			})
		}
	}
}

func (dps *DPSCalculator) avgDamage(at time.Time) int {
	floatingTotal := float64(dps.totalDamage)
	secondsSinceStart := max(1, at.Sub(dps.startTime).Seconds())
	averagedDamage := floatingTotal / secondsSinceStart
	return int(math.Round(averagedDamage))
}

func (dps *DPSCalculator) Add(at time.Time, damage int) {
	dps.lastTime = at
	if dps.totalDamage == 0 {
		dps.startTime = at
		dps.totalDamage = damage
	} else {
		dps.totalDamage += damage
	}

	dps.politeAdd(at, dps.avgDamage(at))
}

func (dps *DPSCalculator) Tick(at time.Time) {
	if dps.lastTime.Add(time.Second * time.Duration(dps.SecondsUntilReset)).Before(at) {
		dps.totalDamage = 0
	}

	dps.politeAdd(at, dps.avgDamage(at))
}
