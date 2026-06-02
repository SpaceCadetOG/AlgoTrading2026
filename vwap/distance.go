package vwap

func Distance(price []float64, baseline []float64) []float64 {
	n := minLen(price, baseline)
	out := make([]float64, n)
	for i := 0; i < n; i++ {
		out[i] = price[i] - baseline[i]
	}
	return out
}

func DistancePct(price []float64, baseline []float64) []float64 {
	n := minLen(price, baseline)
	out := make([]float64, n)
	for i := 0; i < n; i++ {
		if baseline[i] != 0 {
			out[i] = (price[i] - baseline[i]) / baseline[i]
		}
	}
	return out
}

func minLen(a []float64, b []float64) int {
	if len(a) < len(b) {
		return len(a)
	}
	return len(b)
}
