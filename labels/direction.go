package labels

func DirectionLabel(futureReturns []float64, threshold float64) []int {
	out := make([]int, len(futureReturns))
	for i, value := range futureReturns {
		switch {
		case value > threshold:
			out[i] = 1
		case value < -threshold:
			out[i] = -1
		}
	}
	return out
}
