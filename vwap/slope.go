package vwap

func Slope(values []float64, period int) []float64 {
	out := make([]float64, len(values))
	if period <= 0 {
		period = 1
	}
	for i := range values {
		if i < period {
			continue
		}
		out[i] = (values[i] - values[i-period]) / float64(period)
	}
	return out
}
