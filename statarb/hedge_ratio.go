package statarb

func HedgeRatio(x []float64, y []float64) float64 {
	n := min(len(x), len(y))
	if n == 0 {
		return 0
	}

	meanX := 0.0
	meanY := 0.0
	for i := 0; i < n; i++ {
		meanX += x[i]
		meanY += y[i]
	}
	meanX /= float64(n)
	meanY /= float64(n)

	cov := 0.0
	varianceY := 0.0
	for i := 0; i < n; i++ {
		dx := x[i] - meanX
		dy := y[i] - meanY
		cov += dx * dy
		varianceY += dy * dy
	}
	if varianceY == 0 {
		return 0
	}
	return cov / varianceY
}
