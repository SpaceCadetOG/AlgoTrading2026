package paper

import (
	"os"
	"strings"
	"testing"

	"AlgoTrading2026/orderbook"
)

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
	if len(engine.State.RecentDecisions) != 25 {
		t.Fatalf("expected capped recent decisions buffer, got %d", len(engine.State.RecentDecisions))
	}
}
