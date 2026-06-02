package research

import (
	"os"
	"strings"
	"testing"
	"time"

	"AlgoTrading2026/backtest"
)

func TestChapter9PacketContainsRequiredConclusion(t *testing.T) {
	packet := testChapter9Packet()
	required := []string{
		"Chapter 9 backtesting layer is complete.",
		"For-loop backtester is for fast signal/strategy research.",
		"Event-driven backtester is for OMS/system/market-simulator validation.",
		"Results are not expected to match because assumptions differ.",
		"No live or paper trading is enabled.",
		"Ready to proceed to Chapter 10 market adaptation and realism.",
	}

	for _, want := range required {
		found := false
		for _, got := range packet.Conclusion {
			if got == want {
				found = true
				break
			}
		}
		if !found {
			t.Fatalf("missing conclusion %q in %+v", want, packet.Conclusion)
		}
	}
}

func TestChapter9PacketWriters(t *testing.T) {
	packet := testChapter9Packet()
	dir := t.TempDir()
	jsonPath := dir + "/chapter9_packet.json"
	mdPath := dir + "/chapter9_packet.md"

	if err := WriteChapter9PacketJSON(jsonPath, packet); err != nil {
		t.Fatalf("write json: %v", err)
	}
	if err := WriteChapter9PacketMarkdown(mdPath, packet); err != nil {
		t.Fatalf("write markdown: %v", err)
	}

	body, err := os.ReadFile(mdPath)
	if err != nil {
		t.Fatalf("read markdown: %v", err)
	}
	text := strings.ToLower(string(body))
	for _, want := range []string{"for-loop", "event-driven", "readiness for chapter 10", "no live or paper trading"} {
		if !strings.Contains(text, strings.ToLower(want)) {
			t.Fatalf("expected markdown to contain %q", want)
		}
	}
}

func testChapter9Packet() Chapter9Packet {
	assumptions := backtest.DefaultBacktestAssumptions()
	audit := BuildChapter9BacktestAudit(80, 20, assumptions)
	timeReport := BuildChapter9TimeModelReport(100)
	eventReport := BuildChapter9EventDrivenReport(backtest.EventDrivenBacktestResult{
		CandlesProcessed: 10,
		OrdersCreated:    2,
		OrdersFilled:     2,
		FinalPnL:         1,
		ClockStart:       time.UnixMilli(1000).UTC(),
		ClockEnd:         time.UnixMilli(2000).UTC(),
	})
	comparison := BuildChapter9BacktesterComparisonReport(backtest.BacktesterComparisonResult{
		Symbol:                "BTCUSDT",
		Candles:               10,
		ForLoopTrades:         1,
		EventDrivenOrders:     2,
		EventDrivenFills:      2,
		PnLDifference:         1,
		AssumptionsDifference: "different assumptions",
		Recommendation:        "compare honestly",
	})
	return BuildChapter9Packet(audit, timeReport, eventReport, comparison)
}
