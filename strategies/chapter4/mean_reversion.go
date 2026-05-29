package chapter4

import (
	"AlgoTrading2026/indicators"
	"AlgoTrading2026/series"
)

func MeanReversionStrategy(
	frame series.Frame,
	bollingerPeriod int,
	stdevFactor float64,
	rsiPeriod int,
	entryRSI float64,
	exitRSI float64,
) series.SignalResult {
	close := frame.CloseColumn()
	bands := indicators.BollingerBands(close, bollingerPeriod, stdevFactor)
	rsi := indicators.RSI(close, rsiPeriod)
	positions := make([]float64, len(close))

	long := false
	warmup := max(bollingerPeriod, rsiPeriod)
	for i := 0; i < len(close); i++ {
		if i < warmup {
			continue
		}
		switch {
		case (close[i] < bands.Lower[i] || rsi[i] < entryRSI) && !long:
			positions[i] = 1
			long = true
		case (close[i] >= bands.Middle[i] || rsi[i] > exitRSI) && long:
			positions[i] = -1
			long = false
		}
	}

	return resultFromPositions(frame, positions)
}
