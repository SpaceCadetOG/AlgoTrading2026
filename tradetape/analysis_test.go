package tradetape

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestAnalyzeRowsAggregatesMetrics(t *testing.T) {
	rows := []TradeTapeRow{
		RowFromPrint(testPrint("hyperliquid", "BTC", 10, 100, 1, "B")),
		RowFromPrint(testPrint("aster", "BTCUSDT", 20, 101, 2, "A")),
		RowFromPrint(TradeTapePrint{Venue: "lighter", Symbol: "BTC", Timestamp: 30, Price: 102, Size: 3, AggressorSide: UnknownAggressorSide}),
		ErrorRow("lighter", "BTC", os.ErrNotExist),
	}
	analysis := AnalyzeRows(rows)
	if analysis.Rows != 4 || analysis.ValidRows != 3 || analysis.ErrorRows != 1 {
		t.Fatalf("unexpected row counts: %+v", analysis)
	}
	if analysis.Venues != 3 || analysis.VenueSymbols != 2 || analysis.CanonicalSymbols != 1 {
		t.Fatalf("unexpected venue/symbol counts: %+v", analysis)
	}
	if analysis.TotalVolume != 6 || analysis.LargestTrade != 3 || analysis.AverageTradeSize != 2 {
		t.Fatalf("unexpected volume stats: %+v", analysis)
	}
	if analysis.BuyTrades != 1 || analysis.SellTrades != 1 || analysis.UnknownSideTrades != 1 {
		t.Fatalf("unexpected side counts: %+v", analysis)
	}
	if analysis.FirstTimestamp != 10 || analysis.LastTimestamp != 30 {
		t.Fatalf("unexpected timestamps: %+v", analysis)
	}
}

func TestAnalysisWriters(t *testing.T) {
	analysis := AnalyzeRows([]TradeTapeRow{RowFromPrint(testPrint("hyperliquid", "BTC", 1, 100, 1, "B"))})
	dir := t.TempDir()
	jsonPath := filepath.Join(dir, "trade_tape_analysis.json")
	mdPath := filepath.Join(dir, "trade_tape_analysis.md")
	if err := WriteAnalysisJSON(jsonPath, analysis); err != nil {
		t.Fatalf("write json: %v", err)
	}
	if err := WriteAnalysisMarkdown(mdPath, analysis); err != nil {
		t.Fatalf("write markdown: %v", err)
	}
	body, err := os.ReadFile(jsonPath)
	if err != nil {
		t.Fatalf("read json: %v", err)
	}
	var decoded TradeTapeAnalysis
	if err := json.Unmarshal(body, &decoded); err != nil {
		t.Fatalf("decode json: %v", err)
	}
	if decoded.ValidRows != 1 {
		t.Fatalf("unexpected decoded analysis: %+v", decoded)
	}
	mdBody, err := os.ReadFile(mdPath)
	if err != nil {
		t.Fatalf("read markdown: %v", err)
	}
	if !strings.Contains(string(mdBody), "# Trade Tape Analysis") {
		t.Fatalf("missing markdown heading: %s", string(mdBody))
	}
}
