package riskmetrics

import "testing"

func TestSharpe(t *testing.T) {
	got := Sharpe([]float64{1, 2, 3})
	if got <= 0 {
		t.Fatalf("sharpe = %f, want positive", got)
	}
}

func TestSharpeZeroStdDev(t *testing.T) {
	if got := Sharpe([]float64{1, 1, 1}); got != 0 {
		t.Fatalf("sharpe = %f, want 0", got)
	}
}
