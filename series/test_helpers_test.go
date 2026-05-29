package series

import "testing"

func assertFloatSlices(t *testing.T, got []float64, want []float64) {
	t.Helper()

	if len(got) != len(want) {
		t.Fatalf("len = %d, want %d; got=%v want=%v", len(got), len(want), got, want)
	}

	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("value[%d] = %f, want %f; got=%v want=%v", i, got[i], want[i], got, want)
		}
	}
}
