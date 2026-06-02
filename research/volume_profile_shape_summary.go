package research

import (
	"encoding/json"
	"os"
)

type VolumeProfileShapeSummaryItem struct {
	Shape      string  `json:"shape"`
	Confidence float64 `json:"confidence"`
}

type VolumeProfileShapeSummary struct {
	Daily        VolumeProfileShapeSummaryItem `json:"daily"`
	Rolling3D    VolumeProfileShapeSummaryItem `json:"rolling3d"`
	Rolling7D    VolumeProfileShapeSummaryItem `json:"rolling7d"`
	Composite30D VolumeProfileShapeSummaryItem `json:"composite30d"`
}

func BuildVolumeProfileShapeSummary(rows []VolumeProfileShapeStudyRow) VolumeProfileShapeSummary {
	var summary VolumeProfileShapeSummary
	for _, row := range rows {
		item := VolumeProfileShapeSummaryItem{Shape: row.ProfileShape, Confidence: row.ShapeConfidence}
		switch row.Scope {
		case "daily_session":
			summary.Daily = item
		case "rolling_3d":
			summary.Rolling3D = item
		case "rolling_7d":
			summary.Rolling7D = item
		case "composite_30d":
			summary.Composite30D = item
		}
	}
	return summary
}

func WriteVolumeProfileShapeSummaryJSON(path string, summary VolumeProfileShapeSummary) error {
	if err := ensureDir(path); err != nil {
		return err
	}
	body, err := json.MarshalIndent(summary, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, append(body, '\n'), 0644)
}
