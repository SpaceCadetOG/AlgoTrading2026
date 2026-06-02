package research

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestVolumeSetupAccumulationQualityGrouping(t *testing.T) {
	records := []accumulationQualityRecord{
		{Direction: "long_context", ProfileShape: "D_PROFILE", ProfileVolume: 100, ShapeConfidence: 0.8, Accepted: true, DailyPOCConfluence: true, RetestDetected: true, FollowThrough20: 5},
		{Direction: "short_context", ProfileShape: "P_PROFILE", ProfileVolume: 200, ShapeConfidence: 0.6, Rejected: true, Rolling3DPOCConfluence: true, Invalidated: true, FollowThrough20: -2},
		{Direction: "long_context", ProfileShape: "D_PROFILE", ProfileVolume: 300, ShapeConfidence: 0.9, RetestDetected: false, Composite30DPOCConfluence: true, FollowThrough20: 3},
	}
	rows := BuildVolumeSetupAccumulationQualityRows(records)
	summary := BuildVolumeSetupAccumulationQualitySummary(records, rows)
	if summary.TotalSetups != 3 || summary.Accepted != 1 || summary.Rejected != 1 {
		t.Fatalf("summary=%+v", summary)
	}
	if summary.BestShape != "D_PROFILE" || summary.BestDirection != "long_context" {
		t.Fatalf("bad bests: %+v rows=%+v", summary, rows)
	}
	assertQualityGroup(t, rows, "acceptance_state", "accepted")
	assertQualityGroup(t, rows, "direction", "long_context")
	assertQualityGroup(t, rows, "daily_poc_confluence", "with_daily_poc")
	assertQualityGroup(t, rows, "profile_shape", "D_PROFILE")
	assertQualityGroup(t, rows, "retest", "retested")
}

func TestVolumeSetupAccumulationQualityFromFileAndWriters(t *testing.T) {
	dir := t.TempDir()
	inputPath := filepath.Join(dir, "volume_setup_accumulation.csv")
	csvPath := filepath.Join(dir, "volume_setup_accumulation_quality.csv")
	jsonPath := filepath.Join(dir, "volume_setup_accumulation_quality_summary.json")
	mdPath := filepath.Join(dir, "volume_setup_accumulation_quality.md")
	writeTestFile(t, inputPath, strings.Join([]string{
		"timestamp,setup_id,setup_type,direction,accumulation_start,accumulation_end,setup_level,poc,vah,val,profile_shape,shape_confidence,profile_volume,accepted,rejected,daily_poc_confluence,rolling3d_poc_confluence,rolling7d_poc_confluence,composite30d_poc_confluence,retest_detected,invalidated,follow_through_5,follow_through_10,follow_through_20,notes",
		"1,a1,SIDEWAYS_ACCUMULATION,long_context,1,2,100,100,110,90,D_PROFILE,0.8,1000,true,false,true,false,false,false,true,false,1,2,3,test",
		"2,a2,SIDEWAYS_ACCUMULATION,short_context,1,2,100,100,110,90,P_PROFILE,0.6,500,false,true,false,true,false,false,false,true,-1,-2,-3,test",
	}, "\n"))
	rows, summary, err := BuildVolumeSetupAccumulationQualityFromFile(inputPath)
	if err != nil {
		t.Fatalf("build: %v", err)
	}
	if len(rows) == 0 || summary.TotalSetups != 2 || summary.Accepted != 1 || summary.Rejected != 1 {
		t.Fatalf("rows=%+v summary=%+v", rows, summary)
	}
	if err := WriteVolumeSetupAccumulationQualityCSV(csvPath, rows); err != nil {
		t.Fatalf("write csv: %v", err)
	}
	body, err := os.ReadFile(csvPath)
	if err != nil {
		t.Fatalf("read csv: %v", err)
	}
	if !strings.Contains(string(body), "group_type,group_name,setup_count,accepted_count,rejected_count") {
		t.Fatalf("missing csv header: %s", string(body))
	}
	if err := WriteVolumeSetupAccumulationQualitySummaryJSON(jsonPath, summary); err != nil {
		t.Fatalf("write json: %v", err)
	}
	jsonBody, err := os.ReadFile(jsonPath)
	if err != nil {
		t.Fatalf("read json: %v", err)
	}
	var decoded VolumeSetupAccumulationQualitySummary
	if err := json.Unmarshal(jsonBody, &decoded); err != nil {
		t.Fatalf("invalid json: %v", err)
	}
	if decoded.TotalSetups != 2 {
		t.Fatalf("decoded=%+v", decoded)
	}
	if err := WriteVolumeSetupAccumulationQualityMarkdown(mdPath, rows, summary); err != nil {
		t.Fatalf("write md: %v", err)
	}
	md, err := os.ReadFile(mdPath)
	if err != nil {
		t.Fatalf("read md: %v", err)
	}
	if !strings.Contains(string(md), "# Volume Setup #1 Accumulation Quality Review") {
		t.Fatalf("missing markdown title: %s", string(md))
	}
}

func assertQualityGroup(t *testing.T, rows []VolumeSetupAccumulationQualityRow, groupType string, groupName string) {
	t.Helper()
	for _, row := range rows {
		if row.GroupType == groupType && row.GroupName == groupName {
			return
		}
	}
	t.Fatalf("missing group %s/%s in %+v", groupType, groupName, rows)
}
