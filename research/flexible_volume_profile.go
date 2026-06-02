package research

import (
	"encoding/csv"
	"fmt"
	"os"
	"strconv"

	"AlgoTrading2026/exchanges"
	"AlgoTrading2026/volumeprofile"
)

type FlexibleVolumeProfileRow struct {
	ProfileType     string
	ProfileID       string
	StartTime       int64
	EndTime         int64
	Symbol          string
	POC             float64
	VAH             float64
	VAL             float64
	HVNCount        int
	LVNCount        int
	ProfileShape    string
	ShapeConfidence float64
	AcceptanceState string
	ProfileVolume   float64
	Notes           string
}

func BuildFlexibleVolumeProfiles(candles []exchanges.Candle, priceRows []PriceActionFeatureRow, strategyRows []PriceActionStrategyStudyRow, phase3Rows []PriceActionPhase3StudyRow) []volumeprofile.FlexibleProfile {
	windows := make([]volumeprofile.FlexibleProfileWindow, 0)
	windows = append(windows, sidewaysWindows(candles, priceRows)...)
	windows = append(windows, openDriveWindows(candles, strategyRows)...)
	windows = append(windows, phase3Windows(candles, phase3Rows)...)
	return volumeprofile.BuildFlexibleProfiles(candles, windows, volumeprofile.DefaultBinConfig())
}

func BuildFlexibleVolumeProfileRows(profiles []volumeprofile.FlexibleProfile) []FlexibleVolumeProfileRow {
	rows := make([]FlexibleVolumeProfileRow, 0, len(profiles))
	for _, item := range profiles {
		profile := item.Profile
		rows = append(rows, FlexibleVolumeProfileRow{
			ProfileType:     item.ProfileType,
			ProfileID:       item.ProfileID,
			StartTime:       item.StartTime,
			EndTime:         item.EndTime,
			Symbol:          profile.Symbol,
			POC:             profile.POC,
			VAH:             profile.VAH,
			VAL:             profile.VAL,
			HVNCount:        len(profile.HVNs),
			LVNCount:        len(profile.LVNs),
			ProfileShape:    item.ShapeStudy.ProfileShape,
			ShapeConfidence: item.ShapeStudy.ShapeConfidence,
			AcceptanceState: item.AcceptanceState,
			ProfileVolume:   profile.TotalVolume,
			Notes:           "Flexible event-anchored profile using OHLCV candle approximation.",
		})
	}
	return rows
}

func WriteFlexibleVolumeProfileCSV(path string, rows []FlexibleVolumeProfileRow) error {
	if err := ensureDir(path); err != nil {
		return err
	}
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()
	if err := writer.Write([]string{
		"profile_type",
		"profile_id",
		"start_time",
		"end_time",
		"symbol",
		"poc",
		"vah",
		"val",
		"hvn_count",
		"lvn_count",
		"profile_shape",
		"shape_confidence",
		"acceptance_state",
		"profile_volume",
		"notes",
	}); err != nil {
		return err
	}
	for _, row := range rows {
		if err := writer.Write([]string{
			row.ProfileType,
			row.ProfileID,
			strconv.FormatInt(row.StartTime, 10),
			strconv.FormatInt(row.EndTime, 10),
			row.Symbol,
			floatToString(row.POC),
			floatToString(row.VAH),
			floatToString(row.VAL),
			strconv.Itoa(row.HVNCount),
			strconv.Itoa(row.LVNCount),
			row.ProfileShape,
			floatToString(row.ShapeConfidence),
			row.AcceptanceState,
			floatToString(row.ProfileVolume),
			row.Notes,
		}); err != nil {
			return err
		}
	}
	return writer.Error()
}

func sidewaysWindows(candles []exchanges.Candle, rows []PriceActionFeatureRow) []volumeprofile.FlexibleProfileWindow {
	windows := make([]volumeprofile.FlexibleProfileWindow, 0)
	count := 0
	for i, row := range rows {
		if !row.SidewaysDetected {
			continue
		}
		nextDetected := i+1 < len(rows) && rows[i+1].SidewaysDetected
		if nextDetected {
			continue
		}
		startIndex := i - row.SidewaysDuration + 1
		if startIndex < 0 {
			startIndex = 0
		}
		if i >= len(candles) || startIndex >= len(candles) {
			continue
		}
		count++
		windows = append(windows, volumeprofile.FlexibleProfileWindow{
			ProfileType: volumeprofile.FlexibleSidewaysAccumulation,
			ProfileID:   fmt.Sprintf("sideways_accumulation_%d", count),
			StartTime:   candles[startIndex].StartTime,
			EndTime:     candles[i].StartTime,
		})
	}
	return windows
}

func openDriveWindows(candles []exchanges.Candle, rows []PriceActionStrategyStudyRow) []volumeprofile.FlexibleProfileWindow {
	windows := make([]volumeprofile.FlexibleProfileWindow, 0)
	count := 0
	for _, row := range rows {
		if row.Strategy != "open_drive" {
			continue
		}
		index := candleIndexByTime(candles, row.Timestamp)
		if index < 0 {
			continue
		}
		start, end := eventWindow(candles, index, 4, 8)
		count++
		windows = append(windows, volumeprofile.FlexibleProfileWindow{
			ProfileType: volumeprofile.FlexibleOpenDrive,
			ProfileID:   fmt.Sprintf("open_drive_%d", count),
			StartTime:   start,
			EndTime:     end,
		})
	}
	return windows
}

func phase3Windows(candles []exchanges.Candle, rows []PriceActionPhase3StudyRow) []volumeprofile.FlexibleProfileWindow {
	windows := make([]volumeprofile.FlexibleProfileWindow, 0)
	failedCount := 0
	retestCount := 0
	for _, row := range rows {
		index := candleIndexByTime(candles, row.Timestamp)
		if index < 0 {
			continue
		}
		start, end := eventWindow(candles, index, 8, 8)
		switch row.Strategy {
		case "failed_high_auction", "failed_low_auction":
			failedCount++
			windows = append(windows, volumeprofile.FlexibleProfileWindow{
				ProfileType: volumeprofile.FlexibleFailedAuction,
				ProfileID:   fmt.Sprintf("%s_%d", row.Strategy, failedCount),
				StartTime:   start,
				EndTime:     end,
			})
		case "daily_high_retest", "daily_low_retest", "weekly_high_retest", "weekly_low_retest":
			retestCount++
			windows = append(windows, volumeprofile.FlexibleProfileWindow{
				ProfileType: volumeprofile.FlexibleHighLowRetest,
				ProfileID:   fmt.Sprintf("%s_%d", row.Strategy, retestCount),
				StartTime:   start,
				EndTime:     end,
			})
		}
	}
	return windows
}

func candleIndexByTime(candles []exchanges.Candle, timestamp int64) int {
	for i, candle := range candles {
		if candle.StartTime == timestamp {
			return i
		}
	}
	return -1
}

func eventWindow(candles []exchanges.Candle, index int, before int, after int) (int64, int64) {
	start := index - before
	if start < 0 {
		start = 0
	}
	end := index + after
	if end >= len(candles) {
		end = len(candles) - 1
	}
	return candles[start].StartTime, candles[end].StartTime
}
