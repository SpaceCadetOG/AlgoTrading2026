package volatility

import "math"

func NormalizeByVolatility(values []float64, vol []float64) []float64 {
	n := min(len(values), len(vol))
	out := make([]float64, n)
	for i := 0; i < n; i++ {
		if vol[i] == 0 || math.IsNaN(vol[i]) || math.IsInf(vol[i], 0) {
			continue
		}
		out[i] = values[i] / vol[i]
	}
	return out
}
