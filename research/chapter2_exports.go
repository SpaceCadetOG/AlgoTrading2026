package research

import (
	"encoding/csv"
	"os"
	"path/filepath"
	"strconv"

	"AlgoTrading2026/backtest"
)

type Chapter2Indicators struct {
	Time       []int64
	Close      []float64
	SMA        []float64
	EMA        []float64
	APO        []float64
	MACD       []float64
	MACDSignal []float64
	MACDHist   []float64
	BBMiddle   []float64
	BBUpper    []float64
	BBLower    []float64
	RSI        []float64
	StdDev     []float64
	Momentum   []float64
	Support    []float64
	Resistance []float64
	Hour       []int
	DayOfWeek  []int
}

type StrategyComparisonRow struct {
	Strategy string
	Report   backtest.Report
}

func WriteChapter2IndicatorsCSV(path string, values Chapter2Indicators) error {
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
		"sma",
		"ema",
		"apo",
		"macd",
		"macd_signal",
		"macd_histogram",
		"bb_middle",
		"bb_upper",
		"bb_lower",
		"rsi",
		"stddev",
		"momentum",
		"support",
		"resistance",
		"hour",
		"day_of_week",
	}); err != nil {
		return err
	}

	for i, timestamp := range values.Time {
		if err := writer.Write([]string{
			strconv.FormatInt(timestamp, 10),
			floatValue(values.Close, i),
			floatValue(values.SMA, i),
			floatValue(values.EMA, i),
			floatValue(values.APO, i),
			floatValue(values.MACD, i),
			floatValue(values.MACDSignal, i),
			floatValue(values.MACDHist, i),
			floatValue(values.BBMiddle, i),
			floatValue(values.BBUpper, i),
			floatValue(values.BBLower, i),
			floatValue(values.RSI, i),
			floatValue(values.StdDev, i),
			floatValue(values.Momentum, i),
			floatValue(values.Support, i),
			floatValue(values.Resistance, i),
			intValue(values.Hour, i),
			intValue(values.DayOfWeek, i),
		}); err != nil {
			return err
		}
	}

	return writer.Error()
}

func WriteChapter2StrategyComparisonCSV(path string, rows []StrategyComparisonRow) error {
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
		"strategy",
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

	for _, row := range rows {
		report := row.Report
		if err := writer.Write([]string{
			row.Strategy,
			report.Venue,
			report.Symbol,
			report.Interval,
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

func ensureDir(path string) error {
	dir := filepath.Dir(path)
	if dir == "." || dir == "" {
		return nil
	}
	return os.MkdirAll(dir, 0755)
}

func floatValue(values []float64, index int) string {
	if index < 0 || index >= len(values) {
		return floatToString(0)
	}
	return floatToString(values[index])
}

func intValue(values []int, index int) string {
	if index < 0 || index >= len(values) {
		return "0"
	}
	return strconv.Itoa(values[index])
}
