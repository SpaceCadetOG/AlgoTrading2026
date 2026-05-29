package datasets

import (
	"encoding/csv"
	"os"
	"path/filepath"
	"strconv"
)

func WriteCSV(path string, rows []TrainingRow) error {
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
		"time", "venue", "symbol", "interval",
		"close", "volume",
		"sma", "ema", "apo",
		"macd", "macd_signal", "macd_histogram",
		"rsi", "momentum", "stddev",
		"bollinger_middle", "bollinger_upper", "bollinger_lower", "bollinger_width",
		"distance_from_sma", "distance_from_ema",
		"support", "resistance", "distance_from_support", "distance_from_resistance",
		"hour_of_day", "day_of_week",
		"future_return", "direction", "tpsl",
	}); err != nil {
		return err
	}

	for _, row := range rows {
		if err := writer.Write([]string{
			strconv.FormatInt(row.Time, 10),
			row.Venue,
			row.Symbol,
			row.Interval,
			fmtFloat(row.Close),
			fmtFloat(row.Volume),
			fmtFloat(row.SMA),
			fmtFloat(row.EMA),
			fmtFloat(row.APO),
			fmtFloat(row.MACD),
			fmtFloat(row.MACDSignal),
			fmtFloat(row.MACDHistogram),
			fmtFloat(row.RSI),
			fmtFloat(row.Momentum),
			fmtFloat(row.StdDev),
			fmtFloat(row.BollingerMiddle),
			fmtFloat(row.BollingerUpper),
			fmtFloat(row.BollingerLower),
			fmtFloat(row.BollingerWidth),
			fmtFloat(row.DistanceFromSMA),
			fmtFloat(row.DistanceFromEMA),
			fmtFloat(row.Support),
			fmtFloat(row.Resistance),
			fmtFloat(row.DistanceFromSupport),
			fmtFloat(row.DistanceFromResistance),
			strconv.Itoa(row.HourOfDay),
			strconv.Itoa(row.DayOfWeek),
			fmtFloat(row.FutureReturn),
			strconv.Itoa(row.Direction),
			strconv.Itoa(row.TPSL),
		}); err != nil {
			return err
		}
	}

	return writer.Error()
}

func ensureDir(path string) error {
	dir := filepath.Dir(path)
	if dir == "." || dir == "" {
		return nil
	}
	return os.MkdirAll(dir, 0755)
}

func fmtFloat(value float64) string {
	return strconv.FormatFloat(value, 'f', 8, 64)
}
