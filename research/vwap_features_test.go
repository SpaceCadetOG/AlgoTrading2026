package research

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"AlgoTrading2026/exchanges"
)

func TestBuildVWAPFeatureRows(t *testing.T) {
	candles := []exchanges.Candle{
		{Open: "10", High: "10", Low: "10", Close: "10", Volume: "2", StartTime: 1},
		{Open: "20", High: "20", Low: "20", Close: "20", Volume: "1", StartTime: 2},
	}
	rows := BuildVWAPFeatureRows(candles, 1, 1)
	if len(rows) != 2 {
		t.Fatalf("rows=%d want 2", len(rows))
	}
	if rows[0].SessionVWAP != 10 {
		t.Fatalf("session vwap %.4f want 10", rows[0].SessionVWAP)
	}
	if rows[1].DistanceFromVWAP == 0 {
		t.Fatal("expected non-zero distance")
	}
}

func TestWriteVWAPFeaturesCSV(t *testing.T) {
	rows := []VWAPFeatureRow{{
		Timestamp:    1,
		Close:        10,
		Volume:       2,
		SessionVWAP:  10,
		AnchoredVWAP: 10,
		VWAPRegime:   "neutral",
	}}
	path := filepath.Join(t.TempDir(), "vwap_features.csv")
	if err := WriteVWAPFeaturesCSV(path, rows); err != nil {
		t.Fatalf("write csv: %v", err)
	}
	body, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read csv: %v", err)
	}
	text := string(body)
	if !strings.Contains(text, "timestamp,close,volume,session_vwap,anchored_vwap") {
		t.Fatalf("missing header: %s", text)
	}
	if !strings.Contains(text, "neutral") {
		t.Fatalf("missing regime: %s", text)
	}
}
