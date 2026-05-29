package datasets

import (
	"os"
	"strings"
	"testing"

	"AlgoTrading2026/features"
	"AlgoTrading2026/labels"
)

func TestBuildTrainingDatasetJoinsByTime(t *testing.T) {
	featureRows := []features.FeatureRow{
		{Time: 1, Close: 100},
		{Time: 2, Close: 101},
	}
	labelRows := []labels.LabelRow{
		{Time: 2, FutureReturn: 0.1, Direction: 1, TPSL: 1},
	}

	rows := BuildTrainingDataset(featureRows, labelRows)
	if len(rows) != 1 {
		t.Fatalf("rows = %d, want 1", len(rows))
	}
	if rows[0].Time != 2 || rows[0].Direction != 1 {
		t.Fatalf("unexpected row: %+v", rows[0])
	}
}

func TestTrainingDatasetCSVWriter(t *testing.T) {
	path := t.TempDir() + "/training.csv"
	rows := []TrainingRow{{
		FeatureRow: features.FeatureRow{Time: 1, Venue: "aster", Symbol: "BTCUSDT", Interval: "15m", Close: 100},
		Direction:  1,
		TPSL:       1,
	}}

	if err := WriteCSV(path, rows); err != nil {
		t.Fatal(err)
	}
	body, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	text := string(body)
	if !strings.Contains(text, "time,venue,symbol,interval") {
		t.Fatalf("missing header: %s", text)
	}
	if !strings.Contains(text, "future_return,direction,tpsl") {
		t.Fatalf("missing label columns: %s", text)
	}
}
