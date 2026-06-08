package paper

import "testing"

func TestRiskRejectsPoorRRAndLocks(t *testing.T) {
	cfg := DefaultConfig()
	state := NewState(cfg)
	state.SymbolCooldowns["BTCUSDT"] = 1000
	state.SymbolLocks["BTCUSDT"] = 1000

	candidate := Candidate{
		Symbol:      "BTCUSDT",
		Side:        "LONG",
		EntryPrice:  100,
		StopPrice:   99.5,
		TP1:         100.2,
		Liquidity:   1000,
		SpreadPct:   0.01,
		RequiredRR:  1.0,
		FundingRate: 0,
	}
	decision := CheckRisk(state, candidate, cfg, 500)
	if decision.Allowed {
		t.Fatalf("expected rejection, got %+v", decision)
	}
	if len(decision.Reasons) < 3 {
		t.Fatalf("expected multiple reject reasons, got %+v", decision)
	}
}
