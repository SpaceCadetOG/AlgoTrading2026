package research

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"AlgoTrading2026/exchanges"
	"AlgoTrading2026/realism"
)

func TestChapter10DataQualityReport(t *testing.T) {
	quality := realism.CheckMarketDataQuality([]exchanges.Candle{
		{
			Venue:     "aster",
			Symbol:    "BTCUSDT",
			Interval:  "15m",
			Open:      "100",
			High:      "101",
			Low:       "99",
			Close:     "100",
			Volume:    "10",
			StartTime: 0,
			EndTime:   900000,
			Closed:    true,
		},
	}, 900000)
	parity := realism.DefaultHistoricalLiveParityCheck("aster", "BTCUSDT", "15m")
	report := BuildChapter10DataQualityReport("aster", "BTCUSDT", "15m", realism.BuildDataQualityReport(quality, parity))

	if report.Title == "" {
		t.Fatal("expected title")
	}
	if report.DataQuality.TotalCandlesChecked != 1 {
		t.Fatalf("expected one checked candle, got %d", report.DataQuality.TotalCandlesChecked)
	}
	if len(report.Conclusion) == 0 || !strings.Contains(strings.Join(report.Conclusion, " "), "No live or paper trading") {
		t.Fatalf("expected safety conclusion, got %+v", report.Conclusion)
	}
}

func TestChapter10DataQualityWriters(t *testing.T) {
	dir := t.TempDir()
	quality := realism.CheckMarketDataQuality(nil, 900000)
	parity := realism.NewHistoricalLiveParityCheck("test", "BTC", "15m", nil)
	report := BuildChapter10DataQualityReport("test", "BTC", "15m", realism.BuildDataQualityReport(quality, parity))

	jsonPath := filepath.Join(dir, "chapter10_data_quality.json")
	if err := WriteChapter10DataQualityJSON(jsonPath, report); err != nil {
		t.Fatalf("write json: %v", err)
	}
	body, err := os.ReadFile(jsonPath)
	if err != nil {
		t.Fatalf("read json: %v", err)
	}
	var decoded Chapter10DataQualityReport
	if err := json.Unmarshal(body, &decoded); err != nil {
		t.Fatalf("json should decode: %v", err)
	}

	mdPath := filepath.Join(dir, "chapter10_data_quality.md")
	if err := WriteChapter10DataQualityMarkdown(mdPath, report); err != nil {
		t.Fatalf("write markdown: %v", err)
	}
	markdown, err := os.ReadFile(mdPath)
	if err != nil {
		t.Fatalf("read markdown: %v", err)
	}
	text := string(markdown)
	for _, expected := range []string{"Market Data Quality", "Historical/Live Parity", "No real exchange calls are made"} {
		if !strings.Contains(text, expected) {
			t.Fatalf("expected markdown to contain %q\n%s", expected, text)
		}
	}
}
