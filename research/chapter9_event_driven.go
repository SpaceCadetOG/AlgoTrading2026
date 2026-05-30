package research

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"AlgoTrading2026/backtest"
)

type Chapter9EventDrivenReport struct {
	Title         string                             `json:"title"`
	Result        backtest.EventDrivenBacktestResult `json:"result"`
	Implemented   []string                           `json:"implemented"`
	QueueFlow     []string                           `json:"queueFlow"`
	Safety        []string                           `json:"safety"`
	RemainingGaps []string                           `json:"remainingGaps"`
	Conclusion    string                             `json:"conclusion"`
	NextPhase     string                             `json:"nextPhase"`
}

func BuildChapter9EventDrivenReport(result backtest.EventDrivenBacktestResult) Chapter9EventDrivenReport {
	return Chapter9EventDrivenReport{
		Title:  "Chapter 9D Event-Driven Backtester",
		Result: result,
		Implemented: []string{
			"Historical candles replay through backtest.CandleTimeIterator.",
			"SimulatedClock advances to each candle timestamp.",
			"Each candle becomes deterministic crossed-book liquidity for the Chapter 7 test strategy.",
			"Chapter 7 LiquidityProvider, OrderBook, TradingStrategy, OrderManager, and MarketSimulator are reused.",
			"Simulated fills return through the OMS to the strategy.",
		},
		QueueFlow: []string{
			"lp_2_gateway -> OrderBook",
			"ob_2_ts -> TradingStrategy",
			"ts_2_om -> OrderManager",
			"om_2_gw -> MarketSimulator or SimulatedGateway",
			"gw_2_om -> OrderManager",
			"om_2_ts -> TradingStrategy",
			"ms_2_om -> OrderManager audit stream",
		},
		Safety: []string{
			"AllowLiveOrders is forced to false.",
			"No venue adapters are called.",
			"No paper trading is enabled.",
			"No live trading is enabled.",
			"No WebSocket subscriptions or real API calls are made.",
		},
		RemainingGaps: []string{
			"Event-driven strategy is a deterministic crossed-book test strategy only.",
			"Latency and response scheduling are not modeled.",
			"OMS timeout behavior is not integrated with the simulated clock yet.",
			"Partial fills and fill ratios are not modeled in the event-driven simulator.",
			"Dual moving average event-driven strategy comparison remains pending.",
		},
		Conclusion: "Chapter 9D adds an event-driven backtester skeleton using Chapter 7 system queues and Chapter 9 simulated time without enabling paper or live trading.",
		NextPhase:  "Chapter 9E should add the book's dual moving average event-driven comparison while keeping execution simulated.",
	}
}

func WriteChapter9EventDrivenJSON(path string, report Chapter9EventDrivenReport) error {
	if err := ensureDir(path); err != nil {
		return err
	}
	body, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, append(body, '\n'), 0644)
}

func WriteChapter9EventDrivenMarkdown(path string, report Chapter9EventDrivenReport) error {
	if err := ensureDir(path); err != nil {
		return err
	}
	return os.WriteFile(path, []byte(Chapter9EventDrivenMarkdown(report)), 0644)
}

func Chapter9EventDrivenMarkdown(report Chapter9EventDrivenReport) string {
	var b strings.Builder
	b.WriteString("# " + report.Title + "\n\n")
	b.WriteString("## Result\n\n")
	fmt.Fprintf(&b, "- candles_processed: %d\n", report.Result.CandlesProcessed)
	fmt.Fprintf(&b, "- orders_created: %d\n", report.Result.OrdersCreated)
	fmt.Fprintf(&b, "- orders_filled: %d\n", report.Result.OrdersFilled)
	fmt.Fprintf(&b, "- final_cash: %.2f\n", report.Result.FinalCash)
	fmt.Fprintf(&b, "- final_position: %.4f\n", report.Result.FinalPosition)
	fmt.Fprintf(&b, "- final_pnl: %.2f\n", report.Result.FinalPnL)
	fmt.Fprintf(&b, "- clock_start: %s\n", report.Result.ClockStart)
	fmt.Fprintf(&b, "- clock_end: %s\n", report.Result.ClockEnd)
	fmt.Fprintf(&b, "- audit_events: %d\n", report.Result.AuditEvents)

	writeStringList(&b, "Implemented", report.Implemented)
	writeStringList(&b, "Queue Flow", report.QueueFlow)
	writeStringList(&b, "Safety", report.Safety)
	writeStringList(&b, "Remaining Gaps", report.RemainingGaps)

	b.WriteString("\n## Conclusion\n\n")
	b.WriteString(report.Conclusion + "\n")
	b.WriteString("\n## Next Phase\n\n")
	b.WriteString(report.NextPhase + "\n")
	return b.String()
}
