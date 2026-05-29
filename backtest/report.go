package backtest

import "fmt"

type Report struct {
	Venue           string
	Symbol          string
	Interval        string
	CandleCount     int
	StartingBalance float64
	EndingBalance   float64
	Metrics         Metrics
	RejectReasons   []string
	CSVPath         string
	EquityCSVPath   string
	SignalCSVPath   string
}

func (r Report) ConsoleString() string {
	return fmt.Sprintf(`=== BACKTEST REPORT ===
venue=%s
symbol=%s
interval=%s
candles=%d
trades=%d
wins=%d
losses=%d
winRate=%.2f%%
grossPnL=%.2f
fees=%.2f
netPnL=%.2f
maxDrawdown=%.2f
maxDrawdownPct=%.2f%%
averagePnL=%.2f
endingBalance=%.2f
csv=%s
rejects=%d

=== EQUITY SUMMARY ===
startBalance=%.2f
endBalance=%.2f
peakEquity=%.2f
maxDrawdown=%.2f
maxDrawdownPct=%.2f%%

tradeCSV=%s
equityCSV=%s
signalCSV=%s`,
		r.Venue,
		r.Symbol,
		r.Interval,
		r.CandleCount,
		r.Metrics.TotalTrades,
		r.Metrics.Wins,
		r.Metrics.Losses,
		r.Metrics.WinRate,
		r.Metrics.GrossPnL,
		r.Metrics.Fees,
		r.Metrics.NetPnL,
		r.Metrics.MaxDrawdown,
		r.Metrics.MaxDrawdownPct,
		r.Metrics.AveragePnL,
		r.EndingBalance,
		r.CSVPath,
		len(r.RejectReasons),
		r.StartingBalance,
		r.EndingBalance,
		r.Metrics.PeakEquity,
		r.Metrics.MaxDrawdown,
		r.Metrics.MaxDrawdownPct,
		r.CSVPath,
		r.EquityCSVPath,
		r.SignalCSVPath,
	)
}
