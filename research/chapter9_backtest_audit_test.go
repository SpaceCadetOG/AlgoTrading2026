package research

import (
	"os"
	"strings"
	"testing"

	"AlgoTrading2026/backtest"
)

func TestChapter9BacktestAuditWriters(t *testing.T) {
	audit := BuildChapter9BacktestAudit(80, 20, backtest.DefaultBacktestAssumptions())
	dir := t.TempDir()
	jsonPath := dir + "/chapter9_backtest_audit.json"
	mdPath := dir + "/chapter9_backtest_audit.md"

	if err := WriteChapter9BacktestAuditJSON(jsonPath, audit); err != nil {
		t.Fatalf("write json: %v", err)
	}
	if err := WriteChapter9BacktestAuditMarkdown(mdPath, audit); err != nil {
		t.Fatalf("write markdown: %v", err)
	}

	body, err := os.ReadFile(mdPath)
	if err != nil {
		t.Fatalf("read markdown: %v", err)
	}
	text := strings.ToLower(string(body))
	for _, want := range []string{"in_sample_count", "for-loop", "event-driven backtester is not yet complete"} {
		if !strings.Contains(text, want) {
			t.Fatalf("expected markdown to contain %q", want)
		}
	}
}
