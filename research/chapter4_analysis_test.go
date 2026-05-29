package research

import (
	"os"
	"strings"
	"testing"

	"AlgoTrading2026/backtest"
)

func TestWriteChapter4StrategyComparisonCSV(t *testing.T) {
	path := t.TempDir() + "/chapter4_comparison.csv"
	rows := []StrategyComparisonRow{{
		Strategy: "momentum",
		Report: backtest.Report{
			Venue:       "aster",
			Symbol:      "BTCUSDT",
			Interval:    "15m",
			CandleCount: 100,
		},
	}}

	if err := WriteChapter4StrategyComparisonCSV(path, rows); err != nil {
		t.Fatal(err)
	}

	body, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	text := string(body)
	if !strings.Contains(text, "strategy,venue,symbol,interval") {
		t.Fatalf("missing header: %s", text)
	}
	if !strings.Contains(text, "momentum,aster,BTCUSDT,15m") {
		t.Fatalf("missing row: %s", text)
	}
}

func TestWriteChapter4AnalysisCSV(t *testing.T) {
	path := t.TempDir() + "/chapter4_analysis.csv"
	rows := []Chapter4Analysis{{
		Strategy: "momentum",
		Venue:    "aster",
		Symbol:   "BTCUSDT",
		Interval: "15m",
		Notes:    []string{"low_sample_size"},
	}}

	if err := WriteChapter4AnalysisCSV(path, rows); err != nil {
		t.Fatal(err)
	}

	body, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	text := string(body)
	if !strings.Contains(text, "strategy,venue,symbol,interval") {
		t.Fatalf("missing header: %s", text)
	}
	if !strings.Contains(text, "momentum,aster,BTCUSDT,15m") {
		t.Fatalf("missing row: %s", text)
	}
}
