package research

import (
	"os"
	"strings"
	"testing"

	"AlgoTrading2026/backtest"
)

func TestChapter9BacktesterComparisonReportWriters(t *testing.T) {
	result := backtest.BacktesterComparisonResult{
		Symbol:                "BTCUSDT",
		Candles:               10,
		ForLoopTrades:         2,
		EventDrivenOrders:     4,
		EventDrivenFills:      4,
		ForLoopFinalEquity:    10001,
		EventDrivenFinalPnL:   2,
		PnLDifference:         1,
		AssumptionsDifference: "different assumptions",
		Recommendation:        "compare honestly",
		AssumptionComparisons: []string{"for-loop direct fills", "event-driven queue fills"},
	}
	report := BuildChapter9BacktesterComparisonReport(result)
	dir := t.TempDir()
	jsonPath := dir + "/chapter9_backtester_comparison.json"
	mdPath := dir + "/chapter9_backtester_comparison.md"

	if err := WriteChapter9BacktesterComparisonJSON(jsonPath, report); err != nil {
		t.Fatalf("write json: %v", err)
	}
	if err := WriteChapter9BacktesterComparisonMarkdown(mdPath, report); err != nil {
		t.Fatalf("write markdown: %v", err)
	}

	body, err := os.ReadFile(mdPath)
	if err != nil {
		t.Fatalf("read markdown: %v", err)
	}
	text := strings.ToLower(string(body))
	for _, want := range []string{"for-loop", "event-driven", "pnl_difference", "recommendation"} {
		if !strings.Contains(text, want) {
			t.Fatalf("expected markdown to contain %q", want)
		}
	}
}
