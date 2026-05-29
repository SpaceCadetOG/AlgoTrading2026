package series

import "testing"

func TestPositionChanges(t *testing.T) {
	got := PositionChanges([]float64{0, 0, 1, 1, 0})
	want := []float64{0, 0, 1, 0, -1}

	assertFloatSlices(t, got, want)
}
