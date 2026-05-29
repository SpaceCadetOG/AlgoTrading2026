package riskmetrics

import "testing"

func TestCalculateExecutionStats(t *testing.T) {
	got := CalculateExecutionStats(30, 10)
	if got.TradesPerDay != 3 {
		t.Fatalf("trades/day = %f, want 3", got.TradesPerDay)
	}
	if got.TradesPerWeek != 21 {
		t.Fatalf("trades/week = %f, want 21", got.TradesPerWeek)
	}
	if got.TradesPerMonth != 90 {
		t.Fatalf("trades/month = %f, want 90", got.TradesPerMonth)
	}
}
