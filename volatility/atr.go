package volatility

import "math"

func TrueRange(high []float64, low []float64, close []float64) []float64 {
	n := min(len(high), min(len(low), len(close)))
	out := make([]float64, n)
	for i := 0; i < n; i++ {
		rangeValue := high[i] - low[i]
		if i == 0 {
			out[i] = rangeValue
			continue
		}
		highGap := math.Abs(high[i] - close[i-1])
		lowGap := math.Abs(low[i] - close[i-1])
		out[i] = math.Max(rangeValue, math.Max(highGap, lowGap))
	}
	return out
}

func ATR(high []float64, low []float64, close []float64, period int) []float64 {
	return rollingMean(TrueRange(high, low, close), period)
}
