package chapter2

import (
	"AlgoTrading2026/series"
	"AlgoTrading2026/signals"
)

type Config struct {
	Period       int
	FastPeriod   int
	SlowPeriod   int
	SignalPeriod int
	StdDevFactor float64
	RSILow       float64
	RSIHigh      float64
}

func DefaultConfig() Config {
	return Config{
		Period:       20,
		FastPeriod:   12,
		SlowPeriod:   26,
		SignalPeriod: 9,
		StdDevFactor: 2,
		RSILow:       30,
		RSIHigh:      70,
	}
}

func resultFromDirections(frame series.Frame, directions []signals.Direction) series.SignalResult {
	close := frame.CloseColumn()
	times := frame.TimeColumn()
	n := min(len(close), min(len(times), len(directions)))
	close = close[:n]
	times = times[:n]
	directions = directions[:n]

	signal := make([]float64, n)
	positions := make([]float64, n)
	state := 0.0
	for i, direction := range directions {
		next := state
		switch direction {
		case signals.Buy:
			next = 1
		case signals.Sell:
			next = 0
		}
		signal[i] = next
		positions[i] = next - state
		state = next
	}

	return series.SignalResult{
		Times:     times,
		Close:     close,
		Diff:      series.Diff(close),
		Signal:    signal,
		Positions: positions,
	}
}
