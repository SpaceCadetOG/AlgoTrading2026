package research

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestVolumeSetupReversalQualityGrouping(t *testing.T) {
	records := []reversalQualityRecord{
		{Direction: "long_context", ProfileShape: "D_PROFILE", ProfileVolume: 1000, ShapeConfidence: 0.8, Accepted: true, POCConfluence: true, HVNConfluence: true, VAHVALRejection: true, VWAPAlignment: "above_vwap", FollowThrough20: 10},
		{Direction: "short_context", ProfileShape: "P_PROFILE", ProfileVolume: 500, ShapeConfidence: 0.6, Rejected: true, Invalidated: true, VWAPAlignment: "above_vwap", FollowThrough20: -2},
		{Direction: "long_context", ProfileShape: "D_PROFILE", ProfileVolume: 800, ShapeConfidence: 0.7, POCConfluence: true, VWAPAlignment: "below_vwap", FollowThrough20: 0},
	}
	rows := BuildVolumeSetupReversalQualityRows(records)
	summary := BuildVolumeSetupReversalQualitySummary(records, rows)
	if summary.TotalSetups != 3 || summary.Accepted != 1 || summary.Rejected != 1 || summary.Neutral != 1 {
		t.Fatalf("summary=%+v", summary)
	}
	if summary.BestDirection != "long_context" || summary.BestConfluence == "" || summary.BestFilter == "" {
		t.Fatalf("bad summary: %+v rows=%+v", summary, rows)
	}
	assertReversalQualityGroup(t, rows, "acceptance_state", "accepted")
	assertReversalQualityGroup(t, rows, "acceptance_state", "neutral")
	assertReversalQualityGroup(t, rows, "direction", "long_context")
	assertReversalQualityGroup(t, rows, "poc_confluence", "with_poc_confluence")
	assertReversalQualityGroup(t, rows, "hvn_confluence", "with_hvn_confluence")
	assertReversalQualityGroup(t, rows, "vah_val_rejection", "with_vah_val_rejection")
	assertReversalQualityGroup(t, rows, "vwap_alignment", "vwap_aligned")
}

func TestVolumeSetupReversalQualityFromFileAndWriters(t *testing.T) {
	dir := t.TempDir()
	inputPath := filepath.Join(dir, "volume_setup_reversal.csv")
	csvPath := filepath.Join(dir, "volume_setup_reversal_quality.csv")
	jsonPath := filepath.Join(dir, "volume_setup_reversal_quality_summary.json")
	mdPath := filepath.Join(dir, "volume_setup_reversal_quality.md")
	writeTestFile(t, inputPath, strings.Join([]string{
		"timestamp,setup_id,setup_type,direction,reversal_level,failed_auction_type,poc,vah,val,nearest_hvn,nearest_lvn,profile_shape,shape_confidence,profile_volume,accepted,rejected,vwap_alignment,poc_confluence,hvn_confluence,vah_val_rejection,invalidated,follow_through_5,follow_through_10,follow_through_20,notes",
		"1,r1,REVERSAL_TRADE,long_context,100,failed_low_auction,100,110,90,105,95,D_PROFILE,0.8,1000,true,false,above_vwap,true,true,true,false,1,2,3,test",
		"2,r2,REVERSAL_TRADE,short_context,110,failed_high_auction,110,112,100,0,95,P_PROFILE,0.6,500,false,true,above_vwap,false,false,false,true,-1,-2,-3,test",
	}, "\n"))
	rows, summary, err := BuildVolumeSetupReversalQualityFromFile(inputPath)
	if err != nil {
		t.Fatalf("build: %v", err)
	}
	if len(rows) == 0 || summary.TotalSetups != 2 || summary.Accepted != 1 || summary.Rejected != 1 {
		t.Fatalf("rows=%+v summary=%+v", rows, summary)
	}
	if err := WriteVolumeSetupReversalQualityCSV(csvPath, rows); err != nil {
		t.Fatalf("write csv: %v", err)
	}
	body, err := os.ReadFile(csvPath)
	if err != nil {
		t.Fatalf("read csv: %v", err)
	}
	if !strings.Contains(string(body), "group_type,group_name,setup_count,accepted_count,rejected_count") {
		t.Fatalf("missing csv header: %s", string(body))
	}
	if err := WriteVolumeSetupReversalQualitySummaryJSON(jsonPath, summary); err != nil {
		t.Fatalf("write json: %v", err)
	}
	jsonBody, err := os.ReadFile(jsonPath)
	if err != nil {
		t.Fatalf("read json: %v", err)
	}
	var decoded VolumeSetupReversalQualitySummary
	if err := json.Unmarshal(jsonBody, &decoded); err != nil {
		t.Fatalf("invalid json: %v", err)
	}
	if decoded.TotalSetups != 2 {
		t.Fatalf("decoded=%+v", decoded)
	}
	if err := WriteVolumeSetupReversalQualityMarkdown(mdPath, rows, summary); err != nil {
		t.Fatalf("write md: %v", err)
	}
	md, err := os.ReadFile(mdPath)
	if err != nil {
		t.Fatalf("read md: %v", err)
	}
	if !strings.Contains(string(md), "# Volume Setup Reversal Quality Review") {
		t.Fatalf("missing markdown title: %s", string(md))
	}
}

func assertReversalQualityGroup(t *testing.T, rows []VolumeSetupReversalQualityRow, groupType string, groupName string) {
	t.Helper()
	for _, row := range rows {
		if row.GroupType == groupType && row.GroupName == groupName {
			return
		}
	}
	t.Fatalf("missing group %s/%s in %+v", groupType, groupName, rows)
}
