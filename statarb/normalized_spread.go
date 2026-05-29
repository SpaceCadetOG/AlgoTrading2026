package statarb

func NormalizedSpread(x []float64, y []float64, beta float64) []float64 {
	n := min(len(x), len(y))
	out := make([]float64, n)
	for i := 0; i < n; i++ {
		out[i] = x[i] - beta*y[i]
	}
	return out
}
