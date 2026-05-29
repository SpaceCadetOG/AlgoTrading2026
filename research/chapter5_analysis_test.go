package research

import (
	"os"
	"strings"
	"testing"
)

func TestWriteChapter5AnalysisCSV(t *testing.T) {
	path := t.TempDir() + "/chapter5_analysis.csv"
	rows := []Chapter5Analysis{{
		Strategy:          "vol_mean_reversion",
		Venue:             "aster",
		Symbol:            "BTCUSDT",
		Interval:          "15m",
		RegimeNormalCount: 1,
	}}

	if err := WriteChapter5AnalysisCSV(path, rows); err != nil {
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
	if !strings.Contains(text, "vol_mean_reversion,aster,BTCUSDT,15m") {
		t.Fatalf("missing row: %s", text)
	}
}
