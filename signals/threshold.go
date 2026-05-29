package signals

func AboveThreshold(values []float64, threshold float64) []Direction {
	out := make([]Direction, len(values))
	for i, value := range values {
		if value > threshold {
			out[i] = Buy
		}
	}
	return out
}

func BelowThreshold(values []float64, threshold float64) []Direction {
	out := make([]Direction, len(values))
	for i, value := range values {
		if value < threshold {
			out[i] = Sell
		}
	}
	return out
}
