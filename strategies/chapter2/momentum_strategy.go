package chapter2

import (
	"AlgoTrading2026/indicators"
	"AlgoTrading2026/series"
	"AlgoTrading2026/signals"
)

func MomentumStrategy(frame series.Frame, period int) series.SignalResult {
	momentum := indicators.Momentum(frame.CloseColumn(), period)
	directions := signals.MergeDirections(
		signals.CrossAboveValue(momentum, 0),
		signals.CrossBelowValue(momentum, 0),
	)
	return resultFromDirections(frame, directions)
}
