package pairs

func Spread(a []float64, b []float64) []float64 {
	n := min(len(a), len(b))
	out := make([]float64, n)
	for i := 0; i < n; i++ {
		out[i] = a[i] - b[i]
	}
	return out
}
