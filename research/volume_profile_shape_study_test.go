package research

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"AlgoTrading2026/volumeprofile"
)

func TestVolumeProfileShapeStudyWriters(t *testing.T) {
	scoped := []volumeprofile.ScopedProfile{{
		Scope: volumeprofile.ScopeDailySession,
		Profile: volumeprofile.VolumeProfile{
			StartTime:   1,
			EndTime:     2,
			Symbol:      "BTCUSDT",
			Bins:        []volumeprofile.PriceBin{{Low: 0, High: 10, Mid: 5, Volume: 10}, {Low: 10, High: 20, Mid: 15, Volume: 30}, {Low: 20, High: 30, Mid: 25, Volume: 10}},
			POC:         15,
			VAH:         20,
			VAL:         10,
			TotalVolume: 50,
			HVNs:        []volumeprofile.PriceBin{{Low: 10, High: 20, Mid: 15, Volume: 30}},
			LVNs:        nil,
		},
	}}
	rows := BuildVolumeProfileShapeStudyRows(scoped)
	if len(rows) != 1 || rows[0].Scope != volumeprofile.ScopeDailySession || rows[0].ShapeReason == "" {
		t.Fatalf("rows=%+v", rows)
	}

	dir := t.TempDir()
	csvPath := filepath.Join(dir, "volume_profile_shape_study.csv")
	jsonPath := filepath.Join(dir, "volume_profile_shape_summary.json")
	if err := WriteVolumeProfileShapeStudyCSV(csvPath, rows); err != nil {
		t.Fatalf("write csv: %v", err)
	}
	body, err := os.ReadFile(csvPath)
	if err != nil {
		t.Fatalf("read csv: %v", err)
	}
	text := string(body)
	if !strings.Contains(text, "scope,profile_start,profile_end,symbol,poc,vah,val,profile_shape,total_bins,total_volume,upper_volume_pct,lower_volume_pct,hvn_count,lvn_count,distribution_balance,shape_confidence,shape_reason") {
		t.Fatalf("missing header: %s", text)
	}

	summary := BuildVolumeProfileShapeSummary(rows)
	if summary.Daily.Shape == "" {
		t.Fatalf("summary=%+v", summary)
	}
	if err := WriteVolumeProfileShapeSummaryJSON(jsonPath, summary); err != nil {
		t.Fatalf("write json: %v", err)
	}
	jsonBody, err := os.ReadFile(jsonPath)
	if err != nil {
		t.Fatalf("read json: %v", err)
	}
	var decoded VolumeProfileShapeSummary
	if err := json.Unmarshal(jsonBody, &decoded); err != nil {
		t.Fatalf("invalid json: %v", err)
	}
	if decoded.Daily.Shape == "" {
		t.Fatalf("decoded=%+v", decoded)
	}
}
