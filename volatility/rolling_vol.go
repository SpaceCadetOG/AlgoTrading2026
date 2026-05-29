package volatility

import "math"

func RollingStdDev(values []float64, period int) []float64 {
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
		out[i] = math.Sqrt(variance / float64(period))
	}
	return out
}

func rollingMean(values []float64, period int) []float64 {
	out := make([]float64, len(values))
	if period <= 0 {
		return out
	}

	sum := 0.0
	for i, value := range values {
		sum += value
		if i >= period {
			sum -= values[i-period]
		}
		if i >= period-1 {
			out[i] = sum / float64(period)
		}
	}
	return out
}
