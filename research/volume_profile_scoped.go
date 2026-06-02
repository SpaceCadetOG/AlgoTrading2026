package research

import (
	"encoding/csv"
	"os"
	"strconv"

	"AlgoTrading2026/volumeprofile"
)

type ScopedVolumeProfileFeatureRow struct {
	Scope string
	VolumeProfileFeatureRow
}

func BuildScopedVolumeProfileFeatureRows(scoped []volumeprofile.ScopedProfile) []ScopedVolumeProfileFeatureRow {
	rows := make([]ScopedVolumeProfileFeatureRow, 0)
	for _, item := range scoped {
		for _, row := range BuildVolumeProfileFeatureRows(item.Profile) {
			rows = append(rows, ScopedVolumeProfileFeatureRow{
				Scope:                   item.Scope,
				VolumeProfileFeatureRow: row,
			})
		}
	}
	return rows
}

func WriteScopedVolumeProfileFeaturesCSV(path string, rows []ScopedVolumeProfileFeatureRow) error {
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
		"scope",
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
			row.Scope,
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
