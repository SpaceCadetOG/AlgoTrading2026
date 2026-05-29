package riskmetrics

import "testing"

func TestExpectancy(t *testing.T) {
	got := Expectancy(0.5, 2, 1)
	if got != 0.5 {
		t.Fatalf("expectancy = %f, want 0.5", got)
	}
}
