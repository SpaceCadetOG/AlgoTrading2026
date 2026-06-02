package research

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"AlgoTrading2026/volumeprofile"
)

func TestVolumeSetupTrendSummaryAndWriters(t *testing.T) {
	rows := []VolumeSetupTrendRow{{
		Timestamp:          1,
		SetupID:            "trend_leg_1",
		SetupType:          volumeprofile.TrendSetupType,
		Direction:          "long_context",
		TrendStart:         1,
		TrendEnd:           3,
		TrendDirection:     volumeprofile.TrendDirectionBullish,
		TrendStrengthScore: 1.5,
		SetupLevel:         100,
		POC:                100,
		VAH:                110,
		VAL:                90,
		NearestHVN:         105,
		ProfileShape:       volumeprofile.ShapeDProfile,
		ShapeConfidence:    0.8,
		ProfileVolume:      1000,
		Accepted:           true,
		POCRetestDetected:  true,
		HVNRetestDetected:  true,
		VWAPAlignment:      "above_vwap",
		L2BookPressure:     "bid_pressure",
		FollowThrough5:     1,
		FollowThrough10:    2,
		FollowThrough20:    3,
	}}
	summary := BuildVolumeSetupTrendSummary(rows)
	if summary.Setups != 1 || summary.LongContexts != 1 || summary.Accepted != 1 || summary.POCRetests != 1 || summary.HVNRetests != 1 || summary.VWAPAligned != 1 || summary.BidPressureAligned != 1 {
		t.Fatalf("summary=%+v", summary)
	}

	dir := t.TempDir()
	csvPath := filepath.Join(dir, "volume_setup_trend.csv")
	jsonPath := filepath.Join(dir, "volume_setup_trend_summary.json")
	if err := WriteVolumeSetupTrendCSV(csvPath, rows); err != nil {
		t.Fatalf("write csv: %v", err)
	}
	body, err := os.ReadFile(csvPath)
	if err != nil {
		t.Fatalf("read csv: %v", err)
	}
	if !strings.Contains(string(body), "timestamp,setup_id,setup_type,direction,trend_start,trend_end,trend_direction") {
		t.Fatalf("missing csv header: %s", string(body))
	}
	if err := WriteVolumeSetupTrendSummaryJSON(jsonPath, summary); err != nil {
		t.Fatalf("write json: %v", err)
	}
	jsonBody, err := os.ReadFile(jsonPath)
	if err != nil {
		t.Fatalf("read json: %v", err)
	}
	var decoded VolumeSetupTrendSummary
	if err := json.Unmarshal(jsonBody, &decoded); err != nil {
		t.Fatalf("invalid json: %v", err)
	}
	if decoded.AverageTrendStrength != 1.5 {
		t.Fatalf("decoded=%+v", decoded)
	}
}

func TestBuildVolumeSetupTrendFromFiles(t *testing.T) {
	dir := t.TempDir()
	pricePath := filepath.Join(dir, "price_action_features.csv")
	emptyPath := filepath.Join(dir, "empty.csv")
	writeTestFile(t, pricePath, strings.Join([]string{
		"timestamp,open,high,low,close,volume,range_pct,body_pct,upper_wick_pct,lower_wick_pct,aggression_direction,aggression_score,sideways_detected,sideways_high,sideways_low,sideways_duration,initiation_detected,initiation_direction,initiation_start_price,rejection_detected,rejection_direction,rejection_level,support_resistance_flip_detected,flip_level,flip_direction,session_open_level,daily_open_level,previous_daily_high,previous_daily_low,previous_weekly_high,previous_weekly_low,near_previous_high,near_previous_low,strong_high,strong_low,weak_high,weak_low,failed_auction_detected,failed_auction_direction",
		"1,100,112,100,111,10,0,0,0,0,bullish,1.6,false,0,0,0,true,up,100,false,,0,false,0,,0,0,0,0,0,0,false,false,false,false,false,false,false,",
		"2,111,122,111,121,10,0,0,0,0,bullish,1.4,false,0,0,0,false,,0,false,,0,false,0,,0,0,0,0,0,0,false,false,false,false,false,false,false,",
		"3,121,123,120,122,10,0,0,0,0,bullish,1.3,false,0,0,0,false,,0,false,,0,false,0,,0,0,0,0,0,0,false,false,false,false,false,false,false,",
		"4,122,126,124,125,10,0,0,0,0,none,0,false,0,0,0,false,,0,false,,0,false,0,,0,0,0,0,0,0,false,false,false,false,false,false,false,",
	}, "\n"))
	writeTestFile(t, emptyPath, "header\n")
	rows, summary, err := BuildVolumeSetupTrendFromFiles(pricePath, emptyPath, emptyPath, emptyPath, emptyPath, emptyPath, volumeprofile.TrendSetupConfig{MaxLegCandles: 2, RetestLookahead: 5, MinDirectionalCount: 2})
	if err != nil {
		t.Fatalf("build: %v", err)
	}
	if len(rows) != 1 || summary.Setups != 1 || !rows[0].POCRetestDetected {
		t.Fatalf("rows=%+v summary=%+v", rows, summary)
	}
}
