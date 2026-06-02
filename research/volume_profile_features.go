package research

import (
	"encoding/csv"
	"os"
	"strconv"

	"AlgoTrading2026/volumeprofile"
)

type VolumeProfileFeatureRow struct {
	Timestamp    int64
	Venue        string
	Symbol       string
	ProfileStart int64
	ProfileEnd   int64
	BinLow       float64
	BinHigh      float64
	BinMid       float64
	BinVolume    float64
	POC          float64
	VAH          float64
	VAL          float64
	IsPOC        bool
	IsHVN        bool
	IsLVN        bool
	ProfileShape string
	TotalVolume  float64
	Notes        string
}

func BuildVolumeProfileFeatureRows(profile volumeprofile.VolumeProfile) []VolumeProfileFeatureRow {
	rows := make([]VolumeProfileFeatureRow, 0, len(profile.Bins))
	for _, bin := range profile.Bins {
		rows = append(rows, VolumeProfileFeatureRow{
			Timestamp:    profile.EndTime,
			Venue:        profile.Venue,
			Symbol:       profile.Symbol,
			ProfileStart: profile.StartTime,
			ProfileEnd:   profile.EndTime,
			BinLow:       bin.Low,
			BinHigh:      bin.High,
			BinMid:       bin.Mid,
			BinVolume:    bin.Volume,
			POC:          profile.POC,
			VAH:          profile.VAH,
			VAL:          profile.VAL,
			IsPOC:        bin.Mid == profile.POC,
			IsHVN:        volumeprofile.IsNode(bin, profile.HVNs),
			IsLVN:        volumeprofile.IsNode(bin, profile.LVNs),
			ProfileShape: profile.Shape,
			TotalVolume:  profile.TotalVolume,
			Notes:        "OHLCV candle approximation; volume distributed equally across candle high-low bins.",
		})
	}
	return rows
}

func WriteVolumeProfileFeaturesCSV(path string, rows []VolumeProfileFeatureRow) error {
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
		"timestamp",
		"venue",
		"symbol",
		"profile_start",
		"profile_end",
		"bin_low",
		"bin_high",
		"bin_mid",
		"bin_volume",
		"poc",
		"vah",
		"val",
		"is_poc",
		"is_hvn",
		"is_lvn",
		"profile_shape",
		"total_volume",
		"notes",
	}); err != nil {
		return err
	}
	for _, row := range rows {
		if err := writer.Write([]string{
			strconv.FormatInt(row.Timestamp, 10),
			row.Venue,
			row.Symbol,
			strconv.FormatInt(row.ProfileStart, 10),
			strconv.FormatInt(row.ProfileEnd, 10),
			floatToString(row.BinLow),
			floatToString(row.BinHigh),
			floatToString(row.BinMid),
			floatToString(row.BinVolume),
			floatToString(row.POC),
			floatToString(row.VAH),
			floatToString(row.VAL),
			strconv.FormatBool(row.IsPOC),
			strconv.FormatBool(row.IsHVN),
			strconv.FormatBool(row.IsLVN),
			row.ProfileShape,
			floatToString(row.TotalVolume),
			row.Notes,
		}); err != nil {
			return err
		}
	}
	return writer.Error()
}
