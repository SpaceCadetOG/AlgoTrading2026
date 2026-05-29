package indicators

func APO(values []float64, fastPeriod int, slowPeriod int) []float64 {
	fast := EMA(values, fastPeriod)
	slow := EMA(values, slowPeriod)
	out := make([]float64, len(values))
	for i := range values {
		out[i] = fast[i] - slow[i]
	}
	return out
}
