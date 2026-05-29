package statarb

import "math"

func RollingZScore(values []float64, period int) []float64 {
	out := make([]float64, len(values))
	if period <= 0 {
		return out
	}

	for i := range values {
		if i < period-1 {
			continue
		}
		start := i - period + 1
		mean := 0.0
		for j := start; j <= i; j++ {
			mean += values[j]
		}
		mean /= float64(period)

		variance := 0.0
		for j := start; j <= i; j++ {
			delta := values[j] - mean
			variance += delta * delta
		}
		stddev := math.Sqrt(variance / float64(period))
		if stddev == 0 {
			continue
		}
		out[i] = (values[i] - mean) / stddev
	}
	return out
}
