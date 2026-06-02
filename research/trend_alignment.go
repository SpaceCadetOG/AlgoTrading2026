package research

import (
	"encoding/csv"
	"os"
	"strconv"

	"AlgoTrading2026/exchanges"
	"AlgoTrading2026/vwap"
)

type TrendAlignment string

const (
	TrendStrongBull TrendAlignment = "strong_bull"
	TrendBull       TrendAlignment = "bull"
	TrendNeutral    TrendAlignment = "neutral"
	TrendBear       TrendAlignment = "bear"
	TrendStrongBear TrendAlignment = "strong_bear"
)

type TrendAlignmentRow struct {
	Timestamp      int64
	Close          float64
	SessionVWAP    float64
	EMA9           float64
	EMA20          float64
	VWAPSlope      float64
	TrendAlignment TrendAlignment
}

func BuildTrendAlignmentRows(candles []exchanges.Candle) []TrendAlignmentRow {
	close := closeColumn(candles)
	session := vwap.SessionVWAP(candles)
	ema9 := vwap.EMA9(close)
	ema20 := vwap.EMA20(close)
	slopes := vwap.Slope(session, 5)

	rows := make([]TrendAlignmentRow, 0, len(candles))
	for i, candle := range candles {
		rows = append(rows, TrendAlignmentRow{
			Timestamp:      candle.StartTime,
			Close:          close[i],
			SessionVWAP:    valueAt(session, i),
			EMA9:           valueAt(ema9, i),
			EMA20:          valueAt(ema20, i),
			VWAPSlope:      valueAt(slopes, i),
			TrendAlignment: ClassifyTrendAlignment(close[i], valueAt(session, i), valueAt(ema9, i), valueAt(ema20, i), valueAt(slopes, i)),
		})
	}
	return rows
}

func ClassifyTrendAlignment(price float64, sessionVWAP float64, ema9 float64, ema20 float64, vwapSlope float64) TrendAlignment {
	switch {
	case price > sessionVWAP && price > ema9 && ema9 > ema20 && vwapSlope > 0:
		return TrendStrongBull
	case price > sessionVWAP && price > ema20:
		return TrendBull
	case price < sessionVWAP && price < ema9 && ema9 < ema20 && vwapSlope < 0:
		return TrendStrongBear
	case price < sessionVWAP && price < ema20:
		return TrendBear
	default:
		return TrendNeutral
	}
}

func WriteTrendAlignmentCSV(path string, rows []TrendAlignmentRow) error {
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
		"ema9",
		"ema20",
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
			floatToString(row.EMA9),
			floatToString(row.EMA20),
			floatToString(row.VWAPSlope),
			string(row.TrendAlignment),
		}); err != nil {
			return err
		}
	}
	return writer.Error()
}

func closeColumn(candles []exchanges.Candle) []float64 {
	close := make([]float64, len(candles))
	for i, candle := range candles {
		close[i] = candle.CloseFloat()
	}
	return close
}
