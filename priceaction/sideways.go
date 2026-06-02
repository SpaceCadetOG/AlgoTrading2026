package priceaction

import "AlgoTrading2026/exchanges"

type SidewaysFeature struct {
	Detected bool
	High     float64
	Low      float64
	Duration int
}

func Sideways(candles []exchanges.Candle, window int) []SidewaysFeature {
	if window <= 1 {
		window = 8
	}
	out := make([]SidewaysFeature, len(candles))
	for i := range candles {
		if i+1 < window {
			continue
		}
		start := i + 1 - window
		high, low := rangeHighLow(candles, start, i)
		width := high - low
		mid := (high + low) / 2
		netMove := abs(candles[i].CloseFloat() - candles[start].OpenFloat())
		overlaps := overlappingCandles(candles, start, i)
		compressed := mid > 0 && width/mid <= 0.008
		lowDirection := width > 0 && netMove <= width*0.40
		overlapping := overlaps >= window-2
		if compressed && lowDirection && overlapping {
			out[i] = SidewaysFeature{Detected: true, High: high, Low: low, Duration: window}
		}
	}
	return out
}

func rangeHighLow(candles []exchanges.Candle, start int, end int) (float64, float64) {
	high := candles[start].HighFloat()
	low := candles[start].LowFloat()
	for i := start + 1; i <= end; i++ {
		if candles[i].HighFloat() > high {
			high = candles[i].HighFloat()
		}
		if candles[i].LowFloat() < low {
			low = candles[i].LowFloat()
		}
	}
	return high, low
}

func overlappingCandles(candles []exchanges.Candle, start int, end int) int {
	var count int
	for i := start + 1; i <= end; i++ {
		prev := candles[i-1]
		current := candles[i]
		if current.LowFloat() <= prev.HighFloat() && current.HighFloat() >= prev.LowFloat() {
			count++
		}
	}
	return count
}

func abs(value float64) float64 {
	if value < 0 {
		return -value
	}
	return value
}
