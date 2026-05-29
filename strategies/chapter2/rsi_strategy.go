package chapter2

import (
	"AlgoTrading2026/indicators"
	"AlgoTrading2026/series"
	"AlgoTrading2026/signals"
)

func RSIStrategy(frame series.Frame, period int, low float64, high float64) series.SignalResult {
	rsi := indicators.RSI(frame.CloseColumn(), period)
	directions := make([]signals.Direction, len(rsi))
	oversoldSeen := false

	for i := 1; i < len(rsi); i++ {
		if rsi[i] < low {
			oversoldSeen = true
		}
		if oversoldSeen && rsi[i-1] <= low && rsi[i] > low {
			directions[i] = signals.Buy
			oversoldSeen = false
			continue
		}
		if rsi[i-1] <= high && rsi[i] > high {
			directions[i] = signals.Sell
			continue
		}
		if rsi[i-1] >= high && rsi[i] < high {
			directions[i] = signals.Sell
		}
	}
	return resultFromDirections(frame, directions)
}
