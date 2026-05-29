package chapter2

import (
	"AlgoTrading2026/indicators"
	"AlgoTrading2026/series"
	"AlgoTrading2026/signals"
)

// SupportResistanceStrategy uses a simple breakout interpretation:
// buy on a close crossing above rolling resistance and sell on a close crossing below rolling support.
func SupportResistanceStrategy(frame series.Frame, period int) series.SignalResult {
	close := frame.CloseColumn()
	sr := indicators.RollingSupportResistance(frame.HighColumn(), frame.LowColumn(), period)
	priorSupport := make([]float64, len(sr.Support))
	priorResistance := make([]float64, len(sr.Resistance))
	for i := 1; i < len(close); i++ {
		priorSupport[i] = sr.Support[i-1]
		priorResistance[i] = sr.Resistance[i-1]
	}
	directions := signals.MergeDirections(
		signals.CrossAbove(close, priorResistance),
		signals.CrossBelow(close, priorSupport),
	)
	return resultFromDirections(frame, directions)
}
