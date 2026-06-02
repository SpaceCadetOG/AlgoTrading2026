package research

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"AlgoTrading2026/exchanges"
)

func TestBuildPriceActionFeatureRows(t *testing.T) {
	rows := BuildPriceActionFeatureRows(testPriceActionCandles())
	if len(rows) != 10 {
		t.Fatalf("rows=%d want 10", len(rows))
	}
	if rows[0].DailyOpenLevel == 0 {
		t.Fatalf("expected daily open: %+v", rows[0])
	}
	var hasAggression bool
	for _, row := range rows {
		if row.AggressionDirection != "none" && row.AggressionDirection != "" {
			hasAggression = true
		}
	}
	if !hasAggression {
		t.Fatal("expected at least one aggression feature")
	}
}

func TestPriceActionWriters(t *testing.T) {
	rows := BuildPriceActionFeatureRows(testPriceActionCandles())
	summary := BuildPriceActionSummary(rows)

	dir := t.TempDir()
	csvPath := filepath.Join(dir, "price_action_features.csv")
	jsonPath := filepath.Join(dir, "price_action_summary.json")

	if err := WritePriceActionFeaturesCSV(csvPath, rows); err != nil {
		t.Fatalf("write csv: %v", err)
	}
	body, err := os.ReadFile(csvPath)
	if err != nil {
		t.Fatalf("read csv: %v", err)
	}
	text := string(body)
	if !strings.Contains(text, "timestamp,open,high,low,close,volume") {
		t.Fatalf("missing csv header: %s", text)
	}
	if !strings.Contains(text, "failed_auction_detected") {
		t.Fatalf("missing failed auction column: %s", text)
	}

	if err := WritePriceActionSummaryJSON(jsonPath, summary); err != nil {
		t.Fatalf("write json: %v", err)
	}
	jsonBody, err := os.ReadFile(jsonPath)
	if err != nil {
		t.Fatalf("read json: %v", err)
	}
	var decoded PriceActionSummary
	if err := json.Unmarshal(jsonBody, &decoded); err != nil {
		t.Fatalf("invalid json: %v", err)
	}
	if decoded.Candles != len(rows) {
		t.Fatalf("summary candles=%d want %d", decoded.Candles, len(rows))
	}
}

func testPriceActionCandles() []exchanges.Candle {
	return []exchanges.Candle{
		{Open: "100", High: "100.2", Low: "99.8", Close: "100", Volume: "10", StartTime: 1},
		{Open: "100", High: "100.3", Low: "99.9", Close: "100.1", Volume: "10", StartTime: 2},
		{Open: "100.1", High: "100.2", Low: "99.8", Close: "100", Volume: "10", StartTime: 3},
		{Open: "100", High: "100.3", Low: "99.9", Close: "100.1", Volume: "10", StartTime: 4},
		{Open: "100.1", High: "103", Low: "100", Close: "102.8", Volume: "50", StartTime: 5},
		{Open: "102.8", High: "104", Low: "101", Close: "103", Volume: "20", StartTime: 6},
		{Open: "103", High: "110", Low: "102", Close: "103", Volume: "30", StartTime: 7},
		{Open: "103", High: "104", Low: "95", Close: "103", Volume: "30", StartTime: 8},
		{Open: "103", High: "105", Low: "100", Close: "104", Volume: "20", StartTime: 9},
		{Open: "104", High: "106", Low: "103", Close: "105", Volume: "20", StartTime: 10},
	}
}
