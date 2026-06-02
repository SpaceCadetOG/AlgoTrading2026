package research

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"AlgoTrading2026/exchanges"
	"AlgoTrading2026/volumeprofile"
)

func TestBuildFlexibleVolumeProfilesAndSummary(t *testing.T) {
	candles := researchFlexibleCandles(30)
	priceRows := make([]PriceActionFeatureRow, len(candles))
	for i, candle := range candles {
		priceRows[i] = PriceActionFeatureRow{Timestamp: candle.StartTime}
	}
	priceRows[3].SidewaysDetected = true
	priceRows[3].SidewaysDuration = 4
	strategyRows := []PriceActionStrategyStudyRow{{Timestamp: candles[10].StartTime, Strategy: "open_drive"}}
	phase3Rows := []PriceActionPhase3StudyRow{
		{Timestamp: candles[15].StartTime, Strategy: "failed_high_auction"},
		{Timestamp: candles[20].StartTime, Strategy: "daily_high_retest"},
	}

	profiles := BuildFlexibleVolumeProfiles(candles, priceRows, strategyRows, phase3Rows)
	rows := BuildFlexibleVolumeProfileRows(profiles)
	summary := BuildFlexibleVolumeProfileSummary(rows)
	if summary.SidewaysAccumulationProfiles != 1 || summary.OpenDriveProfiles != 1 || summary.FailedAuctionProfiles != 1 || summary.HighLowRetestProfiles != 1 {
		t.Fatalf("summary=%+v rows=%+v", summary, rows)
	}
}

func TestFlexibleVolumeProfileWriters(t *testing.T) {
	rows := []FlexibleVolumeProfileRow{{
		ProfileType:     volumeprofile.FlexibleOpenDrive,
		ProfileID:       "open_drive_1",
		StartTime:       1,
		EndTime:         2,
		Symbol:          "BTCUSDT",
		POC:             100,
		VAH:             110,
		VAL:             90,
		HVNCount:        1,
		LVNCount:        1,
		ProfileShape:    volumeprofile.ShapeDProfile,
		ShapeConfidence: 0.8,
		AcceptanceState: volumeprofile.AcceptanceAccepted,
		ProfileVolume:   10,
		Notes:           "test",
	}}
	summary := BuildFlexibleVolumeProfileSummary(rows)
	dir := t.TempDir()
	csvPath := filepath.Join(dir, "flexible_volume_profile.csv")
	jsonPath := filepath.Join(dir, "flexible_volume_profile_summary.json")
	if err := WriteFlexibleVolumeProfileCSV(csvPath, rows); err != nil {
		t.Fatalf("write csv: %v", err)
	}
	body, err := os.ReadFile(csvPath)
	if err != nil {
		t.Fatalf("read csv: %v", err)
	}
	text := string(body)
	if !strings.Contains(text, "profile_type,profile_id,start_time,end_time,symbol,poc,vah,val,hvn_count,lvn_count,profile_shape,shape_confidence,acceptance_state,profile_volume,notes") {
		t.Fatalf("missing header: %s", text)
	}
	if err := WriteFlexibleVolumeProfileSummaryJSON(jsonPath, summary); err != nil {
		t.Fatalf("write json: %v", err)
	}
	jsonBody, err := os.ReadFile(jsonPath)
	if err != nil {
		t.Fatalf("read json: %v", err)
	}
	var decoded FlexibleVolumeProfileSummary
	if err := json.Unmarshal(jsonBody, &decoded); err != nil {
		t.Fatalf("invalid json: %v", err)
	}
	if decoded.OpenDriveProfiles != 1 || decoded.AcceptedProfiles != 1 || decoded.ShapeCounts[volumeprofile.ShapeDProfile] != 1 {
		t.Fatalf("decoded=%+v", decoded)
	}
}

func researchFlexibleCandles(count int) []exchanges.Candle {
	start := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	candles := make([]exchanges.Candle, 0, count)
	for i := 0; i < count; i++ {
		ts := start.Add(time.Duration(i) * 15 * time.Minute).UnixMilli()
		price := 100 + float64(i)*0.1
		candles = append(candles, exchanges.Candle{
			Venue:     "test",
			Symbol:    "BTCUSDT",
			Interval:  "15m",
			Open:      researchFloat(price),
			High:      researchFloat(price + 1),
			Low:       researchFloat(price - 1),
			Close:     researchFloat(price),
			Volume:    "10",
			StartTime: ts,
			EndTime:   ts + 900000,
		})
	}
	return candles
}

func researchFloat(value float64) string {
	return strconv.FormatFloat(value, 'f', -1, 64)
}
