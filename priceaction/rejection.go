package priceaction

import "AlgoTrading2026/exchanges"

type RejectionDirection string

const (
	RejectionNone    RejectionDirection = "none"
	RejectionBullish RejectionDirection = "bullish"
	RejectionBearish RejectionDirection = "bearish"
)

type RejectionFeature struct {
	Detected  bool
	Direction RejectionDirection
	Level     float64
}

func Rejections(candles []exchanges.Candle) []RejectionFeature {
	out := make([]RejectionFeature, len(candles))
	for i, candle := range candles {
		metrics := Metrics(candle)
		if candle.Range() <= 0 {
			continue
		}
		close := candle.CloseFloat()
		high := candle.HighFloat()
		low := candle.LowFloat()
		switch {
		case metrics.LowerWickPct >= 0.50 && close >= low+candle.Range()*0.60:
			out[i] = RejectionFeature{Detected: true, Direction: RejectionBullish, Level: low}
		case metrics.UpperWickPct >= 0.50 && close <= high-candle.Range()*0.60:
			out[i] = RejectionFeature{Detected: true, Direction: RejectionBearish, Level: high}
		}
	}
	return out
}
