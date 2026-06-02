package research

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"AlgoTrading2026/priceaction"
)

func TestBuildPriceActionPhase3Summary(t *testing.T) {
	rows := []PriceActionPhase3StudyRow{
		{Strategy: "daily_high_retest", Direction: priceaction.Phase3LongContext},
		{Strategy: "weekly_low_retest", Direction: priceaction.Phase3ShortContext, Invalidated: true},
		{Strategy: "strong_high", Direction: priceaction.Phase3ShortContext},
		{Strategy: "weak_low", Direction: priceaction.Phase3BreakoutWatch},
		{Strategy: "failed_low_auction", Direction: priceaction.Phase3LongContext},
	}
	summary := BuildPriceActionPhase3Summary(25, rows)
	if summary.Candles != 25 || summary.DailyHighRetests != 1 || summary.WeeklyLowRetests != 1 || summary.StrongHighs != 1 || summary.WeakLows != 1 || summary.FailedLowAuctions != 1 {
		t.Fatalf("bad setup counts: %+v", summary)
	}
	if summary.LongContexts != 2 || summary.ShortContexts != 2 || summary.BreakoutWatchContexts != 1 || summary.Invalidated != 1 {
		t.Fatalf("bad context counts: %+v", summary)
	}
}

func TestPriceActionPhase3Writers(t *testing.T) {
	rows := []PriceActionPhase3StudyRow{{
		Timestamp:       1,
		Strategy:        "daily_high_retest",
		Direction:       priceaction.Phase3LongContext,
		Level:           100,
		LevelType:       "previous_daily_high",
		Confirmation:    true,
		FollowThrough5:  1,
		FollowThrough10: 2,
		FollowThrough20: 3,
		Notes:           "test",
	}}
	summary := BuildPriceActionPhase3Summary(1, rows)

	dir := t.TempDir()
	csvPath := filepath.Join(dir, "price_action_phase3_study.csv")
	jsonPath := filepath.Join(dir, "price_action_phase3_summary.json")

	if err := WritePriceActionPhase3StudyCSV(csvPath, rows); err != nil {
		t.Fatalf("write csv: %v", err)
	}
	body, err := os.ReadFile(csvPath)
	if err != nil {
		t.Fatalf("read csv: %v", err)
	}
	text := string(body)
	if !strings.Contains(text, "timestamp,strategy,direction,level,level_type,confirmation,invalidated,follow_through_5,follow_through_10,follow_through_20,notes") {
		t.Fatalf("missing header: %s", text)
	}
	if !strings.Contains(text, "daily_high_retest") {
		t.Fatalf("missing row: %s", text)
	}

	if err := WritePriceActionPhase3SummaryJSON(jsonPath, summary); err != nil {
		t.Fatalf("write json: %v", err)
	}
	jsonBody, err := os.ReadFile(jsonPath)
	if err != nil {
		t.Fatalf("read json: %v", err)
	}
	var decoded PriceActionPhase3Summary
	if err := json.Unmarshal(jsonBody, &decoded); err != nil {
		t.Fatalf("invalid json: %v", err)
	}
	if decoded.DailyHighRetests != 1 {
		t.Fatalf("decoded=%+v", decoded)
	}
}
