package chapter2

import (
	"AlgoTrading2026/indicators"
	"AlgoTrading2026/series"
	"AlgoTrading2026/signals"
)

func MACDStrategy(frame series.Frame, fastPeriod int, slowPeriod int, signalPeriod int) series.SignalResult {
	macd := indicators.MACD(frame.CloseColumn(), fastPeriod, slowPeriod, signalPeriod)
	directions := signals.MergeDirections(
		signals.CrossAbove(macd.MACD, macd.Signal),
		signals.CrossBelow(macd.MACD, macd.Signal),
	)
	return resultFromDirections(frame, directions)
}
