package series

import "testing"

func TestDiff(t *testing.T) {
	got := Diff([]float64{10, 9, 11, 12, 10})
	want := []float64{0, -1, 2, 1, -2}

	assertFloatSlices(t, got, want)
}
