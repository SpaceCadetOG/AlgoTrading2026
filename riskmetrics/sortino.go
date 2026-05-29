package riskmetrics

import "math"

func Sortino(returns []float64) float64 {
	if len(returns) == 0 {
		return 0
	}

	mean := 0.0
	downsideSquares := 0.0
	for _, value := range returns {
		mean += value
		if value < 0 {
			downsideSquares += value * value
		}
	}
	mean /= float64(len(returns))

	downsideDeviation := math.Sqrt(downsideSquares / float64(len(returns)))
	if downsideDeviation == 0 {
		return 0
	}
	return mean / downsideDeviation
}
