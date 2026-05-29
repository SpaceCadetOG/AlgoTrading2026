package research

import (
	"encoding/json"
	"os"
	"strings"
	"testing"
	"time"

	"AlgoTrading2026/backtest"
	"AlgoTrading2026/series"
	"AlgoTrading2026/strategy"
)

func TestChapter2AnalysisTradeCalculations(t *testing.T) {
	report := testAnalysisReport()
	trades := []strategy.Trade{
		{OpenedAt: time.UnixMilli(1000).UTC(), ClosedAt: time.UnixMilli(2000).UTC(), RealizedPnL: 1.5},
		{OpenedAt: time.UnixMilli(2000).UTC(), ClosedAt: time.UnixMilli(4000).UTC(), RealizedPnL: -0.5},
	}
	signals := series.SignalResult{
		Times:     []int64{0, 1000, 2000, 3000, 4000},
		Positions: []float64{0, 1, 0, -1, 0},
	}

	got := AnalyzeChapter2Strategy("test", report, trades, signals)

	if got.BestTradePnL != 1.5 {
		t.Fatalf("best = %f, want 1.5", got.BestTradePnL)
	}
	if got.WorstTradePnL != -0.5 {
		t.Fatalf("worst = %f, want -0.5", got.WorstTradePnL)
	}
	if got.AverageTradePnL != 0.5 {
		t.Fatalf("avg = %f, want 0.5", got.AverageTradePnL)
	}
	if got.AverageHoldCandles != 1.5 || got.MinHoldCandles != 1 || got.MaxHoldCandles != 2 {
		t.Fatalf("hold stats = avg %f min %d max %d", got.AverageHoldCandles, got.MinHoldCandles, got.MaxHoldCandles)
	}
}

func TestChapter2AnalysisSignalCountsAndNotes(t *testing.T) {
	report := testAnalysisReport()
	report.Metrics.TotalTrades = 12
	report.CandleCount = 100
	report.Metrics.NetPnL = -2.4
	trades := make([]strategy.Trade, 12)
	for i := range trades {
		trades[i] = strategy.Trade{RealizedPnL: -0.2}
	}
	signals := series.SignalResult{
		Times:     []int64{1, 2, 3, 4},
		Positions: []float64{1, 0, -1, 0},
	}

	got := AnalyzeChapter2Strategy("momentum", report, trades, signals)

	if got.SignalCount != 2 || got.BuySignalCount != 1 || got.SellSignalCount != 1 || got.HoldSignalCount != 2 {
		t.Fatalf("signal counts = %+v", got)
	}
	if got.OvertradeScore != 0.12 {
		t.Fatalf("overtrade = %f, want 0.12", got.OvertradeScore)
	}
	notes := strings.Join(got.Notes, ",")
	if !strings.Contains(notes, "overtrading") || !strings.Contains(notes, "negative_expectancy") {
		t.Fatalf("notes = %v", got.Notes)
	}
}

func TestWriteChapter2AnalysisCSV(t *testing.T) {
	path := t.TempDir() + "/analysis.csv"
	rows := []Chapter2Analysis{{
		Strategy: "rsi",
		Venue:    "aster",
		Symbol:   "BTCUSDT",
		Interval: "15m",
		Notes:    []string{"low_sample_size", "positive_expectancy_sample"},
	}}

	if err := WriteChapter2AnalysisCSV(path, rows); err != nil {
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
	if !strings.Contains(text, "rsi,aster,BTCUSDT,15m") {
		t.Fatalf("missing row: %s", text)
	}
}

func TestWriteChapter2AnalysisJSON(t *testing.T) {
	path := t.TempDir() + "/analysis.json"
	rows := []Chapter2Analysis{{Strategy: "rsi", Notes: []string{"low_sample_size"}}}

	if err := WriteChapter2AnalysisJSON(path, rows); err != nil {
		t.Fatal(err)
	}

	body, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}

	var parsed []Chapter2Analysis
	if err := json.Unmarshal(body, &parsed); err != nil {
		t.Fatal(err)
	}
	if len(parsed) != 1 || parsed[0].Strategy != "rsi" {
		t.Fatalf("unexpected json: %+v", parsed)
	}
}

func testAnalysisReport() backtest.Report {
	return backtest.Report{
		Venue:       "aster",
		Symbol:      "BTCUSDT",
		Interval:    "15m",
		CandleCount: 100,
		Metrics: backtest.Metrics{
			TotalTrades: 2,
			NetPnL:      1,
		},
	}
}
