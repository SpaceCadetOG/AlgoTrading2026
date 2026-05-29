package research

import (
	"encoding/csv"
	"os"
	"strconv"
	"strings"

	"AlgoTrading2026/backtest"
	"AlgoTrading2026/series"
	"AlgoTrading2026/strategy"
)

type Chapter4Analysis = Chapter2Analysis

func AnalyzeChapter4Strategy(name string, report backtest.Report, trades []strategy.Trade, signals series.SignalResult) Chapter4Analysis {
	return AnalyzeChapter2Strategy(name, report, trades, signals)
}

func WriteChapter4StrategyComparisonCSV(path string, rows []StrategyComparisonRow) error {
	return WriteChapter2StrategyComparisonCSV(path, rows)
}

func WriteChapter4AnalysisCSV(path string, rows []Chapter4Analysis) error {
	return WriteChapter2AnalysisCSV(path, rows)
}

func WriteChapter4VenueComparisonCSV(path string, rows []StrategyComparisonRow) error {
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
		"venue",
		"symbol",
		"interval",
		"strategy",
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

	for _, row := range rows {
		report := row.Report
		if err := writer.Write([]string{
			report.Venue,
			report.Symbol,
			report.Interval,
			row.Strategy,
			strconv.Itoa(report.CandleCount),
			strconv.Itoa(report.Metrics.TotalTrades),
			strconv.Itoa(report.Metrics.Wins),
			strconv.Itoa(report.Metrics.Losses),
			floatToString(report.Metrics.WinRate),
			floatToString(report.Metrics.GrossPnL),
			floatToString(report.Metrics.Fees),
			floatToString(report.Metrics.NetPnL),
			floatToString(report.EndingBalance),
			floatToString(report.Metrics.MaxDrawdown),
			floatToString(report.Metrics.MaxDrawdownPct),
		}); err != nil {
			return err
		}
	}

	return writer.Error()
}

func WriteChapter4VenueAnalysisCSV(path string, rows []Chapter4Analysis) error {
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
		"venue",
		"symbol",
		"interval",
		"strategy",
		"candles",
		"trades",
		"wins",
		"losses",
		"win_rate",
		"net_pnl",
		"best_trade_pnl",
		"worst_trade_pnl",
		"average_trade_pnl",
		"average_hold_candles",
		"max_hold_candles",
		"min_hold_candles",
		"signal_count",
		"buy_signal_count",
		"sell_signal_count",
		"hold_signal_count",
		"overtrade_score",
		"notes",
	}); err != nil {
		return err
	}

	for _, row := range rows {
		if err := writer.Write([]string{
			row.Venue,
			row.Symbol,
			row.Interval,
			row.Strategy,
			strconv.Itoa(row.Candles),
			strconv.Itoa(row.Trades),
			strconv.Itoa(row.Wins),
			strconv.Itoa(row.Losses),
			floatToString(row.WinRate),
			floatToString(row.NetPnL),
			floatToString(row.BestTradePnL),
			floatToString(row.WorstTradePnL),
			floatToString(row.AverageTradePnL),
			floatToString(row.AverageHoldCandles),
			strconv.Itoa(row.MaxHoldCandles),
			strconv.Itoa(row.MinHoldCandles),
			strconv.Itoa(row.SignalCount),
			strconv.Itoa(row.BuySignalCount),
			strconv.Itoa(row.SellSignalCount),
			strconv.Itoa(row.HoldSignalCount),
			floatToString(row.OvertradeScore),
			strings.Join(row.Notes, ","),
		}); err != nil {
			return err
		}
	}

	return writer.Error()
}
