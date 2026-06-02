package research

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"AlgoTrading2026/volumeprofile"
)

func TestVolumeSetupReversalSummaryAndWriters(t *testing.T) {
	rows := []VolumeSetupReversalRow{{
		Timestamp:         1,
		SetupID:           "reversal_1",
		SetupType:         volumeprofile.ReversalSetupType,
		Direction:         "long_context",
		ReversalLevel:     100,
		FailedAuctionType: volumeprofile.FailedLowAuction,
		POC:               100,
		VAH:               110,
		VAL:               90,
		NearestHVN:        105,
		NearestLVN:        95,
		ProfileShape:      volumeprofile.ShapeDProfile,
		ShapeConfidence:   0.8,
		ProfileVolume:     1000,
		Accepted:          true,
		VWAPAlignment:     "above_vwap",
		POCConfluence:     true,
		HVNConfluence:     true,
		VAHVALRejection:   true,
		FollowThrough5:    1,
		FollowThrough10:   2,
		FollowThrough20:   3,
	}}
	summary := BuildVolumeSetupReversalSummary(rows)
	if summary.Setups != 1 || summary.LongContexts != 1 || summary.Accepted != 1 || summary.VWAPAligned != 1 || summary.POCConfluence != 1 || summary.HVNConfluence != 1 || summary.VAHVALRejections != 1 {
		t.Fatalf("summary=%+v", summary)
	}

	dir := t.TempDir()
	csvPath := filepath.Join(dir, "volume_setup_reversal.csv")
	jsonPath := filepath.Join(dir, "volume_setup_reversal_summary.json")
	if err := WriteVolumeSetupReversalCSV(csvPath, rows); err != nil {
		t.Fatalf("write csv: %v", err)
	}
	body, err := os.ReadFile(csvPath)
	if err != nil {
		t.Fatalf("read csv: %v", err)
	}
	if !strings.Contains(string(body), "accepted,rejected,vwap_alignment") {
		t.Fatalf("missing csv header: %s", string(body))
	}
	if err := WriteVolumeSetupReversalSummaryJSON(jsonPath, summary); err != nil {
		t.Fatalf("write json: %v", err)
	}
	jsonBody, err := os.ReadFile(jsonPath)
	if err != nil {
		t.Fatalf("read json: %v", err)
	}
	var decoded VolumeSetupReversalSummary
	if err := json.Unmarshal(jsonBody, &decoded); err != nil {
		t.Fatalf("invalid json: %v", err)
	}
	if decoded.AverageFollowThrough20 != 3 {
		t.Fatalf("decoded=%+v", decoded)
	}
}

func TestBuildVolumeSetupReversalFromFiles(t *testing.T) {
	dir := t.TempDir()
	rejectionPath := filepath.Join(dir, "volume_setup_rejection.csv")
	qualityPath := filepath.Join(dir, "volume_setup_rejection_quality.csv")
	phase3Path := filepath.Join(dir, "price_action_phase3_study.csv")
	flexiblePath := filepath.Join(dir, "flexible_volume_profile.csv")
	scopedPath := filepath.Join(dir, "volume_profile_scoped_features.csv")
	vwapPath := filepath.Join(dir, "vwap_features.csv")
	contextPath := filepath.Join(dir, "context_features.csv")

	writeTestFile(t, rejectionPath, strings.Join([]string{
		"timestamp,setup_id,setup_type,direction,rejection_start,rejection_end,rejection_level,rejection_high,rejection_low,setup_level,poc,vah,val,nearest_hvn,nearest_lvn,profile_shape,shape_confidence,profile_volume,accepted,rejected,poc_retest_detected,hvn_retest_detected,vwap_alignment,l2_book_pressure,daily_poc_confluence,rolling3d_poc_confluence,rolling7d_poc_confluence,composite30d_poc_confluence,invalidated,follow_through_5,follow_through_10,follow_through_20,notes",
		"10,r1,REJECTION,short_context,1,3,110,111,90,110,110,112,100,110,95,D_PROFILE,0.8,1000,false,false,true,true,below_vwap,ask_pressure,true,false,false,false,false,1,2,3,test",
		"20,r2,REJECTION,long_context,1,3,90,110,89.5,90,90,100,88,90,85,B_PROFILE,0.7,800,false,false,true,true,above_vwap,bid_pressure,false,false,false,false,false,1,2,3,test",
	}, "\n"))
	writeTestFile(t, phase3Path, strings.Join([]string{
		"timestamp,strategy,direction,level,level_type,confirmation,invalidated,follow_through_5,follow_through_10,follow_through_20,notes",
		"11,failed_high_auction,short_context,110,prior_high,true,false,2,3,4,test",
		"21,failed_low_auction,long_context,90,prior_low,true,false,5,6,7,test",
	}, "\n"))
	writeTestFile(t, qualityPath, "header\n")
	writeTestFile(t, flexiblePath, "header\n")
	writeTestFile(t, scopedPath, "header\n")
	writeTestFile(t, vwapPath, "header\n")
	writeTestFile(t, contextPath, "header\n")

	rows, summary, err := BuildVolumeSetupReversalFromFiles(rejectionPath, qualityPath, phase3Path, flexiblePath, scopedPath, vwapPath, contextPath, volumeprofile.ReversalSetupConfig{MaxEventDistanceMillis: 10, ConfluenceThresholdPct: 0.01, InvalidationPct: 0.02})
	if err != nil {
		t.Fatalf("build: %v", err)
	}
	if len(rows) != 2 || summary.Setups != 2 || summary.LongContexts != 1 || summary.ShortContexts != 1 {
		t.Fatalf("rows=%+v summary=%+v", rows, summary)
	}
	if !rows[0].POCConfluence || !rows[0].HVNConfluence || !rows[0].Accepted || rows[0].FollowThrough20 != 4 {
		t.Fatalf("rows=%+v", rows)
	}
}
