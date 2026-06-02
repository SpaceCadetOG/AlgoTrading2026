package research

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestVolumeSetupRejectionQualityGrouping(t *testing.T) {
	records := []rejectionQualityRecord{
		{Direction: "long_context", ProfileShape: "D_PROFILE", ProfileVolume: 1000, ShapeConfidence: 0.8, Accepted: true, POCRetestDetected: true, HVNRetestDetected: true, VWAPAlignment: "above_vwap", DailyPOCConfluence: true, FollowThrough20: 10},
		{Direction: "short_context", ProfileShape: "P_PROFILE", ProfileVolume: 500, ShapeConfidence: 0.6, Rejected: true, Invalidated: true, VWAPAlignment: "above_vwap", Rolling3DPOCConfluence: true, FollowThrough20: -2},
		{Direction: "long_context", ProfileShape: "D_PROFILE", ProfileVolume: 800, ShapeConfidence: 0.7, POCRetestDetected: true, VWAPAlignment: "below_vwap", Composite30DPOCConfluence: true, FollowThrough20: 4},
	}
	rows := BuildVolumeSetupRejectionQualityRows(records)
	summary := BuildVolumeSetupRejectionQualitySummary(records, rows)
	if summary.TotalSetups != 3 || summary.Accepted != 1 || summary.Rejected != 1 {
		t.Fatalf("summary=%+v", summary)
	}
	if summary.BestDirection != "long_context" || summary.BestShape != "D_PROFILE" || summary.BestConfluence == "" || summary.BestFilter == "" {
		t.Fatalf("bad bests: %+v rows=%+v", summary, rows)
	}
	assertRejectionQualityGroup(t, rows, "acceptance_state", "accepted")
	assertRejectionQualityGroup(t, rows, "direction", "long_context")
	assertRejectionQualityGroup(t, rows, "vwap_alignment", "vwap_aligned")
	assertRejectionQualityGroup(t, rows, "poc_retest", "with_poc_retest")
	assertRejectionQualityGroup(t, rows, "hvn_retest", "with_hvn_retest")
	assertRejectionQualityGroup(t, rows, "daily_poc_confluence", "with_daily_poc")
	assertRejectionQualityGroup(t, rows, "rolling3d_poc_confluence", "with_rolling3d_poc")
	assertRejectionQualityGroup(t, rows, "composite30d_poc_confluence", "with_composite30d_poc")
	assertRejectionQualityGroup(t, rows, "profile_shape", "D_PROFILE")
}

func TestVolumeSetupRejectionQualityFromFileAndWriters(t *testing.T) {
	dir := t.TempDir()
	inputPath := filepath.Join(dir, "volume_setup_rejection.csv")
	csvPath := filepath.Join(dir, "volume_setup_rejection_quality.csv")
	jsonPath := filepath.Join(dir, "volume_setup_rejection_quality_summary.json")
	mdPath := filepath.Join(dir, "volume_setup_rejection_quality.md")
	writeTestFile(t, inputPath, strings.Join([]string{
		"timestamp,setup_id,setup_type,direction,rejection_start,rejection_end,rejection_level,rejection_high,rejection_low,setup_level,poc,vah,val,nearest_hvn,nearest_lvn,profile_shape,shape_confidence,profile_volume,accepted,rejected,poc_retest_detected,hvn_retest_detected,vwap_alignment,l2_book_pressure,daily_poc_confluence,rolling3d_poc_confluence,rolling7d_poc_confluence,composite30d_poc_confluence,invalidated,follow_through_5,follow_through_10,follow_through_20,notes",
		"1,r1,REJECTION,long_context,1,3,90,110,90,100,100,110,90,105,95,D_PROFILE,0.8,1000,true,false,true,true,above_vwap,bid_pressure,true,false,false,false,false,1,2,3,test",
		"2,r2,REJECTION,short_context,1,3,110,110,90,100,100,110,90,0,0,P_PROFILE,0.6,500,false,true,false,false,above_vwap,ask_pressure,false,true,false,false,true,-1,-2,-3,test",
	}, "\n"))
	rows, summary, err := BuildVolumeSetupRejectionQualityFromFile(inputPath)
	if err != nil {
		t.Fatalf("build: %v", err)
	}
	if len(rows) == 0 || summary.TotalSetups != 2 || summary.Accepted != 1 || summary.Rejected != 1 {
		t.Fatalf("rows=%+v summary=%+v", rows, summary)
	}
	if err := WriteVolumeSetupRejectionQualityCSV(csvPath, rows); err != nil {
		t.Fatalf("write csv: %v", err)
	}
	body, err := os.ReadFile(csvPath)
	if err != nil {
		t.Fatalf("read csv: %v", err)
	}
	if !strings.Contains(string(body), "group_type,group_name,setup_count,accepted_count,rejected_count") {
		t.Fatalf("missing csv header: %s", string(body))
	}
	if err := WriteVolumeSetupRejectionQualitySummaryJSON(jsonPath, summary); err != nil {
		t.Fatalf("write json: %v", err)
	}
	jsonBody, err := os.ReadFile(jsonPath)
	if err != nil {
		t.Fatalf("read json: %v", err)
	}
	var decoded VolumeSetupRejectionQualitySummary
	if err := json.Unmarshal(jsonBody, &decoded); err != nil {
		t.Fatalf("invalid json: %v", err)
	}
	if decoded.TotalSetups != 2 {
		t.Fatalf("decoded=%+v", decoded)
	}
	if err := WriteVolumeSetupRejectionQualityMarkdown(mdPath, rows, summary); err != nil {
		t.Fatalf("write md: %v", err)
	}
	md, err := os.ReadFile(mdPath)
	if err != nil {
		t.Fatalf("read md: %v", err)
	}
	if !strings.Contains(string(md), "# Volume Setup #3 Rejection Quality Review") {
		t.Fatalf("missing markdown title: %s", string(md))
	}
}

func assertRejectionQualityGroup(t *testing.T, rows []VolumeSetupRejectionQualityRow, groupType string, groupName string) {
	t.Helper()
	for _, row := range rows {
		if row.GroupType == groupType && row.GroupName == groupName {
			return
		}
	}
	t.Fatalf("missing group %s/%s in %+v", groupType, groupName, rows)
}
