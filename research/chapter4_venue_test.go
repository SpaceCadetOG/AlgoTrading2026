package research

import (
	"os"
	"strings"
	"testing"

	"AlgoTrading2026/backtest"
)

func TestWriteChapter4VenueComparisonCSV(t *testing.T) {
	path := t.TempDir() + "/venue_comparison.csv"
	rows := []StrategyComparisonRow{{
		Strategy: "momentum",
		Report: backtest.Report{
			Venue:       "aster",
			Symbol:      "BTCUSDT",
			Interval:    "15m",
			CandleCount: 100,
		},
	}}

	if err := WriteChapter4VenueComparisonCSV(path, rows); err != nil {
		t.Fatal(err)
	}
	body, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	text := string(body)
	if !strings.Contains(text, "venue,symbol,interval,strategy") {
		t.Fatalf("missing header: %s", text)
	}
	if !strings.Contains(text, "aster,BTCUSDT,15m,momentum") {
		t.Fatalf("missing row: %s", text)
	}
}

func TestWriteChapter4VenueAnalysisCSV(t *testing.T) {
	path := t.TempDir() + "/venue_analysis.csv"
	rows := []Chapter4Analysis{{
		Strategy: "momentum",
		Venue:    "aster",
		Symbol:   "BTCUSDT",
		Interval: "15m",
		Notes:    []string{"low_sample_size"},
	}}

	if err := WriteChapter4VenueAnalysisCSV(path, rows); err != nil {
		t.Fatal(err)
	}
	body, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	text := string(body)
	if !strings.Contains(text, "venue,symbol,interval,strategy") {
		t.Fatalf("missing header: %s", text)
	}
	if !strings.Contains(text, "aster,BTCUSDT,15m,momentum") {
		t.Fatalf("missing row: %s", text)
	}
}
