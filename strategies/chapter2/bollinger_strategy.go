package chapter2

import (
	"AlgoTrading2026/indicators"
	"AlgoTrading2026/series"
	"AlgoTrading2026/signals"
)

// BollingerStrategy is a simple mean-reversion interpretation:
// buy below the lower band, sell on recovery above the middle band or extension above the upper band.
func BollingerStrategy(frame series.Frame, period int, stdevFactor float64) series.SignalResult {
	close := frame.CloseColumn()
	bands := indicators.BollingerBands(close, period, stdevFactor)
	directions := make([]signals.Direction, len(close))
	for i := 1; i < len(close); i++ {
		switch {
		case close[i-1] >= bands.Lower[i-1] && close[i] < bands.Lower[i]:
			directions[i] = signals.Buy
		case close[i-1] <= bands.Middle[i-1] && close[i] > bands.Middle[i]:
			directions[i] = signals.Sell
		case close[i-1] <= bands.Upper[i-1] && close[i] > bands.Upper[i]:
			directions[i] = signals.Sell
		}
	}
	return resultFromDirections(frame, directions)
}
