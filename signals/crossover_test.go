package signals

import "testing"

func TestCrossAboveCrossBelow(t *testing.T) {
	above := CrossAbove([]float64{1, 3}, []float64{2, 2})
	if above[1] != Buy {
		t.Fatalf("cross above = %v, want buy at 1", above)
	}

	below := CrossBelow([]float64{3, 1}, []float64{2, 2})
	if below[1] != Sell {
		t.Fatalf("cross below = %v, want sell at 1", below)
	}
}

