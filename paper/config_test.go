package paper

import (
	"path/filepath"
	"testing"
)

func TestDefaultConfigUsesDataRoot(t *testing.T) {
	t.Setenv("DATA_ROOT", filepath.Join("C:", "tmp", "algo-data"))

	cfg := DefaultConfig()
	root := filepath.Join("C:", "tmp", "algo-data", "paper")

	if cfg.StatePath != filepath.Join(root, "state.json") {
		t.Fatalf("unexpected state path: %s", cfg.StatePath)
	}
	if cfg.PositionsPath != filepath.Join(root, "positions.json") {
		t.Fatalf("unexpected positions path: %s", cfg.PositionsPath)
	}
	if cfg.TradesPath != filepath.Join(root, "trades.csv") {
		t.Fatalf("unexpected trades path: %s", cfg.TradesPath)
	}
	if cfg.EquityPath != filepath.Join(root, "equity.csv") {
		t.Fatalf("unexpected equity path: %s", cfg.EquityPath)
	}
	if cfg.EventsPath != filepath.Join(root, "events.jsonl") {
		t.Fatalf("unexpected events path: %s", cfg.EventsPath)
	}
	if cfg.FundingPath != filepath.Join(root, "funding.csv") {
		t.Fatalf("unexpected funding path: %s", cfg.FundingPath)
	}
}
