package research

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"AlgoTrading2026/volumeprofile"
)

func TestVolumeSetupRejectionSummaryAndWriters(t *testing.T) {
	rows := []VolumeSetupRejectionRow{{
		Timestamp:                 1,
		SetupID:                   "rejection_1",
		SetupType:                 volumeprofile.RejectionSetupType,
		Direction:                 "long_context",
		RejectionStart:            1,
		RejectionEnd:              3,
		RejectionLevel:            90,
		RejectionHigh:             110,
		RejectionLow:              90,
		SetupLevel:                100,
		POC:                       100,
		VAH:                       110,
		VAL:                       90,
		NearestHVN:                105,
		NearestLVN:                95,
		ProfileShape:              volumeprofile.ShapeDProfile,
		ShapeConfidence:           0.8,
		ProfileVolume:             1000,
		Accepted:                  true,
		POCRetestDetected:         true,
		HVNRetestDetected:         true,
		VWAPAlignment:             "above_vwap",
		L2BookPressure:            "bid_pressure",
		DailyPOCConfluence:        true,
		Rolling3DPOCConfluence:    true,
		Rolling7DPOCConfluence:    true,
		Composite30DPOCConfluence: true,
		FollowThrough5:            1,
		FollowThrough10:           2,
		FollowThrough20:           3,
	}}
	summary := BuildVolumeSetupRejectionSummary(rows)
	if summary.Setups != 1 || summary.LongContexts != 1 || summary.Accepted != 1 || summary.POCRetests != 1 || summary.HVNRetests != 1 || summary.VWAPAligned != 1 || summary.BidPressureAligned != 1 || summary.DailyPOCConfluence != 1 {
		t.Fatalf("summary=%+v", summary)
	}

	dir := t.TempDir()
	csvPath := filepath.Join(dir, "volume_setup_rejection.csv")
	jsonPath := filepath.Join(dir, "volume_setup_rejection_summary.json")
	if err := WriteVolumeSetupRejectionCSV(csvPath, rows); err != nil {
		t.Fatalf("write csv: %v", err)
	}
	body, err := os.ReadFile(csvPath)
	if err != nil {
		t.Fatalf("read csv: %v", err)
	}
	if !strings.Contains(string(body), "timestamp,setup_id,setup_type,direction,rejection_start") {
		t.Fatalf("missing csv header: %s", string(body))
	}
	if err := WriteVolumeSetupRejectionSummaryJSON(jsonPath, summary); err != nil {
		t.Fatalf("write json: %v", err)
	}
	jsonBody, err := os.ReadFile(jsonPath)
	if err != nil {
		t.Fatalf("read json: %v", err)
	}
	var decoded VolumeSetupRejectionSummary
	if err := json.Unmarshal(jsonBody, &decoded); err != nil {
		t.Fatalf("invalid json: %v", err)
	}
	if decoded.AverageFollowThrough20 != 3 {
		t.Fatalf("decoded=%+v", decoded)
	}
}

func TestBuildVolumeSetupRejectionFromFiles(t *testing.T) {
	dir := t.TempDir()
	pricePath := filepath.Join(dir, "price_action_features.csv")
	phase3Path := filepath.Join(dir, "price_action_phase3_study.csv")
	flexiblePath := filepath.Join(dir, "flexible_volume_profile.csv")
	scopedPath := filepath.Join(dir, "volume_profile_scoped_features.csv")
	acceptancePath := filepath.Join(dir, "profile_acceptance_quality.csv")
	vwapPath := filepath.Join(dir, "vwap_features.csv")
	contextPath := filepath.Join(dir, "context_features.csv")
	contextL2Path := filepath.Join(dir, "context_features_l2.csv")

	writeTestFile(t, pricePath, strings.Join([]string{
		"timestamp,open,high,low,close,volume,range_pct,body_pct,upper_wick_pct,lower_wick_pct,aggression_direction,aggression_score,sideways_detected,sideways_high,sideways_low,sideways_duration,initiation_detected,initiation_direction,initiation_start_price,rejection_detected,rejection_direction,rejection_level,support_resistance_flip_detected,flip_level,flip_direction,session_open_level,daily_open_level,previous_daily_high,previous_daily_low,previous_weekly_high,previous_weekly_low,near_previous_high,near_previous_low,strong_high,strong_low,weak_high,weak_low,failed_auction_detected,failed_auction_direction",
		"1,105,106,94,95,20,0,0,0,0,bearish,1.5,false,0,0,0,false,,0,false,,0,false,0,,0,0,0,0,0,0,false,false,false,false,false,false,false,",
		"2,95,101,90,100,30,0,0,0,0,none,0,false,0,0,0,false,,0,true,bullish,90,false,0,,0,0,0,0,0,0,false,false,false,false,false,false,false,",
		"3,100,111,99,110,25,0,0,0,0,bullish,1.5,false,0,0,0,false,,0,false,,0,false,0,,0,0,0,0,0,0,false,false,false,false,false,false,false,",
		"4,108,109,94,98,10,0,0,0,0,none,0,false,0,0,0,false,,0,false,,0,false,0,,0,0,0,0,0,0,false,false,false,false,false,false,false,",
		"10,98,116,97,115,10,0,0,0,0,none,0,false,0,0,0,false,,0,false,,0,false,0,,0,0,0,0,0,0,false,false,false,false,false,false,false,",
	}, "\n"))
	writeTestFile(t, scopedPath, strings.Join([]string{
		"scope,timestamp,venue,symbol,profile_start,profile_end,bin_low,bin_high,bin_mid,bin_volume,poc,vah,val,is_poc,is_hvn,is_lvn,profile_shape,total_volume,notes",
		"daily_session,1,aster,BTCUSDT,1,3,90,100,95,10,95,110,90,true,false,false,D_PROFILE,100,test",
		"rolling_3d,1,aster,BTCUSDT,1,3,90,100,95,10,95,110,90,true,false,false,D_PROFILE,100,test",
		"rolling_7d,1,aster,BTCUSDT,1,3,90,100,95,10,95,110,90,true,false,false,D_PROFILE,100,test",
		"composite_30d,1,aster,BTCUSDT,1,3,90,100,95,10,95,110,90,true,false,false,D_PROFILE,100,test",
	}, "\n"))
	writeTestFile(t, vwapPath, strings.Join([]string{
		"timestamp,close,volume,session_vwap,anchored_vwap,distance_from_vwap,distance_from_anchor,vwap_slope,vwap_regime",
		"4,98,10,95,95,3,3,1,above_rising",
	}, "\n"))
	writeTestFile(t, contextPath, strings.Join([]string{
		"timestamp,close,session_vwap,distance_from_vwap,ema9,ema20,price_vs_vwap,price_vs_ema9,price_vs_ema20,ema_alignment,vwap_slope,trend_alignment",
		"4,98,95,3,97,96,above_vwap,above,above,bullish_alignment,1,bull",
	}, "\n"))
	writeTestFile(t, contextL2Path, strings.Join([]string{
		"timestamp,venue,symbol,close,session_vwap,distance_from_vwap,ema9,ema20,ema_alignment,vwap_slope,trend_alignment,spread_pct,imbalance_1pct,bid_depth_1pct,ask_depth_1pct,liquidity_near_vwap,spread_quality,book_pressure,liquidity_quality",
		"4,aster,BTCUSDT,98,95,3,97,96,bullish_alignment,1,bull,0.01,0.5,100,50,20,tight,bid_pressure,strong_near_vwap",
	}, "\n"))
	writeTestFile(t, phase3Path, "header\n")
	writeTestFile(t, flexiblePath, "header\n")
	writeTestFile(t, acceptancePath, "header\n")

	rows, summary, err := BuildVolumeSetupRejectionFromFiles(pricePath, phase3Path, flexiblePath, scopedPath, acceptancePath, vwapPath, contextPath, contextL2Path, volumeprofile.RejectionSetupConfig{ConfirmationLookahead: 3, RetestLookahead: 5, ConfluenceThresholdPct: 0.10})
	if err != nil {
		t.Fatalf("build: %v", err)
	}
	if len(rows) != 1 || summary.Setups != 1 || !rows[0].POCRetestDetected || rows[0].VWAPAlignment != "above_vwap" || rows[0].L2BookPressure != "bid_pressure" {
		t.Fatalf("rows=%+v summary=%+v", rows, summary)
	}
}
