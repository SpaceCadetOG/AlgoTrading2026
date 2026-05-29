package series

func Diff(values []float64) []float64 {
	if len(values) == 0 {
		return nil
	}

	diff := make([]float64, len(values))
	for i := 1; i < len(values); i++ {
		diff[i] = values[i] - values[i-1]
	}
	return diff
}
