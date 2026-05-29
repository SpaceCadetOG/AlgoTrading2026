package riskmetrics

import "testing"

func TestSortino(t *testing.T) {
	got := Sortino([]float64{1, -1, 2})
	if got <= 0 {
		t.Fatalf("sortino = %f, want positive", got)
	}
}

func TestSortinoNoDownside(t *testing.T) {
	if got := Sortino([]float64{1, 2, 3}); got != 0 {
		t.Fatalf("sortino = %f, want 0", got)
	}
}
