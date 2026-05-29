package backtest

import "AlgoTrading2026/strategy"

type Metrics struct {
	TotalTrades    int
	Wins           int
	Losses         int
	WinRate        float64
	GrossPnL       float64
	Fees           float64
	NetPnL         float64
	PeakEquity     float64
	MaxDrawdown    float64
	MaxDrawdownPct float64
	AveragePnL     float64
}

func CalculateMetrics(trades []strategy.Trade) Metrics {
	metrics := Metrics{TotalTrades: len(trades)}
	if len(trades) == 0 {
		return metrics
	}

	for _, trade := range trades {
		if trade.RealizedPnL > 0 {
			metrics.Wins++
		} else if trade.RealizedPnL < 0 {
			metrics.Losses++
		}
		metrics.GrossPnL += trade.GrossPnL
		metrics.Fees += trade.Fees
		metrics.NetPnL += trade.RealizedPnL
	}

	metrics.WinRate = float64(metrics.Wins) / float64(metrics.TotalTrades) * 100
	metrics.AveragePnL = metrics.NetPnL / float64(metrics.TotalTrades)

	return metrics
}
