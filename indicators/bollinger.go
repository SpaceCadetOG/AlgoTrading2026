package indicators

type BollingerResult struct {
	Middle []float64
	Upper  []float64
	Lower  []float64
}

func BollingerBands(values []float64, period int, stdevFactor float64) BollingerResult {
	middle := SMA(values, period)
	stddev := StdDev(values, period)
	upper := make([]float64, len(values))
	lower := make([]float64, len(values))

	for i := range values {
		upper[i] = middle[i] + stdevFactor*stddev[i]
		lower[i] = middle[i] - stdevFactor*stddev[i]
	}

	return BollingerResult{Middle: middle, Upper: upper, Lower: lower}
}
