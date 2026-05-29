package chapter2

import (
	"AlgoTrading2026/indicators"
	"AlgoTrading2026/series"
	"AlgoTrading2026/signals"
)

func APOStrategy(frame series.Frame, fastPeriod int, slowPeriod int) series.SignalResult {
	apo := indicators.APO(frame.CloseColumn(), fastPeriod, slowPeriod)
	directions := signals.MergeDirections(
		signals.CrossAboveValue(apo, 0),
		signals.CrossBelowValue(apo, 0),
	)
	return resultFromDirections(frame, directions)
}
