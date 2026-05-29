package signals

func CrossAbove(a []float64, b []float64) []Direction {
	n := min(len(a), len(b))
	out := make([]Direction, n)
	for i := 1; i < n; i++ {
		if a[i-1] <= b[i-1] && a[i] > b[i] {
			out[i] = Buy
		}
	}
	return out
}

func CrossBelow(a []float64, b []float64) []Direction {
	n := min(len(a), len(b))
	out := make([]Direction, n)
	for i := 1; i < n; i++ {
		if a[i-1] >= b[i-1] && a[i] < b[i] {
			out[i] = Sell
		}
	}
	return out
}
