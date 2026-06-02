package priceaction

import "AlgoTrading2026/exchanges"

func SessionOpenSetups(candles []exchanges.Candle) []StrategySetup {
	return openLevelSetups(candles, "session_open", true)
}

func openLevelSetups(candles []exchanges.Candle, strategy string, session bool) []StrategySetup {
	opens := Opens(candles)
	out := make([]StrategySetup, 0)
	var currentOpen float64
	var movedAbove bool
	var movedBelow bool
	for i, candle := range candles {
		if i >= len(opens) {
			continue
		}
		level := opens[i].DailyOpenLevel
		if session {
			level = opens[i].SessionOpenLevel
		}
		if level == 0 {
			continue
		}
		if i == 0 || level != currentOpen {
			currentOpen = level
			movedAbove = false
			movedBelow = false
		}
		if candle.CloseFloat() > level {
			movedAbove = true
		}
		if candle.CloseFloat() < level {
			movedBelow = true
		}
		if i == 0 || !touchesLevel(candle, level) {
			continue
		}
		if movedAbove && candle.CloseFloat() >= level {
			out = append(out, setupAt(candles, i, strategy, StudyLong, level, level, true, false, "open level retested as support context"))
		} else if movedBelow && candle.CloseFloat() <= level {
			out = append(out, setupAt(candles, i, strategy, StudyShort, level, level, true, false, "open level retested as resistance context"))
		}
	}
	return out
}
