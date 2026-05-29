package backtest

import (
	"encoding/csv"
	"os"
	"path/filepath"
	"strconv"

	"AlgoTrading2026/series"
)

func WriteSignalCSV(path string, signals series.SignalResult) error {
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

	if err := writer.Write([]string{"time", "close", "diff", "signal", "position"}); err != nil {
		return err
	}

	for i, timestamp := range signals.Times {
		if err := writer.Write([]string{
			strconv.FormatInt(timestamp, 10),
			fmtFloat(signalValue(signals.Close, i)),
			fmtFloat(signalValue(signals.Diff, i)),
			fmtFloat(signalValue(signals.Signal, i)),
			fmtFloat(signalValue(signals.Positions, i)),
		}); err != nil {
			return err
		}
	}

	return writer.Error()
}

func signalValue(values []float64, index int) float64 {
	if index < 0 || index >= len(values) {
		return 0
	}
	return values[index]
}
