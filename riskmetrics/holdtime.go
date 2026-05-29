package riskmetrics

import "AlgoTrading2026/strategy"

type HoldStats struct {
	AverageHoldCandles float64
	MaximumHoldCandles int
	MinimumHoldCandles int

	AverageHoldHours float64
}

func CalculateHoldStats(trades []strategy.Trade) HoldStats {
	if len(trades) == 0 {
		return HoldStats{}
	}

	total := 0
	minHold := 0
	maxHold := 0
	for i, trade := range trades {
		hold := int(trade.ClosedAt.Sub(trade.OpenedAt).Minutes() / 15)
		if hold < 1 {
			hold = 1
		}
		total += hold
		if i == 0 || hold < minHold {
			minHold = hold
		}
		if hold > maxHold {
			maxHold = hold
		}
	}

	avg := float64(total) / float64(len(trades))
	return HoldStats{
		AverageHoldCandles: avg,
		MaximumHoldCandles: maxHold,
		MinimumHoldCandles: minHold,
		AverageHoldHours:   avg * 0.25,
	}
}
