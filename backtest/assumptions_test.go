package backtest

import (
	"strings"
	"testing"
)

func TestDefaultBacktestAssumptionsAreHonest(t *testing.T) {
	assumptions := DefaultBacktestAssumptions()

	checks := []string{
		assumptions.EngineModel,
		assumptions.FillModel,
		assumptions.LatencyAssumption,
		assumptions.PartialFillSupport,
		assumptions.MarketImpactSupport,
		assumptions.EventDrivenSupportStatus,
		assumptions.PaperForwardTestingStatus,
	}
	joined := strings.ToLower(strings.Join(checks, " "))

	for _, want := range []string{
		"for-loop",
		"simulated",
		"latency is not fully modeled",
		"partial fills are not fully modeled",
		"market impact is not modeled",
		"event-driven backtester is not yet complete",
		"paper or forward testing is not enabled",
	} {
		if !strings.Contains(joined, want) {
			t.Fatalf("expected assumptions to contain %q in %q", want, joined)
		}
	}

	if len(assumptions.Gaps) == 0 {
		t.Fatal("expected assumption gaps")
	}
}
