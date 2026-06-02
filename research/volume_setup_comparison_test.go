package research

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestVolumeSetupComparisonFromFilesAndWriters(t *testing.T) {
	dir := t.TempDir()
	accumulationPath := filepath.Join(dir, "volume_setup_accumulation.csv")
	trendPath := filepath.Join(dir, "volume_setup_trend.csv")
	rejectionPath := filepath.Join(dir, "volume_setup_rejection.csv")
	reversalPath := filepath.Join(dir, "volume_setup_reversal.csv")
	csvPath := filepath.Join(dir, "volume_setup_comparison.csv")
	jsonPath := filepath.Join(dir, "volume_setup_comparison.json")
	mdPath := filepath.Join(dir, "volume_setup_comparison.md")

	writeTestFile(t, accumulationPath, strings.Join([]string{
		"timestamp,setup_id,setup_type,direction,accumulation_start,accumulation_end,setup_level,poc,vah,val,profile_shape,shape_confidence,profile_volume,accepted,rejected,daily_poc_confluence,rolling3d_poc_confluence,rolling7d_poc_confluence,composite30d_poc_confluence,retest_detected,invalidated,follow_through_5,follow_through_10,follow_through_20,notes",
		"1,a1,SIDEWAYS_ACCUMULATION,long_context,1,2,100,100,110,90,D_PROFILE,0.8,1000,true,false,true,false,false,false,true,false,1,2,3,test",
		"2,a2,SIDEWAYS_ACCUMULATION,short_context,1,2,100,100,110,90,P_PROFILE,0.6,500,false,true,false,true,false,false,false,true,-1,-2,-3,test",
	}, "\n"))
	writeTestFile(t, trendPath, strings.Join([]string{
		"timestamp,setup_id,setup_type,direction,trend_start,trend_end,trend_direction,trend_strength_score,setup_level,poc,vah,val,nearest_hvn,profile_shape,shape_confidence,profile_volume,accepted,rejected,poc_retest_detected,hvn_retest_detected,vwap_alignment,l2_book_pressure,invalidated,follow_through_5,follow_through_10,follow_through_20,notes",
		"1,t1,TREND_LEG,long_context,1,2,bullish,2,100,100,110,90,105,D_PROFILE,0.8,1000,true,false,true,true,above_vwap,bid_pressure,false,2,4,6,test",
		"2,t2,TREND_LEG,short_context,1,2,bearish,2,100,100,110,90,0,P_PROFILE,0.6,500,false,true,false,false,below_vwap,ask_pressure,true,-1,-2,-3,test",
	}, "\n"))
	writeTestFile(t, rejectionPath, strings.Join([]string{
		"timestamp,setup_id,setup_type,direction,rejection_start,rejection_end,rejection_level,rejection_high,rejection_low,setup_level,poc,vah,val,nearest_hvn,nearest_lvn,profile_shape,shape_confidence,profile_volume,accepted,rejected,poc_retest_detected,hvn_retest_detected,vwap_alignment,l2_book_pressure,daily_poc_confluence,rolling3d_poc_confluence,rolling7d_poc_confluence,composite30d_poc_confluence,invalidated,follow_through_5,follow_through_10,follow_through_20,notes",
		"1,r1,REJECTION,long_context,1,3,90,110,90,100,100,110,90,105,95,D_PROFILE,0.8,1000,true,false,true,true,above_vwap,bid_pressure,true,false,false,false,false,3,5,7,test",
		"2,r2,REJECTION,short_context,1,3,110,110,90,100,100,110,90,0,0,P_PROFILE,0.6,500,false,true,false,false,above_vwap,ask_pressure,false,false,false,false,true,-1,-2,-3,test",
	}, "\n"))
	writeTestFile(t, reversalPath, strings.Join([]string{
		"timestamp,setup_id,setup_type,direction,reversal_level,failed_auction_type,poc,vah,val,nearest_hvn,nearest_lvn,profile_shape,shape_confidence,profile_volume,vwap_alignment,poc_confluence,hvn_confluence,vah_val_rejection,invalidated,follow_through_5,follow_through_10,follow_through_20,notes",
		"1,v1,REVERSAL_TRADE,long_context,100,failed_low_auction,100,110,90,105,95,D_PROFILE,0.8,1000,above_vwap,true,true,true,false,4,6,8,test",
		"2,v2,REVERSAL_TRADE,short_context,100,failed_high_auction,100,110,90,0,0,P_PROFILE,0.6,500,below_vwap,false,false,true,false,2,3,4,test",
	}, "\n"))

	rows, summary, err := BuildVolumeSetupComparisonFromFiles(accumulationPath, trendPath, rejectionPath, reversalPath)
	if err != nil {
		t.Fatalf("build: %v", err)
	}
	if len(rows) != 4 || summary.BestFT20Setup != VolumeSetupReversal || summary.LargestSampleSetup == "" {
		t.Fatalf("rows=%+v summary=%+v", rows, summary)
	}
	reversal := rowBySetup(rows, VolumeSetupReversal)
	if reversal.POCConfluenceRate != 0.5 || reversal.HVNConfluenceRate != 0.5 || reversal.RankFT20 != 1 {
		t.Fatalf("reversal row=%+v", reversal)
	}

	if err := WriteVolumeSetupComparisonCSV(csvPath, rows); err != nil {
		t.Fatalf("write csv: %v", err)
	}
	body, err := os.ReadFile(csvPath)
	if err != nil {
		t.Fatalf("read csv: %v", err)
	}
	if !strings.Contains(string(body), "setup_type,setup_count,accepted_count,rejected_count") {
		t.Fatalf("missing csv header: %s", string(body))
	}
	if err := WriteVolumeSetupComparisonJSON(jsonPath, summary); err != nil {
		t.Fatalf("write json: %v", err)
	}
	jsonBody, err := os.ReadFile(jsonPath)
	if err != nil {
		t.Fatalf("read json: %v", err)
	}
	var decoded VolumeSetupComparisonSummary
	if err := json.Unmarshal(jsonBody, &decoded); err != nil {
		t.Fatalf("invalid json: %v", err)
	}
	if decoded.BestFT20Setup != VolumeSetupReversal {
		t.Fatalf("decoded=%+v", decoded)
	}
	if err := WriteVolumeSetupComparisonMarkdown(mdPath, rows, summary); err != nil {
		t.Fatalf("write md: %v", err)
	}
	md, err := os.ReadFile(mdPath)
	if err != nil {
		t.Fatalf("read md: %v", err)
	}
	if !strings.Contains(string(md), "# Volume Profile Setup Comparison") {
		t.Fatalf("missing markdown title: %s", string(md))
	}
}
