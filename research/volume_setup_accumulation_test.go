package research

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"AlgoTrading2026/volumeprofile"
)

func TestVolumeSetupAccumulationSummaryAndWriters(t *testing.T) {
	rows := []VolumeSetupAccumulationRow{{
		Timestamp:                 10,
		SetupID:                   "a1",
		SetupType:                 volumeprofile.FlexibleSidewaysAccumulation,
		Direction:                 volumeprofile.AccumulationDirectionLong,
		AccumulationStart:         1,
		AccumulationEnd:           2,
		SetupLevel:                100,
		POC:                       100,
		VAH:                       110,
		VAL:                       90,
		ProfileShape:              volumeprofile.ShapeDProfile,
		ShapeConfidence:           0.8,
		ProfileVolume:             1000,
		Accepted:                  true,
		DailyPOCConfluence:        true,
		Rolling3DPOCConfluence:    true,
		Rolling7DPOCConfluence:    true,
		Composite30DPOCConfluence: true,
		RetestDetected:            true,
		FollowThrough5:            1,
		FollowThrough10:           2,
		FollowThrough20:           3,
	}}
	summary := BuildVolumeSetupAccumulationSummary(rows)
	if summary.Setups != 1 || summary.LongContexts != 1 || summary.Accepted != 1 || summary.Retests != 1 || summary.AverageFollowThrough10 != 2 {
		t.Fatalf("summary=%+v", summary)
	}

	dir := t.TempDir()
	csvPath := filepath.Join(dir, "volume_setup_accumulation.csv")
	jsonPath := filepath.Join(dir, "volume_setup_accumulation_summary.json")
	if err := WriteVolumeSetupAccumulationCSV(csvPath, rows); err != nil {
		t.Fatalf("write csv: %v", err)
	}
	body, err := os.ReadFile(csvPath)
	if err != nil {
		t.Fatalf("read csv: %v", err)
	}
	text := string(body)
	if !strings.Contains(text, "timestamp,setup_id,setup_type,direction,accumulation_start,accumulation_end,setup_level,poc,vah,val") {
		t.Fatalf("missing header: %s", text)
	}
	if err := WriteVolumeSetupAccumulationSummaryJSON(jsonPath, summary); err != nil {
		t.Fatalf("write json: %v", err)
	}
	jsonBody, err := os.ReadFile(jsonPath)
	if err != nil {
		t.Fatalf("read json: %v", err)
	}
	var decoded VolumeSetupAccumulationSummary
	if err := json.Unmarshal(jsonBody, &decoded); err != nil {
		t.Fatalf("invalid json: %v", err)
	}
	if decoded.DailyPOCConfluence != 1 {
		t.Fatalf("decoded=%+v", decoded)
	}
}

func TestBuildVolumeSetupAccumulationFromFiles(t *testing.T) {
	dir := t.TempDir()
	flexiblePath := filepath.Join(dir, "flexible_volume_profile.csv")
	pricePath := filepath.Join(dir, "price_action_features.csv")
	strategyPath := filepath.Join(dir, "price_action_strategy_study.csv")
	scopedPath := filepath.Join(dir, "volume_profile_scoped_features.csv")
	acceptancePath := filepath.Join(dir, "profile_acceptance_quality.csv")

	writeTestFile(t, flexiblePath, strings.Join([]string{
		"profile_type,profile_id,start_time,end_time,symbol,poc,vah,val,hvn_count,lvn_count,profile_shape,shape_confidence,acceptance_state,profile_volume,notes",
		"SIDEWAYS_ACCUMULATION,a1,1,2,BTCUSDT,100,110,90,1,1,D_PROFILE,0.8,accepted,1000,test",
	}, "\n"))
	writeTestFile(t, pricePath, strings.Join([]string{
		"timestamp,open,high,low,close,volume,range_pct,body_pct,upper_wick_pct,lower_wick_pct,aggression_direction,aggression_score,sideways_detected,sideways_high,sideways_low,sideways_duration,initiation_detected,initiation_direction,initiation_start_price,rejection_detected,rejection_direction,rejection_level,support_resistance_flip_detected,flip_level,flip_direction,session_open_level,daily_open_level,previous_daily_high,previous_daily_low,previous_weekly_high,previous_weekly_low,near_previous_high,near_previous_low,strong_high,strong_low,weak_high,weak_low,failed_auction_detected,failed_auction_direction",
		"3,100,120,111,118,10,0,0,0,0,bullish,2,false,0,0,0,true,up,100,false,,0,false,0,,0,0,0,0,0,0,false,false,false,false,false,false,false,",
		"4,100,101,99,100,10,0,0,0,0,none,0,false,0,0,0,false,,0,false,,0,false,0,,0,0,0,0,0,0,false,false,false,false,false,false,false,",
	}, "\n"))
	writeTestFile(t, strategyPath, "timestamp,strategy,direction,level,entry_zone,confirmation,invalidated,follow_through_5,follow_through_10,follow_through_20,notes\n")
	writeTestFile(t, scopedPath, strings.Join([]string{
		"scope,timestamp,venue,symbol,profile_start,profile_end,bin_low,bin_high,bin_mid,bin_volume,poc,vah,val,is_poc,is_hvn,is_lvn,profile_shape,total_volume,notes",
		"daily_session,10,aster,BTCUSDT,0,10,90,100,95,10,100,110,90,false,false,false,D_PROFILE,1000,test",
	}, "\n"))
	writeTestFile(t, acceptancePath, "category,name,total_profiles,accepted_profiles,rejected_profiles,neutral_profiles,acceptance_rate,dominant_shape,average_confidence,average_hvn_count,average_lvn_count,average_volume,average_distance_from_poc\n")

	rows, summary, err := BuildVolumeSetupAccumulationFromFiles(flexiblePath, pricePath, strategyPath, scopedPath, acceptancePath, volumeprofile.AccumulationSetupConfig{InitiationLookahead: 5, RetestLookahead: 5})
	if err != nil {
		t.Fatalf("build from files: %v", err)
	}
	if len(rows) != 1 || summary.Setups != 1 || !rows[0].DailyPOCConfluence || !rows[0].RetestDetected {
		t.Fatalf("rows=%+v summary=%+v", rows, summary)
	}
}
