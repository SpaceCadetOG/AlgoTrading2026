package riskmetrics

import "math"

func Variance(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}

	mean := 0.0
	for _, value := range values {
		mean += value
	}
	mean /= float64(len(values))

	variance := 0.0
	for _, value := range values {
		delta := value - mean
		variance += delta * delta
	}
	return variance / float64(len(values))
}

func StdDev(values []float64) float64 {
	return math.Sqrt(Variance(values))
}
