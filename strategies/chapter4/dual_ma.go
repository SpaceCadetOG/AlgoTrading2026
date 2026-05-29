package chapter4

import (
	"AlgoTrading2026/indicators"
	"AlgoTrading2026/series"
)

func DualMAStrategy(frame series.Frame, fastPeriod int, slowPeriod int) series.SignalResult {
	close := frame.CloseColumn()
	fast := indicators.SMA(close, fastPeriod)
	slow := indicators.SMA(close, slowPeriod)
	positions := make([]float64, len(close))

	long := false
	for i := 1; i < len(close); i++ {
		if fast[i] == 0 || slow[i] == 0 {
			continue
		}
		switch {
		case fast[i-1] <= slow[i-1] && fast[i] > slow[i] && !long:
			positions[i] = 1
			long = true
		case fast[i-1] >= slow[i-1] && fast[i] < slow[i] && long:
			positions[i] = -1
			long = false
		}
	}

	return resultFromPositions(frame, positions)
}
