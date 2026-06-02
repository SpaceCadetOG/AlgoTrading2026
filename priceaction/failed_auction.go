package priceaction

import "AlgoTrading2026/exchanges"

type FailedAuctionDirection string

const (
	FailedAuctionNone FailedAuctionDirection = "none"
	FailedAuctionUp   FailedAuctionDirection = "up"
	FailedAuctionDown FailedAuctionDirection = "down"
)

type FailedAuctionFeature struct {
	Detected  bool
	Direction FailedAuctionDirection
}

func FailedAuctions(candles []exchanges.Candle, levels []HighLowFeature) []FailedAuctionFeature {
	out := make([]FailedAuctionFeature, len(candles))
	for i, candle := range candles {
		if i >= len(levels) {
			continue
		}
		level := levels[i]
		if level.PreviousDailyHigh > 0 && candle.HighFloat() > level.PreviousDailyHigh && candle.CloseFloat() < level.PreviousDailyHigh {
			out[i] = FailedAuctionFeature{Detected: true, Direction: FailedAuctionUp}
			continue
		}
		if level.PreviousDailyLow > 0 && candle.LowFloat() < level.PreviousDailyLow && candle.CloseFloat() > level.PreviousDailyLow {
			out[i] = FailedAuctionFeature{Detected: true, Direction: FailedAuctionDown}
		}
	}
	return out
}
