package research

import (
	"os"
	"strings"
	"testing"
	"time"

	"AlgoTrading2026/backtest"
)

func TestChapter9EventDrivenReportWriters(t *testing.T) {
	report := BuildChapter9EventDrivenReport(backtest.EventDrivenBacktestResult{
		CandlesProcessed: 2,
		OrdersCreated:    4,
		OrdersFilled:     4,
		FinalCash:        10001,
		FinalPnL:         1,
		ClockStart:       time.UnixMilli(1000).UTC(),
		ClockEnd:         time.UnixMilli(2000).UTC(),
		AuditEvents:      8,
	})
	dir := t.TempDir()
	jsonPath := dir + "/chapter9_event_driven.json"
	mdPath := dir + "/chapter9_event_driven.md"

	if err := WriteChapter9EventDrivenJSON(jsonPath, report); err != nil {
		t.Fatalf("write json: %v", err)
	}
	if err := WriteChapter9EventDrivenMarkdown(mdPath, report); err != nil {
		t.Fatalf("write markdown: %v", err)
	}

	body, err := os.ReadFile(mdPath)
	if err != nil {
		t.Fatalf("read markdown: %v", err)
	}
	text := strings.ToLower(string(body))
	for _, want := range []string{"event-driven", "lp_2_gateway", "allowliveorders is forced to false"} {
		if !strings.Contains(text, strings.ToLower(want)) {
			t.Fatalf("expected markdown to contain %q", want)
		}
	}
}
