package series

func PositionChanges(signal []float64) []float64 {
	return Diff(signal)
}
