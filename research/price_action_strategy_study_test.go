package research

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"AlgoTrading2026/priceaction"
)

func TestBuildPriceActionStrategySummary(t *testing.T) {
	rows := []PriceActionStrategyStudyRow{
		{Strategy: "sr_flip", Direction: priceaction.StudyLong},
		{Strategy: "open_drive", Direction: priceaction.StudyShort, Invalidated: true},
		{Strategy: "abcd", Direction: priceaction.StudyLong},
		{Strategy: "session_open", Direction: priceaction.StudyLong},
		{Strategy: "daily_open", Direction: priceaction.StudyShort},
	}
	summary := BuildPriceActionStrategySummary(10, rows)
	if summary.Candles != 10 || summary.SRFlipSetups != 1 || summary.OpenDriveSetups != 1 || summary.ABCDSetups != 1 {
		t.Fatalf("bad summary: %+v", summary)
	}
	if summary.LongContexts != 3 || summary.ShortContexts != 2 || summary.Invalidated != 1 {
		t.Fatalf("bad context counts: %+v", summary)
	}
}

func TestPriceActionStrategyWriters(t *testing.T) {
	rows := []PriceActionStrategyStudyRow{{
		Timestamp:       1,
		Strategy:        "sr_flip",
		Direction:       priceaction.StudyLong,
		Level:           100,
		EntryZone:       100,
		Confirmation:    true,
		FollowThrough5:  1,
		FollowThrough10: 2,
		FollowThrough20: 3,
		Notes:           "test",
	}}
	summary := BuildPriceActionStrategySummary(1, rows)

	dir := t.TempDir()
	csvPath := filepath.Join(dir, "price_action_strategy_study.csv")
	jsonPath := filepath.Join(dir, "price_action_strategy_summary.json")

	if err := WritePriceActionStrategyStudyCSV(csvPath, rows); err != nil {
		t.Fatalf("write csv: %v", err)
	}
	body, err := os.ReadFile(csvPath)
	if err != nil {
		t.Fatalf("read csv: %v", err)
	}
	text := string(body)
	if !strings.Contains(text, "timestamp,strategy,direction,level,entry_zone,confirmation,invalidated,follow_through_5,follow_through_10,follow_through_20,notes") {
		t.Fatalf("missing header: %s", text)
	}
	if !strings.Contains(text, "sr_flip") {
		t.Fatalf("missing row: %s", text)
	}

	if err := WritePriceActionStrategySummaryJSON(jsonPath, summary); err != nil {
		t.Fatalf("write json: %v", err)
	}
	jsonBody, err := os.ReadFile(jsonPath)
	if err != nil {
		t.Fatalf("read json: %v", err)
	}
	var decoded PriceActionStrategySummary
	if err := json.Unmarshal(jsonBody, &decoded); err != nil {
		t.Fatalf("invalid json: %v", err)
	}
	if decoded.SRFlipSetups != 1 {
		t.Fatalf("decoded=%+v", decoded)
	}
}
