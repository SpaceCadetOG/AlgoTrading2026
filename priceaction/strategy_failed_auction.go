package priceaction

import "AlgoTrading2026/exchanges"

func FailedAuctionSetups(candles []exchanges.Candle) []Phase3Setup {
	rejections := Rejections(candles)
	aggression := Aggression(candles, 20)
	levels := HighsLows(candles, rejections, aggression)
	failed := FailedAuctions(candles, levels)

	out := make([]Phase3Setup, 0)
	for i, feature := range failed {
		if !feature.Detected || i >= len(candles) || i >= len(levels) {
			continue
		}
		switch feature.Direction {
		case FailedAuctionUp:
			level := levels[i].PreviousDailyHigh
			out = append(out, phase3SetupAt(candles, i, "failed_high_auction", Phase3ShortContext, level, "prior_high", true, phase3Invalidated(candles[i], Phase3ShortContext, level), "breached prior high and closed back inside prior range"))
		case FailedAuctionDown:
			level := levels[i].PreviousDailyLow
			out = append(out, phase3SetupAt(candles, i, "failed_low_auction", Phase3LongContext, level, "prior_low", true, phase3Invalidated(candles[i], Phase3LongContext, level), "breached prior low and closed back inside prior range"))
		}
	}
	return out
}
