package risk

import "testing"

func TestRiskControlEngineBlocksEntryLimits(t *testing.T) {
	engine := NewRiskControlEngine(RiskControlConfig{
		MaxTradesPerDay:           1,
		MaxTradeSize:              0.5,
		MaxNotional:               50,
		MaxHoldBars:               10,
		StopLossPct:               0.02,
		MaxVolumeParticipationPct: 1,
	})

	violations := engine.CheckEntry(EntryCheck{
		Timestamp: 0,
		Strategy:  "test",
		Symbol:    "BTC",
		Size:      1,
		Notional:  100,
		Volume:    10,
	})

	if len(violations) != 3 {
		t.Fatalf("expected size, notional, and volume violations, got %d: %+v", len(violations), violations)
	}
	if MostSevereAction(violations) != ActionBlock {
		t.Fatalf("expected block action")
	}
}

func TestRiskControlEngineBlocksMaxTradesPerDay(t *testing.T) {
	engine := NewRiskControlEngine(RiskControlConfig{
		MaxTradesPerDay:           1,
		MaxTradeSize:              10,
		MaxNotional:               1000,
		MaxHoldBars:               10,
		StopLossPct:               0.02,
		MaxVolumeParticipationPct: 100,
	})
	engine.RecordTrade(0)

	violations := engine.CheckEntry(EntryCheck{
		Timestamp: 1,
		Strategy:  "test",
		Symbol:    "BTC",
		Size:      1,
		Notional:  10,
		Volume:    100,
	})
	if len(violations) != 1 || violations[0].Rule != RuleMaxTradesPerDay {
		t.Fatalf("expected max trades violation, got %+v", violations)
	}
}

func TestRiskControlEngineForceExitsStopLossAndHoldTime(t *testing.T) {
	engine := NewRiskControlEngine(RiskControlConfig{
		MaxTradesPerDay:           10,
		MaxTradeSize:              10,
		MaxNotional:               1000,
		MaxHoldBars:               3,
		StopLossPct:               0.05,
		MaxVolumeParticipationPct: 100,
	})

	violations := engine.CheckOpenPosition(PositionCheck{
		Timestamp: 0,
		Strategy:  "test",
		Symbol:    "BTC",
		Entry:     100,
		Mark:      94,
		HoldBars:  3,
	})

	if len(violations) != 2 {
		t.Fatalf("expected stop loss and hold time violations, got %+v", violations)
	}
	if MostSevereAction(violations) != ActionForceExit {
		t.Fatalf("expected force exit action")
	}
}
