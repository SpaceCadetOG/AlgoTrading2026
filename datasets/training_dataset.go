package datasets

import (
	"AlgoTrading2026/features"
	"AlgoTrading2026/labels"
)

type TrainingRow struct {
	features.FeatureRow
	FutureReturn float64
	Direction    int
	TPSL         int
}

func BuildTrainingDataset(featureRows []features.FeatureRow, labelRows []labels.LabelRow) []TrainingRow {
	labelsByTime := make(map[int64]labels.LabelRow, len(labelRows))
	for _, row := range labelRows {
		labelsByTime[row.Time] = row
	}

	rows := make([]TrainingRow, 0, len(featureRows))
	for _, featureRow := range featureRows {
		labelRow, ok := labelsByTime[featureRow.Time]
		if !ok {
			continue
		}
		rows = append(rows, TrainingRow{
			FeatureRow:   featureRow,
			FutureReturn: labelRow.FutureReturn,
			Direction:    labelRow.Direction,
			TPSL:         labelRow.TPSL,
		})
	}
	return rows
}
