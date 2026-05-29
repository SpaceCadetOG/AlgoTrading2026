package backtest

import "testing"

func TestDrawdownTracker(t *testing.T) {
	tracker := NewDrawdownTracker(100)
	tracker.Update(120)
	tracker.Update(90)

	if tracker.PeakEquity != 120 {
		t.Fatalf("peak = %f, want 120", tracker.PeakEquity)
	}
	if tracker.MaxDrawdown != 30 {
		t.Fatalf("max drawdown = %f, want 30", tracker.MaxDrawdown)
	}
	if tracker.MaxDrawdownPct != 25 {
		t.Fatalf("max drawdown pct = %f, want 25", tracker.MaxDrawdownPct)
	}
}
