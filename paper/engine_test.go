package paper

import (
	"context"
	"os"
	"strings"
	"testing"

	runtime "AlgoTrading2026/internal/runtime"
	"AlgoTrading2026/orderbook"
)

type testCandidateBuilder struct {
	candidates []Candidate
}

func (b testCandidateBuilder) BuildCandidates(ctx runtime.StrategyContext) []Candidate {
	out := make([]Candidate, 0, len(b.candidates))
	for _, candidate := range b.candidates {
		candidate.Venue = ctx.Snapshot.Venue
		candidate.Symbol = ctx.Snapshot.Symbol
		candidate.CanonicalSymbol = canonicalFromSymbol(ctx.Snapshot.Symbol)
		candidate.EntryPrice = 100
		candidate.Quantity = 1
		candidate.Leverage = 1
		candidate.SpreadPct = 0.01
		candidate.Liquidity = 1000
		candidate.RequiredRR = 1
		out = append(out, candidate)
	}
	return out
}

func TestEngineFlagsAndPlaybooks(t *testing.T) {
	cfg := tempConfig(t)
	engine, err := NewEngine(cfg)
	if err != nil {
		t.Fatalf("new engine: %v", err)
	}
	if engine.State.LiveEnabled || engine.State.ExecutionEnabled || !engine.State.PaperEnabled {
		t.Fatalf("unexpected engine safety flags: %+v", engine.State)
	}
	if len(engine.Playbooks) != 9 {
		t.Fatalf("expected 9 playbooks, got %d", len(engine.Playbooks))
	}
}

func TestRunLoopExecutesConfiguredRuntimeCycle(t *testing.T) {
	cfg := tempConfig(t)
	cfg.EnableRecorders = false
	cfg.MaxRuntimeCycles = 1
	cfg.Symbols = []string{"BTC"}
	cfg.Venues = []string{"aster"}

	engine, err := NewEngine(cfg)
	if err != nil {
		t.Fatalf("new engine: %v", err)
	}
	engine.SelectUniverse = func(cfg Config) (UniverseSelection, error) {
		return UniverseSelection{
			Mode: "manual",
			Selected: []UniverseEntry{
				{Venue: "aster", Symbol: "BTCUSDT", CanonicalSymbol: "BTC"},
			},
		}, nil
	}
	engine.FetchOrderBook = func(venue string, symbol string) (orderbook.OrderBookSnapshot, error) {
		return testSnapshot(), nil
	}

	cycles := 0
	if err := engine.RunLoop(context.Background(), func(snapshot RuntimeLoopSnapshot) {
		cycles++
		if snapshot.Summary.Decisions == 0 {
			t.Fatalf("expected scanner decisions, got %+v", snapshot.Summary)
		}
	}); err != nil {
		t.Fatalf("run loop: %v", err)
	}
	if cycles != 1 {
		t.Fatalf("cycles=%d want 1", cycles)
	}
	for _, path := range []string{
		cfg.SummaryJSONPath,
		cfg.RejectSummaryJSONPath,
		cfg.UniverseJSONPath,
		cfg.PositionsPath,
	} {
		body, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("read runtime export %s: %v", path, err)
		}
		if len(body) == 0 {
			t.Fatalf("expected runtime export content in %s", path)
		}
	}
}

func TestNewEngineResetStateClearsSavedBlockers(t *testing.T) {
	cfg := tempConfig(t)
	state := NewState(cfg)
	state.SymbolCooldowns["BTCUSDT"] = 9999999999999
	state.SymbolLocks["ETHUSDT"] = 9999999999999
	state.DailyTradeCount["BTCUSDT"] = cfg.MaxTradesPerSymbolPerDay
	state.RecentDecisions = []TelemetryEvent{{Type: "decision"}}
	state.RecentClosed = []PaperPosition{{Venue: "aster", Symbol: "ETHUSDT"}}
	state.OpenPositions = []PaperPosition{{Venue: "aster", Symbol: "BTCUSDT"}}
	if err := SaveState(cfg, state); err != nil {
		t.Fatalf("save state: %v", err)
	}

	cfg.ResetState = true
	engine, err := NewEngine(cfg)
	if err != nil {
		t.Fatalf("new engine: %v", err)
	}
	if len(engine.State.SymbolCooldowns) != 0 || len(engine.State.SymbolLocks) != 0 || len(engine.State.DailyTradeCount) != 0 || len(engine.State.RecentDecisions) != 0 {
		t.Fatalf("expected reset state blockers cleared, got %+v", engine.State)
	}
	if len(engine.State.OpenPositions) != 1 {
		t.Fatalf("reset state should preserve open positions unless requested, got %+v", engine.State.OpenPositions)
	}
	if len(engine.State.RecentClosed) != 0 {
		t.Fatalf("reset state should clear recent closed history, got %+v", engine.State.RecentClosed)
	}
}

func TestNewEngineResetAllClearsOpenPositions(t *testing.T) {
	cfg := tempConfig(t)
	state := NewState(cfg)
	state.OpenPositions = []PaperPosition{{Venue: "aster", Symbol: "BTCUSDT"}}
	state.RecentClosed = []PaperPosition{{Venue: "aster", Symbol: "ETHUSDT"}}
	if err := SaveState(cfg, state); err != nil {
		t.Fatalf("save state: %v", err)
	}

	cfg.ResetAll = true
	engine, err := NewEngine(cfg)
	if err != nil {
		t.Fatalf("new engine: %v", err)
	}
	if len(engine.State.OpenPositions) != 0 || len(engine.State.RecentClosed) != 0 {
		t.Fatalf("expected reset all to clear paper state, got %+v", engine.State)
	}
}

func TestEngineCandidateOpenAndStatusPayload(t *testing.T) {
	cfg := tempConfig(t)
	engine, err := NewEngine(cfg)
	if err != nil {
		t.Fatalf("new engine: %v", err)
	}
	if err := EnsureDataFiles(cfg); err != nil {
		t.Fatalf("ensure data files: %v", err)
	}
	decision, position, err := engine.EvaluateCandidate(Candidate{
		Venue:           "aster",
		Symbol:          "BTCUSDT",
		CanonicalSymbol: "BTC",
		Strategy:        "paper_test",
		Playbook:        "VWAP Successful Reaction",
		Side:            "LONG",
		Score:           0.7,
		Confidence:      0.8,
		EntryPrice:      101,
		StopPrice:       99,
		TP1:             103.5,
		TP2:             105,
		TP3:             107,
		Quantity:        1,
		Leverage:        2,
		SpreadPct:       0.01,
		Liquidity:       1000,
		RequiredRR:      1.0,
	}, testSnapshot(), false)
	if err != nil {
		t.Fatalf("evaluate candidate: %v", err)
	}
	if !decision.Allowed || position == nil {
		t.Fatalf("expected opened position, decision=%+v position=%+v", decision, position)
	}
	payload := engine.StatusPayload()
	if payload.Paper.OpenCount != 1 {
		t.Fatalf("expected one open position, got %+v", payload)
	}
}

func TestEngineFundingAndManagePosition(t *testing.T) {
	cfg := tempConfig(t)
	engine, err := NewEngine(cfg)
	if err != nil {
		t.Fatalf("new engine: %v", err)
	}
	if err := EnsureDataFiles(cfg); err != nil {
		t.Fatalf("ensure data files: %v", err)
	}
	_, _, err = engine.EvaluateCandidate(Candidate{
		Venue:           "aster",
		Symbol:          "BTCUSDT",
		CanonicalSymbol: "BTC",
		Strategy:        "paper_test",
		Playbook:        "VP Reversal + OF Absorption",
		Side:            "LONG",
		Score:           0.7,
		Confidence:      0.8,
		EntryPrice:      101,
		StopPrice:       99,
		TP1:             104,
		TP2:             104,
		TP3:             106,
		Quantity:        1,
		Leverage:        2,
		SpreadPct:       0.01,
		Liquidity:       1000,
		RequiredRR:      1.0,
	}, testSnapshot(), false)
	if err != nil {
		t.Fatalf("open candidate: %v", err)
	}
	if err := engine.ApplyFunding(0, 0.001); err != nil {
		t.Fatalf("apply funding: %v", err)
	}
	snapshot := testSnapshot()
	snapshot.Bids[0].Price = "105"
	snapshot.Bids[1].Price = "104"
	snapshot.Asks[0].Price = "106"
	snapshot.Asks[1].Price = "107"
	if err := engine.ManagePosition(0, snapshot, false, 0, false); err != nil {
		t.Fatalf("manage position: %v", err)
	}
	if len(engine.State.OpenPositions) != 1 || !engine.State.OpenPositions[0].TP1Taken {
		t.Fatalf("expected tp1 partial to mark position, got %+v", engine.State.OpenPositions)
	}
}

func TestRecordersStartWhenEnabledAndRuntimeContinuesOnRecorderErrors(t *testing.T) {
	cfg := tempConfig(t)
	cfg.EnableRecorders = true
	cfg.Symbols = []string{"BTC"}
	cfg.Venues = []string{"aster"}

	tradeCalled := 0
	l2Called := 0
	engine, err := NewEngine(cfg)
	if err != nil {
		t.Fatalf("new engine: %v", err)
	}
	engine.RunTradeRecorder = func() error {
		tradeCalled++
		return os.ErrDeadlineExceeded
	}
	engine.RunL2Recorder = func() error {
		l2Called++
		return nil
	}
	engine.FetchOrderBook = func(venue string, symbol string) (orderbook.OrderBookSnapshot, error) {
		return testSnapshot(), nil
	}
	summary, err := engine.Run()
	if err != nil {
		t.Fatalf("engine run should survive recorder error: %v", err)
	}
	if tradeCalled != 1 || l2Called != 1 {
		t.Fatalf("expected both recorders called, trade=%d l2=%d", tradeCalled, l2Called)
	}
	if !summary.RecordersEnabled || engine.State.RecordersStatus.TradeTape != "error" || engine.State.RecordersStatus.L2 != "enabled" {
		t.Fatalf("unexpected recorder status: summary=%+v state=%+v", summary, engine.State.RecordersStatus)
	}
}

func TestScannerLoopWritesDecisionEventAndRuntimeCandidateOpensAndPersists(t *testing.T) {
	cfg := tempConfig(t)
	cfg.Symbols = []string{"BTC"}
	cfg.Venues = []string{"aster"}
	cfg.AllowSyntheticCandidate = false
	cfg.EnableRecorders = false

	engine, err := NewEngine(cfg)
	if err != nil {
		t.Fatalf("new engine: %v", err)
	}
	engine.SelectUniverse = func(cfg Config) (UniverseSelection, error) {
		return UniverseSelection{
			Mode:            "volume_filter",
			Min24hVolumeUSD: 5000000,
			MaxSymbols:      25,
			Selected: []UniverseEntry{
				{Venue: "aster", Symbol: "BTCUSDT", CanonicalSymbol: "BTC", Volume24hUSD: 10000000, Rank: 1},
			},
		}, nil
	}
	engine.FetchOrderBook = func(venue string, symbol string) (orderbook.OrderBookSnapshot, error) {
		s := orderbook.OrderBookSnapshot{
			Venue:  "aster",
			Symbol: "BTCUSDT",
			Bids: []orderbook.BookLevel{
				{Price: "100.00", Size: "5"},
				{Price: "99.99", Size: "5"},
			},
			Asks: []orderbook.BookLevel{
				{Price: "100.01", Size: "5"},
				{Price: "100.02", Size: "5"},
			},
		}
		s.Venue = "aster"
		s.Symbol = "BTCUSDT"
		return s, nil
	}
	summary, err := engine.Run()
	if err != nil {
		t.Fatalf("run engine: %v", err)
	}
	if summary.Decisions == 0 || summary.Approved == 0 {
		t.Fatalf("expected decisions and approval, got %+v", summary)
	}
	if summary.UniverseMode != "volume_filter" || summary.SelectedSymbols != 1 || summary.Venues != 1 {
		t.Fatalf("unexpected universe summary: %+v", summary)
	}
	if len(engine.State.OpenPositions) == 0 {
		t.Fatalf("expected synthetic candidate to open a position, state=%+v", engine.State)
	}
	if strings.Contains(engine.State.OpenPositions[0].Provenance, "synthetic") {
		t.Fatalf("expected runtime playbook provenance, got %+v", engine.State.OpenPositions[0])
	}
	loaded, err := LoadState(cfg)
	if err != nil {
		t.Fatalf("reload state: %v", err)
	}
	if len(loaded.OpenPositions) == 0 {
		t.Fatalf("expected persisted open position, got %+v", loaded)
	}
	payload := engine.StatusPayload()
	if payload.Paper.RecentDecisionCount == 0 {
		t.Fatalf("expected recent decisions in status payload, got %+v", payload)
	}
	body, err := os.ReadFile(cfg.EventsPath)
	if err != nil {
		t.Fatalf("read events: %v", err)
	}
	if !strings.Contains(string(body), "\"type\":\"decision\"") {
		t.Fatalf("expected decision event in events file: %s", string(body))
	}
	if !strings.Contains(string(body), "\"type\":\"open\"") {
		t.Fatalf("expected open event in events file: %s", string(body))
	}
	if _, err := os.Stat(cfg.UniverseCSVPath); err != nil {
		t.Fatalf("expected universe csv export: %v", err)
	}
	if _, err := os.Stat(cfg.UniverseJSONPath); err != nil {
		t.Fatalf("expected universe json export: %v", err)
	}
}

func TestRuntimeSummaryUsesCycleCountsNotCappedRecentBuffer(t *testing.T) {
	cfg := tempConfig(t)
	cfg.Symbols = []string{"BTC", "ETH", "SOL"}
	cfg.Venues = []string{"aster", "hyperliquid", "lighter"}
	cfg.AllowSyntheticCandidate = false
	cfg.EnableRecorders = false

	engine, err := NewEngine(cfg)
	if err != nil {
		t.Fatalf("new engine: %v", err)
	}
	engine.FetchOrderBook = func(venue string, symbol string) (orderbook.OrderBookSnapshot, error) {
		switch strings.ToLower(venue) {
		case "aster":
			symbol = symbol + "USDT"
		}
		return orderbook.OrderBookSnapshot{
			Venue:  venue,
			Symbol: symbol,
			Bids: []orderbook.BookLevel{
				{Price: "100.00", Size: "5"},
				{Price: "99.99", Size: "5"},
			},
			Asks: []orderbook.BookLevel{
				{Price: "100.01", Size: "5"},
				{Price: "100.02", Size: "5"},
			},
		}, nil
	}

	summary, err := engine.Run()
	if err != nil {
		t.Fatalf("run engine: %v", err)
	}
	if summary.Decisions != 81 {
		t.Fatalf("expected 81 cycle decisions, got %+v", summary)
	}
	if summary.Approved != cfg.MaxOpenPositions {
		t.Fatalf("expected approvals up to max open positions, got %+v", summary)
	}
	if summary.Rejected != 81-cfg.MaxOpenPositions {
		t.Fatalf("expected remaining rejected decisions, got %+v", summary)
	}
	if len(engine.State.RecentDecisions) == 0 || len(engine.State.RecentDecisions) > 25 {
		t.Fatalf("expected bounded recent decisions buffer, got %d", len(engine.State.RecentDecisions))
	}
}

func TestScanDedupesSameVenueSymbolSideByScore(t *testing.T) {
	cfg := tempConfig(t)
	cfg.EnableRecorders = false
	cfg.MaxOpenPositions = 3
	engine, err := NewEngine(cfg)
	if err != nil {
		t.Fatalf("new engine: %v", err)
	}
	engine.BuildCandidates = testCandidateBuilder{candidates: []Candidate{
		{Strategy: "low", Playbook: "low", Side: "LONG", Score: 0.40, Confidence: 0.60, StopPrice: 99, TP1: 102, TP2: 103, TP3: 104},
		{Strategy: "high", Playbook: "high", Side: "LONG", Score: 0.80, Confidence: 0.70, StopPrice: 99, TP1: 102, TP2: 103, TP3: 104},
	}}
	engine.FetchOrderBook = func(venue string, symbol string) (orderbook.OrderBookSnapshot, error) {
		return testSnapshot(), nil
	}

	counts, err := engine.scanOnce([]UniverseEntry{{Venue: "aster", Symbol: "BTCUSDT", CanonicalSymbol: "BTC"}})
	if err != nil {
		t.Fatalf("scan once: %v", err)
	}
	if counts.RejectReasons["deduped_out"] != 1 {
		t.Fatalf("expected one deduped candidate, got %+v", counts)
	}
	if len(engine.State.OpenPositions) != 1 || engine.State.OpenPositions[0].Strategy != "high" {
		t.Fatalf("expected highest score candidate opened, got %+v", engine.State.OpenPositions)
	}
	admission := admissionSummaryRows(counts.Admission)
	if len(admission) != 1 || admission[0].Created != 2 || admission[0].Deduped != 1 || admission[0].Approved != 1 {
		t.Fatalf("unexpected admission summary: %+v", admission)
	}
}

func TestScanKeepsCrossVenueSymbolRoutesDistinct(t *testing.T) {
	cfg := tempConfig(t)
	cfg.EnableRecorders = false
	cfg.MaxOpenPositions = 3
	engine, err := NewEngine(cfg)
	if err != nil {
		t.Fatalf("new engine: %v", err)
	}
	engine.BuildCandidates = testCandidateBuilder{candidates: []Candidate{
		{Strategy: "route", Playbook: "route", Side: "LONG", Score: 0.70, Confidence: 0.70, StopPrice: 99, TP1: 102, TP2: 103, TP3: 104},
	}}
	engine.FetchOrderBook = func(venue string, symbol string) (orderbook.OrderBookSnapshot, error) {
		s := testSnapshot()
		s.Venue = venue
		s.Symbol = symbol
		return s, nil
	}

	counts, err := engine.scanOnce([]UniverseEntry{
		{Venue: "hyperliquid", Symbol: "BTC", CanonicalSymbol: "BTC"},
		{Venue: "lighter", Symbol: "BTC", CanonicalSymbol: "BTC"},
	})
	if err != nil {
		t.Fatalf("scan once: %v", err)
	}
	if counts.RejectReasons["deduped_out"] != 0 || len(engine.State.OpenPositions) != 2 {
		t.Fatalf("cross-venue BTC routes should both survive, counts=%+v positions=%+v", counts, engine.State.OpenPositions)
	}
	if engine.State.DailyTradeCount["hyperliquid:BTC"] != 1 || engine.State.DailyTradeCount["lighter:BTC"] != 1 {
		t.Fatalf("expected venue-specific daily counts, got %+v", engine.State.DailyTradeCount)
	}
}
