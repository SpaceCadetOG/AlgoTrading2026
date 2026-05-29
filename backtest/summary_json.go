package backtest

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type SummaryJSON struct {
	Venue          string  `json:"venue"`
	Symbol         string  `json:"symbol"`
	Interval       string  `json:"interval"`
	Candles        int     `json:"candles"`
	Trades         int     `json:"trades"`
	Wins           int     `json:"wins"`
	Losses         int     `json:"losses"`
	WinRate        float64 `json:"winRate"`
	GrossPnL       float64 `json:"grossPnL"`
	Fees           float64 `json:"fees"`
	NetPnL         float64 `json:"netPnL"`
	StartBalance   float64 `json:"startBalance"`
	EndingBalance  float64 `json:"endingBalance"`
	PeakEquity     float64 `json:"peakEquity"`
	MaxDrawdown    float64 `json:"maxDrawdown"`
	MaxDrawdownPct float64 `json:"maxDrawdownPct"`
	TradeCSV       string  `json:"tradeCSV"`
	EquityCSV      string  `json:"equityCSV"`
	SignalCSV      string  `json:"signalCSV"`
}

func NewSummaryJSON(report Report) SummaryJSON {
	return SummaryJSON{
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
		StartBalance:   report.StartingBalance,
		EndingBalance:  report.EndingBalance,
		PeakEquity:     report.Metrics.PeakEquity,
		MaxDrawdown:    report.Metrics.MaxDrawdown,
		MaxDrawdownPct: report.Metrics.MaxDrawdownPct,
		TradeCSV:       report.CSVPath,
		EquityCSV:      report.EquityCSVPath,
		SignalCSV:      report.SignalCSVPath,
	}
}

func WriteSummaryJSON(path string, report Report) error {
	if dir := filepath.Dir(path); dir != "." && dir != "" {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return err
		}
	}

	body, err := json.MarshalIndent(NewSummaryJSON(report), "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, append(body, '\n'), 0644)
}
