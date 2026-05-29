package backtest

import (
	"encoding/csv"
	"os"
	"path/filepath"
	"strconv"

	"AlgoTrading2026/strategy"
)

func WriteTradesCSV(path string, trades []strategy.Trade) error {
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
		"trade_id",
		"venue",
		"symbol",
		"side",
		"entry_time",
		"exit_time",
		"entry_price",
		"exit_price",
		"size",
		"gross_pnl",
		"fees",
		"net_pnl",
		"exit_reason",
	}); err != nil {
		return err
	}

	for i, trade := range trades {
		if err := writer.Write([]string{
			strconv.Itoa(i + 1),
			trade.Venue,
			trade.Symbol,
			trade.Side,
			trade.OpenedAt.Format("2006-01-02T15:04:05Z07:00"),
			trade.ClosedAt.Format("2006-01-02T15:04:05Z07:00"),
			fmtFloat(trade.EntryPrice),
			fmtFloat(trade.ExitPrice),
			fmtFloat(trade.Size),
			fmtFloat(trade.GrossPnL),
			fmtFloat(trade.Fees),
			fmtFloat(trade.RealizedPnL),
			trade.ExitReason,
		}); err != nil {
			return err
		}
	}

	return writer.Error()
}

func fmtFloat(value float64) string {
	return strconv.FormatFloat(value, 'f', 8, 64)
}
