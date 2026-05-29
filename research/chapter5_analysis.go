package research

import (
	"encoding/csv"
	"os"
	"strconv"
)

type Chapter5Analysis struct {
	Strategy string
	Venue    string
	Symbol   string
	Interval string

	Candles int

	Trades  int
	Wins    int
	Losses  int
	WinRate float64

	GrossPnL float64
	Fees     float64
	NetPnL   float64

	EndingBalance float64

	MaxDrawdown    float64
	MaxDrawdownPct float64

	AverageTrade float64
	AverageHold  float64

	SignalCount int

	OvertradeScore float64

	RegimeLowCount    int
	RegimeNormalCount int
	RegimeHighCount   int
}

func WriteChapter5StrategyComparisonCSV(path string, rows []Chapter5Analysis) error {
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

	if err := writer.Write(chapter5Header()); err != nil {
		return err
	}
	for _, row := range rows {
		if err := writer.Write(chapter5Row(row)); err != nil {
			return err
		}
	}
	return writer.Error()
}

func WriteChapter5AnalysisCSV(path string, rows []Chapter5Analysis) error {
	return WriteChapter5StrategyComparisonCSV(path, rows)
}

func chapter5Header() []string {
	return []string{
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
		"average_trade",
		"average_hold",
		"signal_count",
		"overtrade_score",
		"regime_low_count",
		"regime_normal_count",
		"regime_high_count",
	}
}

func chapter5Row(row Chapter5Analysis) []string {
	return []string{
		row.Strategy,
		row.Venue,
		row.Symbol,
		row.Interval,
		strconv.Itoa(row.Candles),
		strconv.Itoa(row.Trades),
		strconv.Itoa(row.Wins),
		strconv.Itoa(row.Losses),
		floatToString(row.WinRate),
		floatToString(row.GrossPnL),
		floatToString(row.Fees),
		floatToString(row.NetPnL),
		floatToString(row.EndingBalance),
		floatToString(row.MaxDrawdown),
		floatToString(row.MaxDrawdownPct),
		floatToString(row.AverageTrade),
		floatToString(row.AverageHold),
		strconv.Itoa(row.SignalCount),
		floatToString(row.OvertradeScore),
		strconv.Itoa(row.RegimeLowCount),
		strconv.Itoa(row.RegimeNormalCount),
		strconv.Itoa(row.RegimeHighCount),
	}
}
