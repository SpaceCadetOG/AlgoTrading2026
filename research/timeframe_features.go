package research

import (
	"encoding/csv"
	"os"
	"sort"
	"strconv"

	"AlgoTrading2026/exchanges"
	"AlgoTrading2026/vwap"
)

type TimeframeFeatureRow struct {
	Timeframe string
	Timestamp int64
	Close     float64
	EMA9      float64
	EMA20     float64
	VWAP      float64
	Slope     float64
	Alignment vwap.EMAAlignment
	Trend     TrendAlignment
}

func BuildTimeframeFeatureRows(candlesByTimeframe map[string][]exchanges.Candle) []TimeframeFeatureRow {
	timeframes := make([]string, 0, len(candlesByTimeframe))
	for timeframe := range candlesByTimeframe {
		timeframes = append(timeframes, timeframe)
	}
	sort.Strings(timeframes)

	rows := make([]TimeframeFeatureRow, 0)
	for _, timeframe := range timeframes {
		candles := candlesByTimeframe[timeframe]
		close := closeColumn(candles)
		session := vwap.SessionVWAP(candles)
		ema9 := vwap.EMA9(close)
		ema20 := vwap.EMA20(close)
		relationships := vwap.EMARelationships(close)
		slopes := vwap.Slope(session, 5)

		for i, candle := range candles {
			price := close[i]
			vwapValue := valueAt(session, i)
			ema9Value := valueAt(ema9, i)
			ema20Value := valueAt(ema20, i)
			slope := valueAt(slopes, i)
			rows = append(rows, TimeframeFeatureRow{
				Timeframe: timeframe,
				Timestamp: candle.StartTime,
				Close:     price,
				EMA9:      ema9Value,
				EMA20:     ema20Value,
				VWAP:      vwapValue,
				Slope:     slope,
				Alignment: emaRelationshipAt(relationships, i).Alignment,
				Trend:     ClassifyTrendAlignment(price, vwapValue, ema9Value, ema20Value, slope),
			})
		}
	}
	return rows
}

func WriteTimeframeFeaturesCSV(path string, rows []TimeframeFeatureRow) error {
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
		"timeframe",
		"timestamp",
		"close",
		"ema9",
		"ema20",
		"vwap",
		"slope",
		"alignment",
		"trend",
	}); err != nil {
		return err
	}

	for _, row := range rows {
		if err := writer.Write([]string{
			row.Timeframe,
			strconv.FormatInt(row.Timestamp, 10),
			floatToString(row.Close),
			floatToString(row.EMA9),
			floatToString(row.EMA20),
			floatToString(row.VWAP),
			floatToString(row.Slope),
			string(row.Alignment),
			string(row.Trend),
		}); err != nil {
			return err
		}
	}
	return writer.Error()
}
