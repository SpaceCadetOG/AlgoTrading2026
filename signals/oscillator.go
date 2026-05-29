package signals

func CrossAboveValue(values []float64, threshold float64) []Direction {
	level := constantLine(len(values), threshold)
	return CrossAbove(values, level)
}

func CrossBelowValue(values []float64, threshold float64) []Direction {
	level := constantLine(len(values), threshold)
	return CrossBelow(values, level)
}

func MergeDirections(parts ...[]Direction) []Direction {
	n := 0
	for _, part := range parts {
		if len(part) > n {
			n = len(part)
		}
	}

	out := make([]Direction, n)
	for _, part := range parts {
		for i, value := range part {
			if value != Hold {
				out[i] = value
			}
		}
	}
	return out
}

func constantLine(n int, value float64) []float64 {
	out := make([]float64, n)
	for i := range out {
		out[i] = value
	}
	return out
}
