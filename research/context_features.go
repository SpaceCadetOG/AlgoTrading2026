package research

import (
	"encoding/csv"
	"os"
	"strconv"

	"AlgoTrading2026/exchanges"
	"AlgoTrading2026/vwap"
)

type ContextFeatureRow struct {
	Timestamp        int64
	Close            float64
	SessionVWAP      float64
	DistanceFromVWAP float64
	EMA9             float64
	EMA20            float64
	PriceVsVWAP      float64
	PriceVsEMA9      float64
	PriceVsEMA20     float64
	EMAAlignment     vwap.EMAAlignment
	VWAPSlope        float64
	TrendAlignment   TrendAlignment
}

func BuildContextFeatureRows(candles []exchanges.Candle) []ContextFeatureRow {
	close := closeColumn(candles)
	session := vwap.SessionVWAP(candles)
	distance := vwap.Distance(close, session)
	ema9 := vwap.EMA9(close)
	ema20 := vwap.EMA20(close)
	relationships := vwap.EMARelationships(close)
	slopes := vwap.Slope(session, 5)

	rows := make([]ContextFeatureRow, 0, len(candles))
	for i, candle := range candles {
		price := close[i]
		sessionValue := valueAt(session, i)
		ema9Value := valueAt(ema9, i)
		ema20Value := valueAt(ema20, i)
		slope := valueAt(slopes, i)
		relationship := emaRelationshipAt(relationships, i)
		rows = append(rows, ContextFeatureRow{
			Timestamp:        candle.StartTime,
			Close:            price,
			SessionVWAP:      sessionValue,
			DistanceFromVWAP: valueAt(distance, i),
			EMA9:             ema9Value,
			EMA20:            ema20Value,
			PriceVsVWAP:      price - sessionValue,
			PriceVsEMA9:      relationship.PriceVsEMA9,
			PriceVsEMA20:     relationship.PriceVsEMA20,
			EMAAlignment:     relationship.Alignment,
			VWAPSlope:        slope,
			TrendAlignment:   ClassifyTrendAlignment(price, sessionValue, ema9Value, ema20Value, slope),
		})
	}
	return rows
}

func WriteContextFeaturesCSV(path string, rows []ContextFeatureRow) error {
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
		"close",
		"session_vwap",
		"distance_from_vwap",
		"ema9",
		"ema20",
		"price_vs_vwap",
		"price_vs_ema9",
		"price_vs_ema20",
		"ema_alignment",
		"vwap_slope",
		"trend_alignment",
	}); err != nil {
		return err
	}

	for _, row := range rows {
		if err := writer.Write([]string{
			strconv.FormatInt(row.Timestamp, 10),
			floatToString(row.Close),
			floatToString(row.SessionVWAP),
			floatToString(row.DistanceFromVWAP),
			floatToString(row.EMA9),
			floatToString(row.EMA20),
			floatToString(row.PriceVsVWAP),
			floatToString(row.PriceVsEMA9),
			floatToString(row.PriceVsEMA20),
			string(row.EMAAlignment),
			floatToString(row.VWAPSlope),
			string(row.TrendAlignment),
		}); err != nil {
			return err
		}
	}
	return writer.Error()
}

func emaRelationshipAt(values []vwap.EMARelationship, i int) vwap.EMARelationship {
	if i < 0 || i >= len(values) {
		return vwap.EMARelationship{Alignment: vwap.EMAAlignmentMixed}
	}
	return values[i]
}
