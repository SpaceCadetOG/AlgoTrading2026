package research

import (
	"os"
	"strings"
	"testing"

	"AlgoTrading2026/realism"
)

func TestChapter10RealismModelsReport(t *testing.T) {
	report := BuildChapter10RealismModelsReport(realism.DefaultAggregateRealismModel())
	if report.Model.OverallRisk != realism.SeverityHigh {
		t.Fatalf("expected default report to be high risk, got %s", report.Model.OverallRisk)
	}
	if len(report.Conclusion) == 0 {
		t.Fatal("expected conclusion")
	}
}

func TestChapter10RealismModelsWriters(t *testing.T) {
	report := BuildChapter10RealismModelsReport(realism.DefaultAggregateRealismModel())
	dir := t.TempDir()
	jsonPath := dir + "/chapter10_realism_models.json"
	mdPath := dir + "/chapter10_realism_models.md"

	if err := WriteChapter10RealismModelsJSON(jsonPath, report); err != nil {
		t.Fatalf("write json: %v", err)
	}
	if err := WriteChapter10RealismModelsMarkdown(mdPath, report); err != nil {
		t.Fatalf("write markdown: %v", err)
	}

	body, err := os.ReadFile(mdPath)
	if err != nil {
		t.Fatalf("read markdown: %v", err)
	}
	text := strings.ToLower(string(body))
	for _, want := range []string{"latency model", "place-in-line", "market impact", "fill assumptions", "no live or paper trading"} {
		if !strings.Contains(text, want) {
			t.Fatalf("expected markdown to contain %q", want)
		}
	}
}
