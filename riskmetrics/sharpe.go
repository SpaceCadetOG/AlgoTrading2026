package riskmetrics

func Sharpe(returns []float64) float64 {
	if len(returns) == 0 {
		return 0
	}

	mean := 0.0
	for _, value := range returns {
		mean += value
	}
	mean /= float64(len(returns))

	stddev := StdDev(returns)
	if stddev == 0 {
		return 0
	}
	return mean / stddev
}
