package paper

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestPersistenceReloadAndRecorderWrites(t *testing.T) {
	cfg := tempConfig(t)
	state := NewState(cfg)
	state.Balance = 900
	state.OpenPositions = []PaperPosition{{ID: "p1", Symbol: "BTCUSDT"}}

	if err := EnsureDataFiles(cfg); err != nil {
		t.Fatalf("ensure files: %v", err)
	}
	if err := SaveState(cfg, state); err != nil {
		t.Fatalf("save state: %v", err)
	}
	loaded, err := LoadState(cfg)
	if err != nil {
		t.Fatalf("load state: %v", err)
	}
	if loaded.Balance != 900 || len(loaded.OpenPositions) != 1 {
		t.Fatalf("unexpected loaded state: %+v", loaded)
	}

	if err := AppendEquity(cfg, 1, loaded); err != nil {
		t.Fatalf("append equity: %v", err)
	}
	if err := AppendFunding(cfg, FundingEvent{Timestamp: 1, PositionID: "p1", Symbol: "BTCUSDT", Side: "LONG", Rate: 0.001, Amount: -1}); err != nil {
		t.Fatalf("append funding: %v", err)
	}
	if err := AppendEvent(cfg, TelemetryEvent{Timestamp: 1, Type: "decision"}); err != nil {
		t.Fatalf("append event: %v", err)
	}
	if err := AppendTrade(cfg, 1, loaded.OpenPositions[0], 101, 1, 2, "tp1"); err != nil {
		t.Fatalf("append trade: %v", err)
	}

	for _, path := range []string{cfg.TradesPath, cfg.EquityPath, cfg.EventsPath, cfg.FundingPath} {
		body, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("read %s: %v", path, err)
		}
		if len(body) == 0 {
			t.Fatalf("expected content in %s", path)
		}
	}
}

func tempConfig(t *testing.T) Config {
	t.Helper()
	dir := t.TempDir()
	cfg := DefaultConfig()
	cfg.Mode = "paper"
	cfg.UniverseMode = "manual"
	cfg.StatePath = filepath.Join(dir, "state.json")
	cfg.PositionsPath = filepath.Join(dir, "positions.json")
	cfg.TradesPath = filepath.Join(dir, "trades.csv")
	cfg.EquityPath = filepath.Join(dir, "equity.csv")
	cfg.EventsPath = filepath.Join(dir, "events.jsonl")
	cfg.FundingPath = filepath.Join(dir, "funding.csv")
	cfg.UniverseCSVPath = filepath.Join(dir, "universe.csv")
	cfg.UniverseJSONPath = filepath.Join(dir, "universe.json")
	cfg.DiscoveredUniversePath = filepath.Join(dir, "discovered_universe.json")
	cfg.QualifiedUniversePath = filepath.Join(dir, "qualified_universe.json")
	cfg.SelectedUniversePath = filepath.Join(dir, "selected_universe.json")
	cfg.QualificationDiagnosticsPath = filepath.Join(dir, "qualification_diagnostics.json")
	cfg.QualifiedNotSelectedPath = filepath.Join(dir, "qualified_not_selected.json")
	cfg.UniverseBiasDiagnosticsPath = filepath.Join(dir, "universe_bias_diagnostics.json")
	cfg.CandidateCoveragePath = filepath.Join(dir, "candidate_coverage.json")
	cfg.CandidateQualitySummaryPath = filepath.Join(dir, "candidate_quality_summary.json")
	cfg.CandidateQualityByVenuePath = filepath.Join(dir, "candidate_quality_by_venue.json")
	cfg.CandidateQualityBySymbolPath = filepath.Join(dir, "candidate_quality_by_symbol.json")
	cfg.RankingPath = filepath.Join(dir, "ranking.json")
	cfg.RankingDiagnosticsPath = filepath.Join(dir, "ranking_diagnostics.json")
	cfg.StrategyReviewCandidatesPath = filepath.Join(dir, "strategy_review_candidates.jsonl")
	cfg.StrategyReviewRejectionsPath = filepath.Join(dir, "strategy_review_rejections.jsonl")
	cfg.StrategyReviewSelectedPath = filepath.Join(dir, "strategy_review_selected.jsonl")
	cfg.StrategyReviewPositionsPath = filepath.Join(dir, "strategy_review_positions.jsonl")
	cfg.StrategyReviewCyclesPath = filepath.Join(dir, "strategy_review_cycles.jsonl")
	cfg.OpportunityCoveragePath = filepath.Join(dir, "opportunity_coverage.json")
	cfg.OpportunityCoverageByVenuePath = filepath.Join(dir, "opportunity_coverage_by_venue.json")
	cfg.OvernightStrategySummaryJSONPath = filepath.Join(dir, "overnight_strategy_summary.json")
	cfg.OvernightStrategySummaryMarkdownPath = filepath.Join(dir, "overnight_strategy_summary.md")
	cfg.SummaryJSONPath = filepath.Join(dir, "summary.json")
	cfg.RejectSummaryJSONPath = filepath.Join(dir, "reject_summary.json")
	return cfg
}

func TestEnsureDataFilesCreatesHeaders(t *testing.T) {
	cfg := tempConfig(t)
	if err := EnsureDataFiles(cfg); err != nil {
		t.Fatalf("ensure files: %v", err)
	}
	body, err := os.ReadFile(cfg.TradesPath)
	if err != nil {
		t.Fatalf("read trades header: %v", err)
	}
	if !strings.Contains(string(body), "position_id") {
		t.Fatalf("expected csv header, got %s", string(body))
	}
}

func TestBuildCandidateQualitySummarizesNonMajors(t *testing.T) {
	rows := []CandidateCoverageRow{
		{Venue: "hyperliquid", Symbol: "BTC", CanonicalSymbol: "BTC", Major: true, Candidates: 2, Approved: 1, Rejected: 1, ConfidenceSum: 1.2, RejectReasons: map[string]int{"poor_rr": 1}},
		{Venue: "hyperliquid", Symbol: "ARB", CanonicalSymbol: "ARB", Major: false, Candidates: 3, Approved: 2, Rejected: 1, ConfidenceSum: 2.1, RejectReasons: map[string]int{"spread_too_wide": 1}},
	}
	summary, byVenue, bySymbol := BuildCandidateQuality(rows)
	if summary.NonMajorSymbols != 1 || summary.NonMajorCandidates != 3 || summary.NonMajorApproved != 2 {
		t.Fatalf("unexpected quality summary: %+v", summary)
	}
	if summary.AverageConfidence < 0.65 || summary.AverageConfidence > 0.67 || summary.NonMajorAverageConfidence < 0.69 || summary.NonMajorAverageConfidence > 0.71 {
		t.Fatalf("unexpected confidence aggregation: %+v", summary)
	}
	if len(byVenue) != 1 || byVenue[0].Candidates != 5 {
		t.Fatalf("unexpected venue quality: %+v", byVenue)
	}
	if len(bySymbol) != 2 {
		t.Fatalf("unexpected symbol quality: %+v", bySymbol)
	}
}

func TestOvernightReviewArtifacts(t *testing.T) {
	cfg := tempConfig(t)
	now := int64(1770500000000)
	event := TelemetryEvent{
		Timestamp:  now,
		Type:       "CANDIDATE_CREATED",
		Venue:      "hyperliquid",
		Symbol:     "ARB",
		Strategy:   "VWAP",
		Playbook:   "VWAP Successful Reaction",
		Side:       "LONG",
		Confidence: 0.72,
		Score:      0.81,
		Entry:      1.23,
	}
	if err := AppendStrategyReviewEvent(cfg, "candidate", event); err != nil {
		t.Fatalf("append candidate review: %v", err)
	}
	if err := AppendStrategyReviewSelected(cfg, []UniverseEntry{{Venue: "hyperliquid", Symbol: "ARB", CanonicalSymbol: "ARB", Rank: 1}}, now); err != nil {
		t.Fatalf("append selected review: %v", err)
	}
	candidates, err := os.ReadFile(cfg.StrategyReviewCandidatesPath)
	if err != nil || !strings.Contains(string(candidates), `"symbol":"ARB"`) {
		t.Fatalf("expected candidate review jsonl, err=%v body=%s", err, string(candidates))
	}
	selected, err := os.ReadFile(cfg.StrategyReviewSelectedPath)
	if err != nil || !strings.Contains(string(selected), `"selected"`) {
		t.Fatalf("expected selected review jsonl, err=%v body=%s", err, string(selected))
	}
}

func TestOpportunityCoverageAndOvernightSummaryArtifacts(t *testing.T) {
	cfg := tempConfig(t)
	summary := RuntimeSummary{
		Mode:            "paper",
		ExecutionMode:   "simulated",
		SelectedSymbols: 2,
		Approved:        1,
		Rejected:        1,
		SelectedUniverse: []UniverseEntry{
			{Venue: "hyperliquid", Symbol: "BTC", CanonicalSymbol: "BTC"},
			{Venue: "hyperliquid", Symbol: "ARB", CanonicalSymbol: "ARB"},
		},
		CandidateCoverage: []CandidateCoverageRow{
			{Venue: "hyperliquid", Symbol: "BTC", CanonicalSymbol: "BTC", Major: true, Candidates: 1, Approved: 1, ConfidenceSum: 0.7},
			{Venue: "hyperliquid", Symbol: "ARB", CanonicalSymbol: "ARB", Major: false, Candidates: 2, Rejected: 1, ConfidenceSum: 1.2, RejectReasons: map[string]int{"poor_rr": 1}},
		},
		RejectReasons: map[string]int{"poor_rr": 1},
	}
	if err := WriteOpportunityCoverageJSON(cfg, summary); err != nil {
		t.Fatalf("write opportunity coverage: %v", err)
	}
	if err := WriteOvernightStrategySummary(cfg, summary); err != nil {
		t.Fatalf("write overnight summary: %v", err)
	}
	coverage, err := os.ReadFile(cfg.OpportunityCoveragePath)
	if err != nil || !strings.Contains(string(coverage), `"nonMajorCandidates": 2`) {
		t.Fatalf("expected opportunity coverage, err=%v body=%s", err, string(coverage))
	}
	md, err := os.ReadFile(cfg.OvernightStrategySummaryMarkdownPath)
	if err != nil || !strings.Contains(string(md), "Overnight Strategy Summary") {
		t.Fatalf("expected overnight markdown, err=%v body=%s", err, string(md))
	}
}
