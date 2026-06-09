package paper

import "testing"

func TestRiskRejectsPoorRRAndLocks(t *testing.T) {
	cfg := DefaultConfig()
	state := NewState(cfg)
	state.SymbolCooldowns["aster:BTCUSDT"] = 1000
	state.SymbolLocks["aster:BTCUSDT"] = 1000

	candidate := Candidate{
		Venue:       "aster",
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

func TestRiskStateKeysAreVenueSpecific(t *testing.T) {
	cfg := DefaultConfig()
	state := NewState(cfg)
	state.SymbolCooldowns["hyperliquid:BTC"] = 1000
	state.DailyTradeCount["hyperliquid:BTC"] = cfg.MaxTradesPerSymbolPerDay

	candidate := Candidate{
		Venue:       "lighter",
		Symbol:      "BTC",
		Side:        "LONG",
		EntryPrice:  100,
		StopPrice:   99,
		TP1:         102,
		Liquidity:   1000,
		SpreadPct:   0.01,
		RequiredRR:  1.0,
		FundingRate: 0,
	}
	decision := CheckRisk(state, candidate, cfg, 500)
	for _, reason := range decision.Reasons {
		if reason == "symbol_cooldown" || reason == "max_trades_per_symbol_per_day" {
			t.Fatalf("venue-specific hyperliquid state should not block lighter candidate: %+v", decision)
		}
	}
}
