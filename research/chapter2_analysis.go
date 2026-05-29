package research

import (
	"encoding/csv"
	"encoding/json"
	"math"
	"os"
	"strconv"
	"strings"
	"time"

	"AlgoTrading2026/backtest"
	"AlgoTrading2026/series"
	"AlgoTrading2026/strategy"
)

type Chapter2Analysis struct {
	Strategy string `json:"strategy"`
	Venue    string `json:"venue"`
	Symbol   string `json:"symbol"`
	Interval string `json:"interval"`
	Candles  int    `json:"candles"`

	Trades  int     `json:"trades"`
	Wins    int     `json:"wins"`
	Losses  int     `json:"losses"`
	WinRate float64 `json:"win_rate"`

	GrossPnL float64 `json:"gross_pnl"`
	Fees     float64 `json:"fees"`
	NetPnL   float64 `json:"net_pnl"`

	MaxDrawdown    float64 `json:"max_drawdown"`
	MaxDrawdownPct float64 `json:"max_drawdown_pct"`

	BestTradePnL    float64 `json:"best_trade_pnl"`
	WorstTradePnL   float64 `json:"worst_trade_pnl"`
	AverageTradePnL float64 `json:"average_trade_pnl"`

	AverageHoldCandles float64 `json:"average_hold_candles"`
	MaxHoldCandles     int     `json:"max_hold_candles"`
	MinHoldCandles     int     `json:"min_hold_candles"`

	SignalCount     int `json:"signal_count"`
	BuySignalCount  int `json:"buy_signal_count"`
	SellSignalCount int `json:"sell_signal_count"`
	HoldSignalCount int `json:"hold_signal_count"`

	OvertradeScore float64  `json:"overtrade_score"`
	Notes          []string `json:"notes"`
}

func AnalyzeChapter2Strategy(name string, report backtest.Report, trades []strategy.Trade, signals series.SignalResult) Chapter2Analysis {
	analysis := Chapter2Analysis{
		Strategy:       name,
		Venue:          report.Venue,
		Symbol:         report.Symbol,
		Interval:       report.Interval,
		Candles:        report.CandleCount,
		Trades:         report.Metrics.TotalTrades,
		Wins:           report.Metrics.Wins,
		Losses:         report.Metrics.Losses,
		WinRate:        report.Metrics.WinRate,
		GrossPnL:       report.Metrics.GrossPnL,
		Fees:           report.Metrics.Fees,
		NetPnL:         report.Metrics.NetPnL,
		MaxDrawdown:    report.Metrics.MaxDrawdown,
		MaxDrawdownPct: report.Metrics.MaxDrawdownPct,
	}

	applyTradeStats(&analysis, trades, signals.Times)
	applySignalStats(&analysis, signals)
	analysis.Notes = chapter2Notes(analysis)

	return analysis
}

func WriteChapter2AnalysisCSV(path string, rows []Chapter2Analysis) error {
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
		"max_drawdown",
		"max_drawdown_pct",
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
			floatToString(row.MaxDrawdown),
			floatToString(row.MaxDrawdownPct),
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

func WriteChapter2AnalysisJSON(path string, rows []Chapter2Analysis) error {
	if err := ensureDir(path); err != nil {
		return err
	}

	body, err := json.MarshalIndent(rows, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, append(body, '\n'), 0644)
}

func applyTradeStats(analysis *Chapter2Analysis, trades []strategy.Trade, times []int64) {
	if len(trades) == 0 {
		return
	}

	best := trades[0].RealizedPnL
	worst := trades[0].RealizedPnL
	holdTotal := 0
	minHold := 0
	maxHold := 0
	interval := candleInterval(times)

	for i, trade := range trades {
		if trade.RealizedPnL > best {
			best = trade.RealizedPnL
		}
		if trade.RealizedPnL < worst {
			worst = trade.RealizedPnL
		}

		hold := holdCandles(trade.OpenedAt, trade.ClosedAt, interval)
		holdTotal += hold
		if i == 0 || hold < minHold {
			minHold = hold
		}
		if hold > maxHold {
			maxHold = hold
		}
	}

	analysis.BestTradePnL = best
	analysis.WorstTradePnL = worst
	analysis.AverageTradePnL = analysis.NetPnL / float64(len(trades))
	analysis.AverageHoldCandles = float64(holdTotal) / float64(len(trades))
	analysis.MaxHoldCandles = maxHold
	analysis.MinHoldCandles = minHold
}

func applySignalStats(analysis *Chapter2Analysis, signals series.SignalResult) {
	for _, position := range signals.Positions {
		switch {
		case position > 0:
			analysis.SignalCount++
			analysis.BuySignalCount++
		case position < 0:
			analysis.SignalCount++
			analysis.SellSignalCount++
		default:
			analysis.HoldSignalCount++
		}
	}
	if analysis.Candles > 0 {
		analysis.OvertradeScore = float64(analysis.Trades) / float64(analysis.Candles)
	}
}

func chapter2Notes(analysis Chapter2Analysis) []string {
	notes := make([]string, 0, 4)
	if analysis.Trades < 5 {
		notes = append(notes, "low_sample_size")
	}
	if analysis.OvertradeScore > 0.10 {
		notes = append(notes, "overtrading")
	}
	if analysis.AverageTradePnL < 0 {
		notes = append(notes, "negative_expectancy")
	} else if analysis.AverageTradePnL > 0 {
		notes = append(notes, "positive_expectancy_sample")
	}
	return notes
}

func candleInterval(times []int64) time.Duration {
	if len(times) < 2 {
		return 0
	}
	for i := 1; i < len(times); i++ {
		diff := times[i] - times[i-1]
		if diff > 0 {
			return time.Duration(diff) * time.Millisecond
		}
	}
	return 0
}

func holdCandles(openedAt time.Time, closedAt time.Time, interval time.Duration) int {
	if interval <= 0 || closedAt.Before(openedAt) {
		return 0
	}
	count := int(math.Round(float64(closedAt.Sub(openedAt)) / float64(interval)))
	if count < 1 {
		return 1
	}
	return count
}
