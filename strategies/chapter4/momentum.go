package chapter4

import (
	"AlgoTrading2026/indicators"
	"AlgoTrading2026/series"
)

func MomentumStrategy(frame series.Frame, period int) series.SignalResult {
	close := frame.CloseColumn()
	momentum := indicators.Momentum(close, period)
	positions := make([]float64, len(close))

	long := false
	for i, value := range momentum {
		switch {
		case value > 0 && !long:
			positions[i] = 1
			long = true
		case value < 0 && long:
			positions[i] = -1
			long = false
		}
	}

	return resultFromPositions(frame, positions)
}

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
