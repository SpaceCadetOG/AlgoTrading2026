package riskmetrics

import (
	"testing"
	"time"

	"AlgoTrading2026/strategy"
)

func TestCalculateHoldStats(t *testing.T) {
	start := time.Unix(0, 0).UTC()
	trades := []strategy.Trade{
		{OpenedAt: start, ClosedAt: start.Add(30 * time.Minute)},
		{OpenedAt: start, ClosedAt: start.Add(60 * time.Minute)},
	}

	got := CalculateHoldStats(trades)
	if got.AverageHoldCandles != 3 {
		t.Fatalf("avg hold candles = %f, want 3", got.AverageHoldCandles)
	}
	if got.MinimumHoldCandles != 2 || got.MaximumHoldCandles != 4 {
		t.Fatalf("min/max = %d/%d, want 2/4", got.MinimumHoldCandles, got.MaximumHoldCandles)
	}
	if got.AverageHoldHours != 0.75 {
		t.Fatalf("avg hold hours = %f, want 0.75", got.AverageHoldHours)
	}
}
