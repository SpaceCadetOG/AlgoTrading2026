package research

import (
	"os"
	"strings"
	"testing"
)

func TestChapter9TimeModelReportWriters(t *testing.T) {
	report := BuildChapter9TimeModelReport(42)
	dir := t.TempDir()
	jsonPath := dir + "/chapter9_time_model.json"
	mdPath := dir + "/chapter9_time_model.md"

	if err := WriteChapter9TimeModelJSON(jsonPath, report); err != nil {
		t.Fatalf("write json: %v", err)
	}
	if err := WriteChapter9TimeModelMarkdown(mdPath, report); err != nil {
		t.Fatalf("write markdown: %v", err)
	}

	body, err := os.ReadFile(mdPath)
	if err != nil {
		t.Fatalf("read markdown: %v", err)
	}
	text := strings.ToLower(string(body))
	for _, want := range []string{"simulated clock", "wall-clock", "oms timeout", "event-driven backtester"} {
		if !strings.Contains(text, want) {
			t.Fatalf("expected markdown to contain %q", want)
		}
	}
}
