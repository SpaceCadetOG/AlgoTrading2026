package research

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"AlgoTrading2026/volumeprofile"
)

func TestScopedVolumeProfileWriters(t *testing.T) {
	scoped := []volumeprofile.ScopedProfile{{
		Scope: volumeprofile.ScopeDailySession,
		Profile: volumeprofile.VolumeProfile{
			StartTime:   1,
			EndTime:     2,
			Venue:       "aster",
			Symbol:      "BTCUSDT",
			Bins:        []volumeprofile.PriceBin{{Low: 100, High: 110, Mid: 105, Volume: 10}},
			POC:         105,
			VAH:         110,
			VAL:         100,
			TotalVolume: 10,
			Shape:       volumeprofile.ShapeDProfile,
		},
	}}
	rows := BuildScopedVolumeProfileFeatureRows(scoped)
	if len(rows) != 1 || rows[0].Scope != volumeprofile.ScopeDailySession {
		t.Fatalf("rows=%+v", rows)
	}

	dir := t.TempDir()
	csvPath := filepath.Join(dir, "volume_profile_scoped_features.csv")
	jsonPath := filepath.Join(dir, "volume_profile_scoped_summary.json")
	if err := WriteScopedVolumeProfileFeaturesCSV(csvPath, rows); err != nil {
		t.Fatalf("write csv: %v", err)
	}
	body, err := os.ReadFile(csvPath)
	if err != nil {
		t.Fatalf("read csv: %v", err)
	}
	text := string(body)
	if !strings.Contains(text, "scope,timestamp,venue,symbol,profile_start,profile_end,bin_low,bin_high,bin_mid,bin_volume,poc,vah,val,is_poc,is_hvn,is_lvn,profile_shape,total_volume,notes") {
		t.Fatalf("missing header: %s", text)
	}
	if !strings.Contains(text, volumeprofile.ScopeDailySession) {
		t.Fatalf("missing scope row: %s", text)
	}

	summary := BuildScopedVolumeProfileSummary(scoped)
	if len(summary.Scopes) != 1 || summary.Scopes[0].Profiles != 1 {
		t.Fatalf("summary=%+v", summary)
	}
	if err := WriteScopedVolumeProfileSummaryJSON(jsonPath, summary); err != nil {
		t.Fatalf("write json: %v", err)
	}
	jsonBody, err := os.ReadFile(jsonPath)
	if err != nil {
		t.Fatalf("read json: %v", err)
	}
	var decoded ScopedVolumeProfileSummary
	if err := json.Unmarshal(jsonBody, &decoded); err != nil {
		t.Fatalf("invalid json: %v", err)
	}
	if len(decoded.Scopes) != 1 || decoded.Scopes[0].Scope != volumeprofile.ScopeDailySession {
		t.Fatalf("decoded=%+v", decoded)
	}
}
