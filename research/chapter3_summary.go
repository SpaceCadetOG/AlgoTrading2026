package research

import (
	"encoding/json"
	"math"
	"os"

	"AlgoTrading2026/vwap"
)

type VWAPChapter3Summary struct {
	TotalCandles            int     `json:"totalCandles"`
	BullAlignmentCount      int     `json:"bullAlignmentCount"`
	BearAlignmentCount      int     `json:"bearAlignmentCount"`
	MixedCount              int     `json:"mixedCount"`
	StrongBullCount         int     `json:"strongBullCount"`
	StrongBearCount         int     `json:"strongBearCount"`
	AverageDistanceFromVWAP float64 `json:"averageDistanceFromVWAP"`
	AverageVWAPSlope        float64 `json:"averageVWAPSlope"`
}

func BuildVWAPChapter3Summary(rows []ContextFeatureRow) VWAPChapter3Summary {
	summary := VWAPChapter3Summary{TotalCandles: len(rows)}
	var distanceTotal float64
	var slopeTotal float64

	for _, row := range rows {
		switch row.EMAAlignment {
		case vwap.EMAAlignmentBullish:
			summary.BullAlignmentCount++
		case vwap.EMAAlignmentBearish:
			summary.BearAlignmentCount++
		default:
			summary.MixedCount++
		}
		switch row.TrendAlignment {
		case TrendStrongBull:
			summary.StrongBullCount++
		case TrendStrongBear:
			summary.StrongBearCount++
		}
		distanceTotal += math.Abs(row.DistanceFromVWAP)
		slopeTotal += row.VWAPSlope
	}

	if len(rows) > 0 {
		summary.AverageDistanceFromVWAP = distanceTotal / float64(len(rows))
		summary.AverageVWAPSlope = slopeTotal / float64(len(rows))
	}
	return summary
}

func WriteVWAPChapter3SummaryJSON(path string, summary VWAPChapter3Summary) error {
	if err := ensureDir(path); err != nil {
		return err
	}
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	return encoder.Encode(summary)
}
