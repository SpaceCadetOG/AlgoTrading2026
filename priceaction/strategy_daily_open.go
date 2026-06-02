package priceaction

import "AlgoTrading2026/exchanges"

func DailyOpenSetups(candles []exchanges.Candle) []StrategySetup {
	return openLevelSetups(candles, "daily_open", false)
}
