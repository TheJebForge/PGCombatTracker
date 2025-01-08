package collectors

import (
	"PGCombatTracker/abstract"
	"PGCombatTracker/ui/components"
	"PGCombatTracker/utils"
	"fmt"
	"math"
	"time"
)

func dpsZeroPoint(point components.TimePoint) components.TimePoint {
	return components.TimePoint{
		Time:    point.Time.Add(-time.Millisecond),
		Value:   0,
		Details: DPSDetail(0),
	}
}

func NewDPSCalculatorForChart(chart *components.TimeBasedChart, settings *abstract.Settings) *DPSCalculator {
	return NewDPSCalculator(
		func(point components.TimePoint) {
			dp := chart.DataPoints
			last := len(dp) - 1

			if len(dp) == 0 {
				chart.Add(dpsZeroPoint(point))
			}

			if len(dp) < 2 {
				chart.Add(point)
			} else {
				if dp[last].Value == dp[last-1].Value && dp[last].Value == point.Value && point.Value == 0 {
					dp[last].Time = point.Time
				} else {
					if dp[last].Value == 0 {
						chart.Add(dpsZeroPoint(point))
					}
					chart.Add(point)
				}
			}
		},
		settings,
	)
}

func NewDPSCalculatorForController(controller *components.TimeController, settings *abstract.Settings) *DPSCalculator {
	return NewDPSCalculator(
		func(point components.TimePoint) {
			dp := controller.BaseChart.DataPoints
			last := len(dp) - 1

			if len(dp) == 0 {
				controller.Add(dpsZeroPoint(point))
			}

			if len(dp) < 2 {
				controller.Add(point)
			} else {
				if dp[last].Value == dp[last-1].Value && dp[last].Value == point.Value && point.Value == 0 {
					dp[last].Time = point.Time
					controller.FullTimeFrame = controller.FullTimeFrame.Expand(point.Time)
					controller.RecalculateTimeFrame()
				} else {
					if dp[last].Value == 0 {
						controller.Add(dpsZeroPoint(point))
					}
					controller.Add(point)
				}
			}
		},
		settings,
	)
}

func NewDPSCalculator(pointsConsumer func(components.TimePoint), settings *abstract.Settings) *DPSCalculator {
	return &DPSCalculator{
		PointsConsumer:    pointsConsumer,
		SecondsUntilReset: settings.SecondsUntilDPSReset,
	}
}

type DPSCalculator struct {
	PointsConsumer    func(components.TimePoint)
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

func (dps *DPSCalculator) sendData(at time.Time, value int) {
	dps.PointsConsumer(components.TimePoint{
		Time:    at,
		Value:   value,
		Details: DPSDetail(value),
	})
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

	dps.sendData(at, dps.avgDamage(at))
}

func (dps *DPSCalculator) Tick(at time.Time) {
	if dps.lastTime.Add(time.Second * time.Duration(dps.SecondsUntilReset)).Before(at) {
		dps.totalDamage = 0
	}

	dps.sendData(at, dps.avgDamage(at))
}
