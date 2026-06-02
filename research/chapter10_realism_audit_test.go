package research

import (
	"os"
	"strings"
	"testing"

	"AlgoTrading2026/backtest"
	"AlgoTrading2026/realism"
)

func TestBuildChapter10RealismAudit(t *testing.T) {
	audit := BuildChapter10RealismAudit(backtest.BacktesterComparisonResult{
		Symbol:              "BTCUSDT",
		Candles:             25,
		EventDrivenFinalPnL: 2252.12,
		PnLDifference:       2252.38,
	})

	if audit.Dislocation.BiasDirection != realism.BiasOptimistic {
		t.Fatalf("expected optimistic dislocation, got %s", audit.Dislocation.BiasDirection)
	}
	if audit.EventDrivenClassification == "" {
		t.Fatal("expected event-driven classification")
	}
	if audit.Grade != realism.GradeF {
		t.Fatalf("expected current realism grade F, got %s", audit.Grade)
	}
}

func TestChapter10RealismAuditWriters(t *testing.T) {
	audit := BuildChapter10RealismAudit(backtest.BacktesterComparisonResult{
		Symbol:              "BTCUSDT",
		Candles:             25,
		EventDrivenFinalPnL: 2252.12,
		PnLDifference:       2252.38,
	})
	dir := t.TempDir()
	jsonPath := dir + "/chapter10_realism_audit.json"
	mdPath := dir + "/chapter10_realism_audit.md"

	if err := WriteChapter10RealismAuditJSON(jsonPath, audit); err != nil {
		t.Fatalf("write json: %v", err)
	}
	if err := WriteChapter10RealismAuditMarkdown(mdPath, audit); err != nil {
		t.Fatalf("write markdown: %v", err)
	}

	body, err := os.ReadFile(mdPath)
	if err != nil {
		t.Fatalf("read markdown: %v", err)
	}
	text := strings.ToLower(string(body))
	for _, want := range []string{"simulation dislocation", "system-validation only", "no live or paper trading", "place-in-line"} {
		if !strings.Contains(text, want) {
			t.Fatalf("expected markdown to contain %q", want)
		}
	}
}
