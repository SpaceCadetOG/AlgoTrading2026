package priceaction

import "AlgoTrading2026/exchanges"

type FlipDirection string

const (
	FlipNone                FlipDirection = "none"
	FlipSupportToResistance FlipDirection = "support_to_resistance"
	FlipResistanceToSupport FlipDirection = "resistance_to_support"
)

type FlipFeature struct {
	Detected  bool
	Level     float64
	Direction FlipDirection
}

func SupportResistanceFlips(candles []exchanges.Candle, rejections []RejectionFeature) []FlipFeature {
	out := make([]FlipFeature, len(candles))
	for i, candle := range candles {
		if i == 0 {
			continue
		}
		for j := i - 1; j >= 0 && j >= i-30; j-- {
			if j >= len(rejections) || !rejections[j].Detected {
				continue
			}
			level := rejections[j].Level
			if level == 0 {
				continue
			}
			tolerance := level * 0.001
			returned := candle.LowFloat() <= level+tolerance && candle.HighFloat() >= level-tolerance
			if !returned {
				continue
			}
			switch rejections[j].Direction {
			case RejectionBullish:
				breached := priorCloseBelow(candles, j+1, i, level)
				if breached && candle.CloseFloat() < level {
					out[i] = FlipFeature{Detected: true, Level: level, Direction: FlipSupportToResistance}
					j = -1
				}
			case RejectionBearish:
				breached := priorCloseAbove(candles, j+1, i, level)
				if breached && candle.CloseFloat() > level {
					out[i] = FlipFeature{Detected: true, Level: level, Direction: FlipResistanceToSupport}
					j = -1
				}
			}
		}
	}
	return out
}

func priorCloseBelow(candles []exchanges.Candle, start int, end int, level float64) bool {
	for i := start; i < end && i < len(candles); i++ {
		if candles[i].CloseFloat() < level {
			return true
		}
	}
	return false
}

func priorCloseAbove(candles []exchanges.Candle, start int, end int, level float64) bool {
	for i := start; i < end && i < len(candles); i++ {
		if candles[i].CloseFloat() > level {
			return true
		}
	}
	return false
}
