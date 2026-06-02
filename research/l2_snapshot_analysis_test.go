package research

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"AlgoTrading2026/l2recorder"
)

func TestL2SnapshotAnalysisWriters(t *testing.T) {
	analysis := l2recorder.SnapshotAnalysis{
		DatasetQuality: l2recorder.DatasetQuality{
			TotalRows: 3,
			ValidRows: 3,
			ValidPct:  100,
		},
		VenueSummaries: []l2recorder.VenueSummary{
			{Venue: "aster", SnapshotCount: 1, ValidSnapshotCount: 1, AverageSpreadPct: 0.01},
			{Venue: "hyperliquid", SnapshotCount: 1, ValidSnapshotCount: 1, AverageSpreadPct: 0.02},
		},
		CrossVenue: l2recorder.CrossVenueSummary{
			TightestSpreadVenue:   "aster",
			WidestSpreadVenue:     "hyperliquid",
			DeepestLiquidityVenue: "hyperliquid",
		},
		LiquidityRanking: l2recorder.LiquidityRankings{
			TightestSpreads: []l2recorder.VenueRanking{{Venue: "aster", Value: 0.01}},
		},
		Notes: []string{"test note"},
	}

	dir := t.TempDir()
	jsonPath := filepath.Join(dir, "l2_snapshot_analysis.json")
	mdPath := filepath.Join(dir, "l2_snapshot_analysis.md")

	if err := WriteL2SnapshotAnalysisJSON(jsonPath, analysis); err != nil {
		t.Fatalf("write json: %v", err)
	}
	body, err := os.ReadFile(jsonPath)
	if err != nil {
		t.Fatalf("read json: %v", err)
	}
	var decoded l2recorder.SnapshotAnalysis
	if err := json.Unmarshal(body, &decoded); err != nil {
		t.Fatalf("invalid json: %v", err)
	}
	if decoded.DatasetQuality.ValidRows != 3 {
		t.Fatalf("decoded=%+v", decoded.DatasetQuality)
	}

	if err := WriteL2SnapshotAnalysisMarkdown(mdPath, analysis); err != nil {
		t.Fatalf("write markdown: %v", err)
	}
	md, err := os.ReadFile(mdPath)
	if err != nil {
		t.Fatalf("read markdown: %v", err)
	}
	text := string(md)
	if !strings.Contains(text, "# Historical L2 Snapshot Analysis") {
		t.Fatalf("missing title: %s", text)
	}
	if !strings.Contains(text, "Tightest spread venue: aster") {
		t.Fatalf("missing finding: %s", text)
	}
}
