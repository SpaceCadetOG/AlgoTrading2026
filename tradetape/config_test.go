package tradetape

import (
	"path/filepath"
	"testing"
)

func TestDefaultRecorderConfigUsesDataRoot(t *testing.T) {
	t.Setenv("DATA_ROOT", filepath.Join("C:", "tmp", "algo-data"))

	cfg := DefaultRecorderConfig()
	if cfg.OutputPath != filepath.Join("C:", "tmp", "algo-data", "trade_tape", "trades.csv") {
		t.Fatalf("unexpected output path: %s", cfg.OutputPath)
	}
}

func TestDefaultBackfillConfigUsesDataRoot(t *testing.T) {
	t.Setenv("DATA_ROOT", filepath.Join("C:", "tmp", "algo-data"))

	cfg := DefaultBackfillConfig()
	if cfg.OutputPath != filepath.Join("C:", "tmp", "algo-data", "trade_tape", "backfill", "aster_btcusdt.csv") {
		t.Fatalf("unexpected output path: %s", cfg.OutputPath)
	}
	if cfg.DedupedOutputPath != filepath.Join("C:", "tmp", "algo-data", "trade_tape", "backfill", "aster_btcusdt_deduped.csv") {
		t.Fatalf("unexpected deduped path: %s", cfg.DedupedOutputPath)
	}
}
