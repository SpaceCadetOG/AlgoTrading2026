package backtest

import (
	"encoding/csv"
	"os"
	"path/filepath"
	"strconv"
)

func WriteEquityCSV(path string, curve EquityCurve) error {
	if dir := filepath.Dir(path); dir != "." && dir != "" {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return err
		}
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
		"signal",
		"position",
		"holdings",
		"cash",
		"total_equity",
		"unrealized_pnl",
		"realized_pnl",
	}); err != nil {
		return err
	}

	for _, point := range curve.Points {
		if err := writer.Write([]string{
			strconv.FormatInt(point.Time, 10),
			fmtFloat(point.Close),
			fmtFloat(point.Signal),
			fmtFloat(point.Position),
			fmtFloat(point.Holdings),
			fmtFloat(point.Cash),
			fmtFloat(point.TotalEquity),
			fmtFloat(point.UnrealizedPnL),
			fmtFloat(point.RealizedPnL),
		}); err != nil {
			return err
		}
	}

	return writer.Error()
}
