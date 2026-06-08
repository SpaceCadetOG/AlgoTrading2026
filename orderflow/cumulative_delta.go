package orderflow

func CumulativeDelta(bars []FootprintBar) []float64 {
	out := make([]float64, len(bars))
	var running float64
	for i, bar := range bars {
		running += Delta(bar)
		out[i] = running
	}
	return out
}
