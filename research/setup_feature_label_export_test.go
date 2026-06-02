package research

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"AlgoTrading2026/exchanges"
)

func TestBuildSetupFeatureLabelExportAndWriters(t *testing.T) {
	dir := t.TempDir()
	paths := SetupFeatureLabelPaths{
		ReversalPath:     filepath.Join(dir, "volume_setup_reversal.csv"),
		AccumulationPath: filepath.Join(dir, "volume_setup_accumulation.csv"),
		TrendPath:        filepath.Join(dir, "volume_setup_trend.csv"),
		RejectionPath:    filepath.Join(dir, "volume_setup_rejection.csv"),
		VWAPPath:         filepath.Join(dir, "vwap_features.csv"),
		ContextPath:      filepath.Join(dir, "context_features.csv"),
		ScopedPath:       filepath.Join(dir, "volume_profile_scoped_features.csv"),
		PriceActionPath:  filepath.Join(dir, "price_action_features.csv"),
	}
	writeTestFile(t, paths.VWAPPath, strings.Join([]string{
		"timestamp,close,volume,session_vwap,anchored_vwap,distance_from_vwap,distance_from_anchor,vwap_slope,vwap_regime",
		"1,100,10,99,99,1,1,0.1,above_rising",
	}, "\n"))
	writeTestFile(t, paths.ContextPath, strings.Join([]string{
		"timestamp,close,session_vwap,distance_from_vwap,ema9,ema20,price_vs_vwap,price_vs_ema9,price_vs_ema20,ema_alignment,vwap_slope,trend_alignment",
		"1,100,99,1,99,98,1,1,2,bullish_alignment,0.1,bull",
	}, "\n"))
	writeTestFile(t, paths.PriceActionPath, strings.Join([]string{
		"timestamp,open,high,low,close,volume,range_pct,body_pct,upper_wick_pct,lower_wick_pct,aggression_direction,aggression_score,sideways_detected,sideways_high,sideways_low,sideways_duration,initiation_detected,initiation_direction,initiation_start_price,rejection_detected,rejection_direction,rejection_level,support_resistance_flip_detected,flip_level,flip_direction,session_open_level,daily_open_level,previous_daily_high,previous_daily_low,previous_weekly_high,previous_weekly_low,near_previous_high,near_previous_low,strong_high,strong_low,weak_high,weak_low,failed_auction_detected,failed_auction_direction",
		"1,100,101,99,100,10,0,0,0,0,none,0,false,0,0,0,false,,0,false,,0,false,0,,100,100,0,0,0,0,false,false,false,false,false,false,false,",
	}, "\n"))
	writeTestFile(t, paths.ScopedPath, "scope,timestamp,venue,symbol,profile_start,profile_end,bin_low,bin_high,bin_mid,bin_volume,poc,vah,val,is_poc,is_hvn,is_lvn,profile_shape,total_volume,notes\n")
	writeTestFile(t, paths.ReversalPath, strings.Join([]string{
		"timestamp,setup_id,setup_type,direction,reversal_level,failed_auction_type,poc,vah,val,nearest_hvn,nearest_lvn,profile_shape,shape_confidence,profile_volume,accepted,rejected,vwap_alignment,poc_confluence,hvn_confluence,vah_val_rejection,invalidated,follow_through_5,follow_through_10,follow_through_20,notes",
		"1,r1,REVERSAL_TRADE,long_context,100,failed_low_auction,100,110,90,105,95,D_PROFILE,0.8,1000,true,false,above_vwap,true,true,true,false,1,2,3,test",
	}, "\n"))
	writeTestFile(t, paths.AccumulationPath, "timestamp,setup_id,setup_type,direction,accumulation_start,accumulation_end,setup_level,poc,vah,val,profile_shape,shape_confidence,profile_volume,accepted,rejected,daily_poc_confluence,rolling3d_poc_confluence,rolling7d_poc_confluence,composite30d_poc_confluence,retest_detected,invalidated,follow_through_5,follow_through_10,follow_through_20,notes\n")
	writeTestFile(t, paths.TrendPath, "timestamp,setup_id,setup_type,direction,trend_start,trend_end,trend_direction,trend_strength_score,setup_level,poc,vah,val,nearest_hvn,profile_shape,shape_confidence,profile_volume,accepted,rejected,poc_retest_detected,hvn_retest_detected,vwap_alignment,l2_book_pressure,invalidated,follow_through_5,follow_through_10,follow_through_20,notes\n")
	writeTestFile(t, paths.RejectionPath, "timestamp,setup_id,setup_type,direction,rejection_start,rejection_end,rejection_level,rejection_high,rejection_low,setup_level,poc,vah,val,nearest_hvn,nearest_lvn,profile_shape,shape_confidence,profile_volume,accepted,rejected,poc_retest_detected,hvn_retest_detected,vwap_alignment,l2_book_pressure,daily_poc_confluence,rolling3d_poc_confluence,rolling7d_poc_confluence,composite30d_poc_confluence,invalidated,follow_through_5,follow_through_10,follow_through_20,notes\n")
	candles := make([]exchanges.Candle, 25)
	for i := range candles {
		price := "100"
		if i == 1 {
			price = "102"
		}
		candles[i] = exchanges.Candle{StartTime: int64(i + 1), Open: "100", High: price, Low: "99", Close: "100", Volume: "1"}
	}
	rows, summary, err := BuildSetupFeatureLabelExport(paths, candles)
	if err != nil {
		t.Fatalf("build: %v", err)
	}
	if len(rows) != 1 || summary.Rows != 1 || summary.BySetupType["REVERSAL"] != 1 {
		t.Fatalf("rows=%+v summary=%+v", rows, summary)
	}
	if rows[0].Feature.DistanceToPOCPct != 0 || rows[0].Label.LabelAcceptance != "accepted" {
		t.Fatalf("row=%+v", rows[0])
	}
	csvPath := filepath.Join(dir, "setup_features_labels.csv")
	jsonPath := filepath.Join(dir, "setup_labels_summary.json")
	if err := WriteSetupFeatureLabelCSV(csvPath, rows); err != nil {
		t.Fatalf("write csv: %v", err)
	}
	body, err := os.ReadFile(csvPath)
	if err != nil {
		t.Fatalf("read csv: %v", err)
	}
	if !strings.Contains(string(body), "setup_type,setup_id,timestamp,direction") {
		t.Fatalf("missing header: %s", string(body))
	}
	if err := WriteSetupLabelsSummaryJSON(jsonPath, summary); err != nil {
		t.Fatalf("write json: %v", err)
	}
	jsonBody, err := os.ReadFile(jsonPath)
	if err != nil {
		t.Fatalf("read json: %v", err)
	}
	var decoded SetupLabelsSummary
	if err := json.Unmarshal(jsonBody, &decoded); err != nil {
		t.Fatalf("invalid json: %v", err)
	}
	if decoded.AcceptanceLabels["accepted"] != 1 {
		t.Fatalf("decoded=%+v", decoded)
	}
}
