package tradetape

import (
	"path/filepath"
	"testing"
)

type fakeBackfillProvider struct {
	prints []TradeTapePrint
}

func (f fakeBackfillProvider) Venue() string { return "fake" }

func (f fakeBackfillProvider) BackfillTrades(symbol string, startMS int64, endMS int64) ([]TradeTapePrint, error) {
	return f.prints, nil
}

func TestRunBackfillWritesRawAndDedupedRows(t *testing.T) {
	dir := t.TempDir()
	cfg := BackfillConfig{
		Symbol:            "BTCUSDT",
		StartMS:           1,
		EndMS:             2,
		OutputPath:        filepath.Join(dir, "raw.csv"),
		DedupedOutputPath: filepath.Join(dir, "deduped.csv"),
	}
	provider := fakeBackfillProvider{prints: []TradeTapePrint{
		dedupePrint("1", 100, 1, "BUY"),
		dedupePrint("1", 100, 1, "BUY"),
	}}
	rawRows, dedupedRows, summary, err := RunBackfill(provider, cfg)
	if err != nil {
		t.Fatalf("run backfill: %v", err)
	}
	if len(rawRows) != 2 || len(dedupedRows) != 1 || summary.DuplicateRows != 1 {
		t.Fatalf("unexpected backfill output: raw=%+v deduped=%+v summary=%+v", rawRows, dedupedRows, summary)
	}
	readRaw, err := ReadRows(cfg.OutputPath)
	if err != nil || len(readRaw) != 2 {
		t.Fatalf("read raw rows: len=%d err=%v", len(readRaw), err)
	}
	readDeduped, err := ReadRows(cfg.DedupedOutputPath)
	if err != nil || len(readDeduped) != 1 {
		t.Fatalf("read deduped rows: len=%d err=%v", len(readDeduped), err)
	}
}
