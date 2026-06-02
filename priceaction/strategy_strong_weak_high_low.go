package priceaction

import "AlgoTrading2026/exchanges"

func StrongWeakHighLowSetups(candles []exchanges.Candle) []Phase3Setup {
	rejections := Rejections(candles)
	aggression := Aggression(candles, 20)
	levels := HighsLows(candles, rejections, aggression)

	out := make([]Phase3Setup, 0)
	for i, candle := range candles {
		if i >= len(levels) {
			break
		}
		level := levels[i]
		if level.StrongHigh {
			out = append(out, phase3SetupAt(candles, i, "strong_high", Phase3ShortContext, highLevel(candle, rejections, i), "swing_high", true, phase3Invalidated(candle, Phase3ShortContext, highLevel(candle, rejections, i)), "bearish rejection with aggressive opposite movement"))
		}
		if level.StrongLow {
			out = append(out, phase3SetupAt(candles, i, "strong_low", Phase3LongContext, lowLevel(candle, rejections, i), "swing_low", true, phase3Invalidated(candle, Phase3LongContext, lowLevel(candle, rejections, i)), "bullish rejection with aggressive opposite movement"))
		}
		if level.WeakHigh {
			out = append(out, phase3SetupAt(candles, i, "weak_high", Phase3BreakoutWatch, nearestHighLevel(candle, level), "swing_high", true, false, "prior high nearby without strong rejection; breakout watch context"))
		}
		if level.WeakLow {
			out = append(out, phase3SetupAt(candles, i, "weak_low", Phase3BreakoutWatch, nearestLowLevel(candle, level), "swing_low", true, false, "prior low nearby without strong rejection; breakout watch context"))
		}
	}
	return out
}

func highLevel(candle exchanges.Candle, rejections []RejectionFeature, i int) float64 {
	if i >= 0 && i < len(rejections) && rejections[i].Level > 0 {
		return rejections[i].Level
	}
	return candle.HighFloat()
}

func lowLevel(candle exchanges.Candle, rejections []RejectionFeature, i int) float64 {
	if i >= 0 && i < len(rejections) && rejections[i].Level > 0 {
		return rejections[i].Level
	}
	return candle.LowFloat()
}

func nearestHighLevel(candle exchanges.Candle, level HighLowFeature) float64 {
	close := candle.CloseFloat()
	best := candle.HighFloat()
	bestDistance := abs(close - best)
	for _, candidate := range []float64{level.PreviousDailyHigh, level.PreviousWeeklyHigh} {
		if candidate <= 0 {
			continue
		}
		distance := abs(close - candidate)
		if distance < bestDistance {
			best = candidate
			bestDistance = distance
		}
	}
	return best
}

func nearestLowLevel(candle exchanges.Candle, level HighLowFeature) float64 {
	close := candle.CloseFloat()
	best := candle.LowFloat()
	bestDistance := abs(close - best)
	for _, candidate := range []float64{level.PreviousDailyLow, level.PreviousWeeklyLow} {
		if candidate <= 0 {
			continue
		}
		distance := abs(close - candidate)
		if distance < bestDistance {
			best = candidate
			bestDistance = distance
		}
	}
	return best
}
