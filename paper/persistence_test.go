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
	cfg.StatePath = filepath.Join(dir, "state.json")
	cfg.PositionsPath = filepath.Join(dir, "positions.json")
	cfg.TradesPath = filepath.Join(dir, "trades.csv")
	cfg.EquityPath = filepath.Join(dir, "equity.csv")
	cfg.EventsPath = filepath.Join(dir, "events.jsonl")
	cfg.FundingPath = filepath.Join(dir, "funding.csv")
	cfg.UniverseCSVPath = filepath.Join(dir, "universe.csv")
	cfg.UniverseJSONPath = filepath.Join(dir, "universe.json")
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
