package signals

func BandsSignal(close []float64, upper []float64, lower []float64) []Direction {
	n := min(len(close), min(len(upper), len(lower)))
	out := make([]Direction, n)
	for i := 0; i < n; i++ {
		switch {
		case close[i] < lower[i]:
			out[i] = Buy
		case close[i] > upper[i]:
			out[i] = Sell
		}
	}
	return out
}
