package research

import (
	"encoding/csv"
	"os"
	"strconv"

	"AlgoTrading2026/volatility"
)

type Chapter5VolatilityRow struct {
	Time          int64
	Close         float64
	High          float64
	Low           float64
	TrueRange     float64
	ATR           float64
	LogReturn     float64
	RealizedVol   float64
	RollingStdDev float64
	Regime        volatility.Regime
}

func WriteChapter5VolatilityCSV(path string, rows []Chapter5VolatilityRow) error {
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
		"time",
		"close",
		"high",
		"low",
		"true_range",
		"atr",
		"log_return",
		"realized_vol",
		"rolling_stddev",
		"regime",
	}); err != nil {
		return err
	}

	for _, row := range rows {
		if err := writer.Write([]string{
			strconv.FormatInt(row.Time, 10),
			floatToString(row.Close),
			floatToString(row.High),
			floatToString(row.Low),
			floatToString(row.TrueRange),
			floatToString(row.ATR),
			floatToString(row.LogReturn),
			floatToString(row.RealizedVol),
			floatToString(row.RollingStdDev),
			string(row.Regime),
		}); err != nil {
			return err
		}
	}

	return writer.Error()
}
