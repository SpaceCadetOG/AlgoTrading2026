package chapter2

import (
	"AlgoTrading2026/indicators"
	"AlgoTrading2026/series"
	"AlgoTrading2026/signals"
)

func SMAStrategy(frame series.Frame, period int) series.SignalResult {
	close := frame.CloseColumn()
	sma := indicators.SMA(close, period)
	directions := signals.MergeDirections(
		signals.CrossAbove(close, sma),
		signals.CrossBelow(close, sma),
	)
	return resultFromDirections(frame, directions)
}
