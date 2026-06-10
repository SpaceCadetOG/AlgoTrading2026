package paper

import (
	"os"
	"strings"
	"testing"
)

func TestConfigAliasesPreferSmallOperatorNames(t *testing.T) {
	t.Setenv("MAX_SYMBOLS", "17")
	t.Setenv("PAPER_MAX_SYMBOLS", "3")
	t.Setenv("VENUES", "aster,hyperliquid")
	t.Setenv("PAPER_VENUES", "lighter")
	t.Setenv("MIN_24H_VOLUME_USD", "123")
	t.Setenv("PAPER_MIN_24H_VOLUME_USD", "999")
	t.Setenv("MAX_OPEN_POSITIONS", "7")
	t.Setenv("PAPER_MAX_OPEN_POSITIONS", "2")
	t.Setenv("RESET_PAPER_STATE", "true")

	cfg := DefaultConfig()
	if cfg.MaxSymbols != 17 || cfg.Min24hVolumeUSD != 123 || cfg.MaxOpenPositions != 7 || !cfg.ResetState {
		t.Fatalf("small operator aliases not applied: %+v", cfg)
	}
	if len(cfg.Venues) != 2 || cfg.Venues[0] != "aster" || cfg.Venues[1] != "hyperliquid" {
		t.Fatalf("unexpected venues: %+v", cfg.Venues)
	}
}

func TestConfigKnobClassificationIncludesNoConsoleStyle(t *testing.T) {
	for _, knob := range ConfigKnobs() {
		if strings.EqualFold(knob.Name, "PAPER_CONSOLE_STYLE") {
			t.Fatalf("PAPER_CONSOLE_STYLE must not exist")
		}
	}
	if len(ConfigKnobsByClass("operator")) == 0 || len(ConfigKnobsByClass("live_safety")) == 0 {
		t.Fatalf("expected operator and live safety classifications")
	}
}

func TestBuildRankingOrdersOpenApprovedBlockedUnavailable(t *testing.T) {
	summary := RuntimeSummary{
		SelectedUniverse: []UniverseEntry{
			{Venue: "aster", Symbol: "OPENUSDT", CanonicalSymbol: "OPEN", Volume24hUSD: 10_000_000, VolumeKnown: true, LastPrice: 10, Rank: 1},
			{Venue: "aster", Symbol: "APPUSDT", CanonicalSymbol: "APP", Volume24hUSD: 20_000_000, VolumeKnown: true, LastPrice: 20, Rank: 2},
			{Venue: "aster", Symbol: "BLOCKUSDT", CanonicalSymbol: "BLOCK", Volume24hUSD: 50_000_000, VolumeKnown: true, LastPrice: 50, Rank: 3},
			{Venue: "lighter", Symbol: "NOBK", CanonicalSymbol: "NOBK", LastPrice: 0, Rank: 4},
		},
		CandidateCoverage: []CandidateCoverageRow{
			{Venue: "aster", Symbol: "OPENUSDT", SnapshotFetched: true, Candidates: 1, Approved: 1, AverageConfidence: 0.60, RejectReasons: map[string]int{}},
			{Venue: "aster", Symbol: "APPUSDT", SnapshotFetched: true, Candidates: 1, Approved: 1, AverageConfidence: 0.95, RejectReasons: map[string]int{}},
			{Venue: "aster", Symbol: "BLOCKUSDT", SnapshotFetched: true, Candidates: 1, Rejected: 1, AverageConfidence: 0.95, RejectReasons: map[string]int{"max_open_positions": 1}},
			{Venue: "lighter", Symbol: "NOBK", SnapshotFetched: false, RejectReasons: map[string]int{"orderbook_fetch_failed": 1}},
		},
	}
	state := EngineState{
		OpenPositions: []PaperPosition{{Venue: "aster", Symbol: "OPENUSDT", Side: "LONG"}},
		RecentDecisions: []TelemetryEvent{
			{Type: "decision", Venue: "aster", Symbol: "APPUSDT", Side: "LONG", Decision: "approved", Score: 0.95, Confidence: 0.95, Entry: 20, StopDistance: 1},
			{Type: "decision", Venue: "aster", Symbol: "BLOCKUSDT", Side: "LONG", Decision: "rejected", Score: 0.95, Confidence: 0.95, Reasons: []string{"max_open_positions"}},
		},
	}
	rows := BuildRanking(summary, state)
	if len(rows) != 4 {
		t.Fatalf("rows=%d want 4", len(rows))
	}
	if rows[0].State != "open" || rows[1].State != "approved" || rows[2].State != "blocked" || rows[3].State != "unavailable" {
		t.Fatalf("unexpected ranking states: %+v", rows)
	}
	if rows[3].Penalties["unknown_volume"] == 0 || rows[3].Penalties["no_orderbook"] == 0 {
		t.Fatalf("expected unknown/no-book penalties on unavailable row: %+v", rows[3])
	}
}

func TestWriteRankingJSONWritesDiagnostics(t *testing.T) {
	cfg := tempConfig(t)
	summary := RuntimeSummary{
		SelectedUniverse: []UniverseEntry{{Venue: "hyperliquid", Symbol: "DOGE", CanonicalSymbol: "DOGE", Volume24hUSD: 7_000_000, VolumeKnown: true, Rank: 1}},
		CandidateCoverage: []CandidateCoverageRow{
			{Venue: "hyperliquid", Symbol: "DOGE", SnapshotFetched: true, Candidates: 1, Approved: 1, AverageConfidence: 0.8, RejectReasons: map[string]int{}},
		},
	}
	state := EngineState{RecentDecisions: []TelemetryEvent{{Type: "decision", Venue: "hyperliquid", Symbol: "DOGE", Side: "LONG", Decision: "approved", Score: 0.8, Confidence: 0.8}}}
	if err := WriteRankingJSON(cfg, summary, state); err != nil {
		t.Fatalf("write ranking: %v", err)
	}
	for _, path := range []string{cfg.RankingPath, cfg.RankingDiagnosticsPath} {
		body, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("read %s: %v", path, err)
		}
		if !strings.Contains(string(body), "DOGE") {
			t.Fatalf("expected DOGE in %s: %s", path, string(body))
		}
	}
}
