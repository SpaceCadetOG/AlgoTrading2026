package chapter4

import (
	"AlgoTrading2026/indicators"
	"AlgoTrading2026/series"
)

func TurtleBreakoutStrategy(frame series.Frame, entryPeriod int, exitPeriod int) series.SignalResult {
	close := frame.CloseColumn()
	entry := indicators.RollingSupportResistance(frame.HighColumn(), frame.LowColumn(), entryPeriod)
	exit := indicators.RollingSupportResistance(frame.HighColumn(), frame.LowColumn(), exitPeriod)
	positions := make([]float64, len(close))

	long := false
	for i := 1; i < len(close); i++ {
		priorEntryHigh := entry.Resistance[i-1]
		priorExitLow := exit.Support[i-1]
		switch {
		case close[i] > priorEntryHigh && !long:
			positions[i] = 1
			long = true
		case close[i] < priorExitLow && long:
			positions[i] = -1
			long = false
		}
	}

	return resultFromPositions(frame, positions)
}
