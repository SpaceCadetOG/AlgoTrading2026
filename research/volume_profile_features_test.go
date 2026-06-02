package research

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"AlgoTrading2026/volumeprofile"
)

func TestVolumeProfileFeatureWriter(t *testing.T) {
	profile := volumeprofile.VolumeProfile{
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
	}
	rows := BuildVolumeProfileFeatureRows(profile)
	if len(rows) != 1 || !rows[0].IsPOC {
		t.Fatalf("rows=%+v", rows)
	}

	path := filepath.Join(t.TempDir(), "volume_profile_features.csv")
	if err := WriteVolumeProfileFeaturesCSV(path, rows); err != nil {
		t.Fatalf("write csv: %v", err)
	}
	body, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read csv: %v", err)
	}
	text := string(body)
	if !strings.Contains(text, "timestamp,venue,symbol,profile_start,profile_end,bin_low,bin_high,bin_mid,bin_volume,poc,vah,val,is_poc,is_hvn,is_lvn,profile_shape,total_volume,notes") {
		t.Fatalf("missing header: %s", text)
	}
	if !strings.Contains(text, "D_PROFILE") {
		t.Fatalf("missing row: %s", text)
	}
}

func TestVolumeProfileSummaryWriter(t *testing.T) {
	profile := volumeprofile.VolumeProfile{
		Bins:        []volumeprofile.PriceBin{{Volume: 10}},
		POC:         105,
		VAH:         110,
		VAL:         100,
		TotalVolume: 10,
		HVNs:        []volumeprofile.PriceBin{{Volume: 10}},
		LVNs:        []volumeprofile.PriceBin{{Volume: 1}},
		Shape:       volumeprofile.ShapePProfile,
	}
	summary := BuildVolumeProfileSummary([]volumeprofile.VolumeProfile{profile})
	if summary.Profiles != 1 || summary.Bins != 1 || summary.PProfiles != 1 || summary.HVNs != 1 || summary.LVNs != 1 {
		t.Fatalf("summary=%+v", summary)
	}

	path := filepath.Join(t.TempDir(), "volume_profile_summary.json")
	if err := WriteVolumeProfileSummaryJSON(path, summary); err != nil {
		t.Fatalf("write json: %v", err)
	}
	body, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read json: %v", err)
	}
	var decoded VolumeProfileSummary
	if err := json.Unmarshal(body, &decoded); err != nil {
		t.Fatalf("invalid json: %v", err)
	}
	if decoded.PProfiles != 1 || len(decoded.Notes) == 0 {
		t.Fatalf("decoded=%+v", decoded)
	}
}
