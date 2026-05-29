package backtest

import (
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"AlgoTrading2026/exchanges"
)

type VenueComparisonInput struct {
	Venue    string
	Symbol   string
	Interval string
	Limit    int
	Reader   exchanges.CandleReader
	Config   Config
}

type VenueComparisonResult struct {
	Venue  string
	Symbol string
	Report Report
	Error  string
}

func RunVenueComparison(inputs []VenueComparisonInput) []VenueComparisonResult {
	results := make([]VenueComparisonResult, 0, len(inputs))
	for _, input := range inputs {
		result := VenueComparisonResult{Venue: input.Venue, Symbol: input.Symbol}
		if input.Reader == nil {
			result.Error = "missing candle reader"
			results = append(results, result)
			continue
		}

		candles, err := input.Reader.GetCandles(input.Symbol, input.Interval, input.Limit)
		if err != nil {
			result.Error = err.Error()
			results = append(results, result)
			continue
		}

		cfg := input.Config
		if cfg.Venue == "" {
			cfg.Venue = input.Venue
		}
		if cfg.Symbol == "" {
			cfg.Symbol = input.Symbol
		}
		if cfg.Interval == "" {
			cfg.Interval = input.Interval
		}
		if cfg.StartingBalance == 0 {
			cfg.StartingBalance = 10000
		}
		if cfg.FixedNotional == 0 {
			cfg.FixedNotional = 100
		}

		result.Report = RunBookSignalBacktest(candles, cfg)
		results = append(results, result)
	}

	return results
}

func WriteVenueComparisonCSV(path string, results []VenueComparisonResult) error {
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
		"venue",
		"symbol",
		"interval",
		"candles",
		"trades",
		"wins",
		"losses",
		"win_rate",
		"gross_pnl",
		"fees",
		"net_pnl",
		"ending_balance",
		"max_drawdown",
		"max_drawdown_pct",
	}); err != nil {
		return err
	}

	for _, result := range results {
		if result.Error != "" {
			continue
		}
		report := result.Report
		if err := writer.Write([]string{
			report.Venue,
			report.Symbol,
			report.Interval,
			strconv.Itoa(report.CandleCount),
			strconv.Itoa(report.Metrics.TotalTrades),
			strconv.Itoa(report.Metrics.Wins),
			strconv.Itoa(report.Metrics.Losses),
			fmtFloat(report.Metrics.WinRate),
			fmtFloat(report.Metrics.GrossPnL),
			fmtFloat(report.Metrics.Fees),
			fmtFloat(report.Metrics.NetPnL),
			fmtFloat(report.EndingBalance),
			fmtFloat(report.Metrics.MaxDrawdown),
			fmtFloat(report.Metrics.MaxDrawdownPct),
		}); err != nil {
			return fmt.Errorf("write comparison row for %s: %w", result.Venue, err)
		}
	}

	return writer.Error()
}
