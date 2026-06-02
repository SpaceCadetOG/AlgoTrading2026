package research

import (
	"encoding/csv"
	"os"
	"strconv"

	"AlgoTrading2026/volumeprofile"
)

type VolumeProfileShapeStudyRow struct {
	Scope               string
	ProfileStart        int64
	ProfileEnd          int64
	Symbol              string
	POC                 float64
	VAH                 float64
	VAL                 float64
	ProfileShape        string
	TotalBins           int
	TotalVolume         float64
	UpperVolumePct      float64
	LowerVolumePct      float64
	HVNCount            int
	LVNCount            int
	DistributionBalance float64
	ShapeConfidence     float64
	ShapeReason         string
}

func BuildVolumeProfileShapeStudyRows(scoped []volumeprofile.ScopedProfile) []VolumeProfileShapeStudyRow {
	rows := make([]VolumeProfileShapeStudyRow, 0, len(scoped))
	for _, item := range scoped {
		profile := item.Profile
		study := volumeprofile.AnalyzeProfileShape(profile)
		rows = append(rows, VolumeProfileShapeStudyRow{
			Scope:               item.Scope,
			ProfileStart:        profile.StartTime,
			ProfileEnd:          profile.EndTime,
			Symbol:              profile.Symbol,
			POC:                 profile.POC,
			VAH:                 profile.VAH,
			VAL:                 profile.VAL,
			ProfileShape:        study.ProfileShape,
			TotalBins:           len(profile.Bins),
			TotalVolume:         profile.TotalVolume,
			UpperVolumePct:      study.UpperVolumePct,
			LowerVolumePct:      study.LowerVolumePct,
			HVNCount:            len(profile.HVNs),
			LVNCount:            len(profile.LVNs),
			DistributionBalance: study.DistributionBalance,
			ShapeConfidence:     study.ShapeConfidence,
			ShapeReason:         study.ShapeReason,
		})
	}
	return rows
}

func WriteVolumeProfileShapeStudyCSV(path string, rows []VolumeProfileShapeStudyRow) error {
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
		"profile_start",
		"profile_end",
		"symbol",
		"poc",
		"vah",
		"val",
		"profile_shape",
		"total_bins",
		"total_volume",
		"upper_volume_pct",
		"lower_volume_pct",
		"hvn_count",
		"lvn_count",
		"distribution_balance",
		"shape_confidence",
		"shape_reason",
	}); err != nil {
		return err
	}
	for _, row := range rows {
		if err := writer.Write([]string{
			row.Scope,
			strconv.FormatInt(row.ProfileStart, 10),
			strconv.FormatInt(row.ProfileEnd, 10),
			row.Symbol,
			floatToString(row.POC),
			floatToString(row.VAH),
			floatToString(row.VAL),
			row.ProfileShape,
			strconv.Itoa(row.TotalBins),
			floatToString(row.TotalVolume),
			floatToString(row.UpperVolumePct),
			floatToString(row.LowerVolumePct),
			strconv.Itoa(row.HVNCount),
			strconv.Itoa(row.LVNCount),
			floatToString(row.DistributionBalance),
			floatToString(row.ShapeConfidence),
			row.ShapeReason,
		}); err != nil {
			return err
		}
	}
	return writer.Error()
}
