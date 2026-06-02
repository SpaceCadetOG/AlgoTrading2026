package research

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"AlgoTrading2026/exchanges"
	"AlgoTrading2026/vwap"
)

func TestClassifyTrendAlignment(t *testing.T) {
	if got := ClassifyTrendAlignment(105, 100, 104, 102, 1); got != TrendStrongBull {
		t.Fatalf("got %s want strong bull", got)
	}
	if got := ClassifyTrendAlignment(95, 100, 96, 98, -1); got != TrendStrongBear {
		t.Fatalf("got %s want strong bear", got)
	}
	if got := ClassifyTrendAlignment(101, 100, 99, 102, 0); got != TrendNeutral {
		t.Fatalf("got %s want neutral", got)
	}
}

func TestBuildContextFeatureRows(t *testing.T) {
	candles := testVWAPChapter3Candles()
	rows := BuildContextFeatureRows(candles)
	if len(rows) != len(candles) {
		t.Fatalf("rows=%d want %d", len(rows), len(candles))
	}
	if rows[0].EMAAlignment != vwap.EMAAlignmentMixed {
		t.Fatalf("first alignment=%s want mixed", rows[0].EMAAlignment)
	}
	if rows[len(rows)-1].EMA9 == 0 || rows[len(rows)-1].EMA20 == 0 {
		t.Fatalf("expected EMA values: %+v", rows[len(rows)-1])
	}
}

func TestBuildTimeframeFeatureRows(t *testing.T) {
	rows := BuildTimeframeFeatureRows(map[string][]exchanges.Candle{
		"1m": testVWAPChapter3Candles(),
		"3m": testVWAPChapter3Candles(),
	})
	if len(rows) != 2*len(testVWAPChapter3Candles()) {
		t.Fatalf("rows=%d", len(rows))
	}
	if rows[0].Timeframe != "1m" {
		t.Fatalf("first timeframe=%s want 1m", rows[0].Timeframe)
	}
}

func TestBuildVWAPChapter3Summary(t *testing.T) {
	rows := []ContextFeatureRow{
		{EMAAlignment: vwap.EMAAlignmentBullish, TrendAlignment: TrendStrongBull, DistanceFromVWAP: 2, VWAPSlope: 1},
		{EMAAlignment: vwap.EMAAlignmentBearish, TrendAlignment: TrendStrongBear, DistanceFromVWAP: -4, VWAPSlope: -1},
		{EMAAlignment: vwap.EMAAlignmentMixed, TrendAlignment: TrendNeutral, DistanceFromVWAP: 0, VWAPSlope: 0},
	}
	summary := BuildVWAPChapter3Summary(rows)
	if summary.TotalCandles != 3 || summary.BullAlignmentCount != 1 || summary.BearAlignmentCount != 1 || summary.MixedCount != 1 {
		t.Fatalf("unexpected summary: %+v", summary)
	}
	if summary.StrongBullCount != 1 || summary.StrongBearCount != 1 {
		t.Fatalf("unexpected trend summary: %+v", summary)
	}
	if summary.AverageDistanceFromVWAP != 2 {
		t.Fatalf("avg distance %.4f want 2", summary.AverageDistanceFromVWAP)
	}
}

func TestVWAPChapter3Writers(t *testing.T) {
	dir := t.TempDir()
	contextPath := filepath.Join(dir, "context_features.csv")
	timeframePath := filepath.Join(dir, "timeframe_features.csv")
	trendPath := filepath.Join(dir, "trend_alignment.csv")
	summaryPath := filepath.Join(dir, "chapter3_summary.json")

	contextRows := BuildContextFeatureRows(testVWAPChapter3Candles())
	if err := WriteContextFeaturesCSV(contextPath, contextRows); err != nil {
		t.Fatalf("write context: %v", err)
	}
	timeframeRows := BuildTimeframeFeatureRows(map[string][]exchanges.Candle{"1m": testVWAPChapter3Candles()})
	if err := WriteTimeframeFeaturesCSV(timeframePath, timeframeRows); err != nil {
		t.Fatalf("write timeframe: %v", err)
	}
	trendRows := BuildTrendAlignmentRows(testVWAPChapter3Candles())
	if err := WriteTrendAlignmentCSV(trendPath, trendRows); err != nil {
		t.Fatalf("write trend: %v", err)
	}
	if err := WriteVWAPChapter3SummaryJSON(summaryPath, BuildVWAPChapter3Summary(contextRows)); err != nil {
		t.Fatalf("write summary: %v", err)
	}

	assertFileContains(t, contextPath, "timestamp,close,session_vwap,distance_from_vwap,ema9,ema20")
	assertFileContains(t, timeframePath, "timeframe,timestamp,close,ema9,ema20,vwap,slope,alignment,trend")
	assertFileContains(t, trendPath, "timestamp,close,session_vwap,ema9,ema20,vwap_slope,trend_alignment")
	assertFileContains(t, summaryPath, `"totalCandles"`)
}

func testVWAPChapter3Candles() []exchanges.Candle {
	return []exchanges.Candle{
		{Open: "100", High: "101", Low: "99", Close: "100", Volume: "10", StartTime: 1},
		{Open: "101", High: "102", Low: "100", Close: "101", Volume: "11", StartTime: 2},
		{Open: "102", High: "103", Low: "101", Close: "102", Volume: "12", StartTime: 3},
		{Open: "103", High: "104", Low: "102", Close: "103", Volume: "13", StartTime: 4},
		{Open: "104", High: "105", Low: "103", Close: "104", Volume: "14", StartTime: 5},
	}
}

func assertFileContains(t *testing.T, path string, needle string) {
	t.Helper()
	body, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	if !strings.Contains(string(body), needle) {
		t.Fatalf("%s missing %q: %s", path, needle, string(body))
	}
}
