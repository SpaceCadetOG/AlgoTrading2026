package research

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestVolumeSetupTrendQualityGrouping(t *testing.T) {
	records := []trendQualityRecord{
		{Direction: "long_context", TrendStrengthScore: 2.6, ProfileShape: "D_PROFILE", ProfileVolume: 1000, ShapeConfidence: 0.8, Accepted: true, POCRetestDetected: true, HVNRetestDetected: true, VWAPAlignment: "above_vwap", FollowThrough20: 10},
		{Direction: "short_context", TrendStrengthScore: 1.6, ProfileShape: "P_PROFILE", ProfileVolume: 500, ShapeConfidence: 0.6, Rejected: true, Invalidated: true, VWAPAlignment: "above_vwap", FollowThrough20: -2},
		{Direction: "long_context", TrendStrengthScore: 2.1, ProfileShape: "D_PROFILE", ProfileVolume: 800, ShapeConfidence: 0.7, POCRetestDetected: true, VWAPAlignment: "below_vwap", FollowThrough20: 4},
	}
	rows := BuildVolumeSetupTrendQualityRows(records)
	summary := BuildVolumeSetupTrendQualitySummary(records, rows)
	if summary.TotalSetups != 3 || summary.Accepted != 1 || summary.Rejected != 1 {
		t.Fatalf("summary=%+v", summary)
	}
	if summary.BestDirection != "long_context" || summary.BestShape != "D_PROFILE" || summary.BestTrendStrengthBucket != "strong" {
		t.Fatalf("bad bests: %+v rows=%+v", summary, rows)
	}
	assertTrendQualityGroup(t, rows, "acceptance_state", "accepted")
	assertTrendQualityGroup(t, rows, "direction", "long_context")
	assertTrendQualityGroup(t, rows, "poc_retest", "with_poc_retest")
	assertTrendQualityGroup(t, rows, "hvn_retest", "with_hvn_retest")
	assertTrendQualityGroup(t, rows, "vwap_alignment", "vwap_aligned")
	assertTrendQualityGroup(t, rows, "profile_shape", "D_PROFILE")
	assertTrendQualityGroup(t, rows, "trend_strength_bucket", "strong")
}

func TestVolumeSetupTrendQualityFromFileAndWriters(t *testing.T) {
	dir := t.TempDir()
	inputPath := filepath.Join(dir, "volume_setup_trend.csv")
	csvPath := filepath.Join(dir, "volume_setup_trend_quality.csv")
	jsonPath := filepath.Join(dir, "volume_setup_trend_quality_summary.json")
	mdPath := filepath.Join(dir, "volume_setup_trend_quality.md")
	writeTestFile(t, inputPath, strings.Join([]string{
		"timestamp,setup_id,setup_type,direction,trend_start,trend_end,trend_direction,trend_strength_score,setup_level,poc,vah,val,nearest_hvn,profile_shape,shape_confidence,profile_volume,accepted,rejected,poc_retest_detected,hvn_retest_detected,vwap_alignment,l2_book_pressure,invalidated,follow_through_5,follow_through_10,follow_through_20,notes",
		"1,t1,TREND_LEG,long_context,1,2,bullish,2.5,100,100,110,90,105,D_PROFILE,0.8,1000,true,false,true,true,above_vwap,bid_pressure,false,1,2,3,test",
		"2,t2,TREND_LEG,short_context,1,2,bearish,1.5,100,100,110,90,0,P_PROFILE,0.6,500,false,true,false,false,above_vwap,ask_pressure,true,-1,-2,-3,test",
	}, "\n"))
	rows, summary, err := BuildVolumeSetupTrendQualityFromFile(inputPath)
	if err != nil {
		t.Fatalf("build: %v", err)
	}
	if len(rows) == 0 || summary.TotalSetups != 2 || summary.Accepted != 1 || summary.Rejected != 1 {
		t.Fatalf("rows=%+v summary=%+v", rows, summary)
	}
	if err := WriteVolumeSetupTrendQualityCSV(csvPath, rows); err != nil {
		t.Fatalf("write csv: %v", err)
	}
	body, err := os.ReadFile(csvPath)
	if err != nil {
		t.Fatalf("read csv: %v", err)
	}
	if !strings.Contains(string(body), "group_type,group_name,setup_count,accepted_count,rejected_count") {
		t.Fatalf("missing csv header: %s", string(body))
	}
	if err := WriteVolumeSetupTrendQualitySummaryJSON(jsonPath, summary); err != nil {
		t.Fatalf("write json: %v", err)
	}
	jsonBody, err := os.ReadFile(jsonPath)
	if err != nil {
		t.Fatalf("read json: %v", err)
	}
	var decoded VolumeSetupTrendQualitySummary
	if err := json.Unmarshal(jsonBody, &decoded); err != nil {
		t.Fatalf("invalid json: %v", err)
	}
	if decoded.TotalSetups != 2 {
		t.Fatalf("decoded=%+v", decoded)
	}
	if err := WriteVolumeSetupTrendQualityMarkdown(mdPath, rows, summary); err != nil {
		t.Fatalf("write md: %v", err)
	}
	md, err := os.ReadFile(mdPath)
	if err != nil {
		t.Fatalf("read md: %v", err)
	}
	if !strings.Contains(string(md), "# Volume Setup #2 Trend Quality Review") {
		t.Fatalf("missing markdown title: %s", string(md))
	}
}

func assertTrendQualityGroup(t *testing.T, rows []VolumeSetupTrendQualityRow, groupType string, groupName string) {
	t.Helper()
	for _, row := range rows {
		if row.GroupType == groupType && row.GroupName == groupName {
			return
		}
	}
	t.Fatalf("missing group %s/%s in %+v", groupType, groupName, rows)
}
