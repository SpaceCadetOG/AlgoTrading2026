package research

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestBuildProfileAcceptanceQualityRows(t *testing.T) {
	flexible := []flexibleProfileRecord{
		{ProfileType: "SIDEWAYS_ACCUMULATION", StartTime: 10, EndTime: 20, POC: 100, VAH: 110, VAL: 90, HVNCount: 2, LVNCount: 1, ProfileShape: "D_PROFILE", ShapeConfidence: 0.9, AcceptanceState: "accepted", ProfileVolume: 1000},
		{ProfileType: "SIDEWAYS_ACCUMULATION", StartTime: 30, EndTime: 40, POC: 120, VAH: 130, VAL: 100, HVNCount: 1, LVNCount: 3, ProfileShape: "P_PROFILE", ShapeConfidence: 0.6, AcceptanceState: "rejected", ProfileVolume: 500},
		{ProfileType: "OPEN_DRIVE", StartTime: 50, EndTime: 60, POC: 140, VAH: 150, VAL: 130, HVNCount: 1, LVNCount: 1, ProfileShape: "D_PROFILE", ShapeConfidence: 0.7, AcceptanceState: "accepted", ProfileVolume: 700},
	}
	scopes := []profileScopeWindow{{Scope: "daily_session", ProfileStart: 0, ProfileEnd: 45}, {Scope: "rolling_3d", ProfileStart: 0, ProfileEnd: 100}}
	rows := BuildProfileAcceptanceQualityRows(flexible, scopes)
	summary := BuildProfileAcceptanceQualitySummary(rows)
	if summary.AcceptedProfiles != 2 || summary.RejectedProfiles != 1 {
		t.Fatalf("summary=%+v rows=%+v", summary, rows)
	}
	if summary.BestProfileType != "OPEN_DRIVE" || summary.BestShape != "D_PROFILE" {
		t.Fatalf("bad bests: %+v", summary)
	}
}

func TestProfileAcceptanceQualityFromFilesAndWriters(t *testing.T) {
	dir := t.TempDir()
	flexiblePath := filepath.Join(dir, "flexible_volume_profile.csv")
	shapePath := filepath.Join(dir, "volume_profile_shape_study.csv")
	scopedPath := filepath.Join(dir, "volume_profile_scoped_features.csv")
	outCSV := filepath.Join(dir, "profile_acceptance_quality.csv")
	outJSON := filepath.Join(dir, "profile_acceptance_quality_summary.json")

	writeTestFile(t, flexiblePath, strings.Join([]string{
		"profile_type,profile_id,start_time,end_time,symbol,poc,vah,val,hvn_count,lvn_count,profile_shape,shape_confidence,acceptance_state,profile_volume,notes",
		"SIDEWAYS_ACCUMULATION,s1,10,20,BTCUSDT,100,110,90,2,1,D_PROFILE,0.9,accepted,1000,test",
		"FAILED_AUCTION,f1,30,40,BTCUSDT,120,130,100,1,3,P_PROFILE,0.6,rejected,500,test",
	}, "\n"))
	writeTestFile(t, shapePath, strings.Join([]string{
		"scope,profile_start,profile_end,symbol,poc,vah,val,profile_shape,total_bins,total_volume,upper_volume_pct,lower_volume_pct,hvn_count,lvn_count,distribution_balance,shape_confidence,shape_reason",
		"daily_session,0,25,BTCUSDT,100,110,90,D_PROFILE,10,1000,0.4,0.4,2,1,0,0.9,BALANCED_DISTRIBUTION",
	}, "\n"))
	writeTestFile(t, scopedPath, strings.Join([]string{
		"scope,timestamp,venue,symbol,profile_start,profile_end,bin_low,bin_high,bin_mid,bin_volume,poc,vah,val,is_poc,is_hvn,is_lvn,profile_shape,total_volume,notes",
		"daily_session,25,aster,BTCUSDT,0,25,90,100,95,10,100,110,90,false,false,false,D_PROFILE,1000,test",
	}, "\n"))

	rows, summary, err := BuildProfileAcceptanceQualityFromFiles(flexiblePath, shapePath, scopedPath)
	if err != nil {
		t.Fatalf("build from files: %v", err)
	}
	if summary.AcceptedProfiles != 1 || summary.RejectedProfiles != 1 || len(rows) == 0 {
		t.Fatalf("summary=%+v rows=%+v", summary, rows)
	}
	if err := WriteProfileAcceptanceQualityCSV(outCSV, rows); err != nil {
		t.Fatalf("write csv: %v", err)
	}
	body, err := os.ReadFile(outCSV)
	if err != nil {
		t.Fatalf("read csv: %v", err)
	}
	if !strings.Contains(string(body), "category,name,total_profiles,accepted_profiles,rejected_profiles,neutral_profiles,acceptance_rate,dominant_shape") {
		t.Fatalf("missing csv header: %s", string(body))
	}
	if err := WriteProfileAcceptanceQualitySummaryJSON(outJSON, summary); err != nil {
		t.Fatalf("write json: %v", err)
	}
	jsonBody, err := os.ReadFile(outJSON)
	if err != nil {
		t.Fatalf("read json: %v", err)
	}
	var decoded ProfileAcceptanceQualitySummary
	if err := json.Unmarshal(jsonBody, &decoded); err != nil {
		t.Fatalf("invalid json: %v", err)
	}
	if decoded.BestProfileType == "" {
		t.Fatalf("decoded=%+v", decoded)
	}
}

func writeTestFile(t *testing.T, path string, body string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(body+"\n"), 0644); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}
