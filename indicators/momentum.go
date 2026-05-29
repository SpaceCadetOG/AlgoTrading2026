package indicators

func Momentum(values []float64, period int) []float64 {
	out := make([]float64, len(values))
	if period <= 0 {
		return out
	}

	for i := period; i < len(values); i++ {
		out[i] = values[i] - values[i-period]
	}
	return out
}
