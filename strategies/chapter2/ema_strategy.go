package chapter2

import (
	"AlgoTrading2026/indicators"
	"AlgoTrading2026/series"
	"AlgoTrading2026/signals"
)

func EMAStrategy(frame series.Frame, period int) series.SignalResult {
	close := frame.CloseColumn()
	ema := indicators.EMA(close, period)
	directions := signals.MergeDirections(
		signals.CrossAbove(close, ema),
		signals.CrossBelow(close, ema),
	)
	return resultFromDirections(frame, directions)
}
