package research

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestInstrumentUniverseWriters(t *testing.T) {
	rows := BuildInstrumentUniverse([]OrderBookFeatureRow{{Venue: "aster", Symbol: "BTCUSDT", Mid: 100, SpreadPct: 0.0001, Valid: true}}, map[string]int{"aster:BTCUSDT": 100})
	if len(rows) == 0 || rows[0].Rank != 1 {
		t.Fatalf("rows=%+v", rows)
	}
	summary := BuildInstrumentUniverseSummary(rows)
	if summary.Venues == 0 || summary.Symbols == 0 || summary.TopInstrument == "" {
		t.Fatalf("summary=%+v", summary)
	}
	dir := t.TempDir()
	csvPath := filepath.Join(dir, "instrument_universe.csv")
	jsonPath := filepath.Join(dir, "instrument_universe_summary.json")
	if err := WriteInstrumentUniverseCSV(csvPath, rows); err != nil {
		t.Fatalf("write csv: %v", err)
	}
	body, err := os.ReadFile(csvPath)
	if err != nil {
		t.Fatalf("read csv: %v", err)
	}
	if !strings.Contains(string(body), "venue,symbol,asset_type") {
		t.Fatalf("missing header: %s", string(body))
	}
	if err := WriteInstrumentUniverseSummaryJSON(jsonPath, summary); err != nil {
		t.Fatalf("write json: %v", err)
	}
	jsonBody, err := os.ReadFile(jsonPath)
	if err != nil {
		t.Fatalf("read json: %v", err)
	}
	var decoded InstrumentUniverseSummary
	if err := json.Unmarshal(jsonBody, &decoded); err != nil {
		t.Fatalf("invalid json: %v", err)
	}
	if decoded.Symbols != summary.Symbols {
		t.Fatalf("decoded=%+v summary=%+v", decoded, summary)
	}
}
