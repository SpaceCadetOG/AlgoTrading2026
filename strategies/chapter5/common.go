package chapter5

import (
	"AlgoTrading2026/series"
	"AlgoTrading2026/volatility"
)

func resultFromPositions(frame series.Frame, positions []float64) series.SignalResult {
	close := frame.CloseColumn()
	times := frame.TimeColumn()
	n := min(len(close), min(len(times), len(positions)))
	close = close[:n]
	times = times[:n]
	positions = positions[:n]

	signal := make([]float64, n)
	state := 0.0
	for i, position := range positions {
		if position > 0 {
			state = 1
		} else if position < 0 {
			state = 0
		}
		signal[i] = state
	}

	return series.SignalResult{
		Times:     times,
		Close:     close,
		Diff:      series.Diff(close),
		Signal:    signal,
		Positions: positions,
	}
}

func regimeForFrame(frame series.Frame) []volatility.Regime {
	close := frame.CloseColumn()
	return volatility.DefaultRegime(volatility.RealizedVolatility(close, 20))
}

func CountRegimes(regimes []volatility.Regime) (int, int, int) {
	low := 0
	normal := 0
	high := 0
	for _, regime := range regimes {
		switch regime {
		case volatility.RegimeLow:
			low++
		case volatility.RegimeHigh:
			high++
		default:
			normal++
		}
	}
	return low, normal, high
}
