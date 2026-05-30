package research

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"AlgoTrading2026/backtest"
)

type Chapter9BacktesterComparisonReport struct {
	Title          string                              `json:"title"`
	Result         backtest.BacktesterComparisonResult `json:"result"`
	Implemented    []string                            `json:"implemented"`
	Interpretation []string                            `json:"interpretation"`
	Conclusion     string                              `json:"conclusion"`
	NextPhase      string                              `json:"nextPhase"`
}

func BuildChapter9BacktesterComparisonReport(result backtest.BacktesterComparisonResult) Chapter9BacktesterComparisonReport {
	return Chapter9BacktesterComparisonReport{
		Title:  "Chapter 9E For-Loop vs Event-Driven Backtester Comparison",
		Result: result,
		Implemented: []string{
			"Runs the existing for-loop candle-driven backtester on the same candle sample.",
			"Runs the Chapter 9D event-driven backtester on the same candle sample.",
			"Computes PnL difference without forcing assumptions or results to match.",
			"Documents why the two backtesters are expected to diverge.",
		},
		Interpretation: []string{
			"For-loop backtesting is best for fast signal research.",
			"Event-driven backtesting is best for validating system flow, OMS behavior, market simulator behavior, and time assumptions.",
			"PnL differences are diagnostic, not errors, because the engines model different execution assumptions.",
		},
		Conclusion: "Chapter 9E compares for-loop and event-driven backtests honestly while keeping execution simulated.",
		NextPhase:  "Chapter 9F should add the book's dual moving average event-driven comparison using simulated queues and simulated time.",
	}
}

func WriteChapter9BacktesterComparisonJSON(path string, report Chapter9BacktesterComparisonReport) error {
	if err := ensureDir(path); err != nil {
		return err
	}
	body, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, append(body, '\n'), 0644)
}

func WriteChapter9BacktesterComparisonMarkdown(path string, report Chapter9BacktesterComparisonReport) error {
	if err := ensureDir(path); err != nil {
		return err
	}
	return os.WriteFile(path, []byte(Chapter9BacktesterComparisonMarkdown(report)), 0644)
}

func Chapter9BacktesterComparisonMarkdown(report Chapter9BacktesterComparisonReport) string {
	var b strings.Builder
	b.WriteString("# " + report.Title + "\n\n")
	b.WriteString("## Result\n\n")
	fmt.Fprintf(&b, "- symbol: %s\n", report.Result.Symbol)
	fmt.Fprintf(&b, "- candles: %d\n", report.Result.Candles)
	fmt.Fprintf(&b, "- for_loop_trades: %d\n", report.Result.ForLoopTrades)
	fmt.Fprintf(&b, "- event_driven_orders: %d\n", report.Result.EventDrivenOrders)
	fmt.Fprintf(&b, "- event_driven_fills: %d\n", report.Result.EventDrivenFills)
	fmt.Fprintf(&b, "- for_loop_final_equity: %.2f\n", report.Result.ForLoopFinalEquity)
	fmt.Fprintf(&b, "- event_driven_final_pnl: %.2f\n", report.Result.EventDrivenFinalPnL)
	fmt.Fprintf(&b, "- pnl_difference: %.2f\n", report.Result.PnLDifference)
	fmt.Fprintf(&b, "- assumptions_difference: %s\n", report.Result.AssumptionsDifference)
	fmt.Fprintf(&b, "- recommendation: %s\n", report.Result.Recommendation)

	writeStringList(&b, "Implemented", report.Implemented)
	writeStringList(&b, "Interpretation", report.Interpretation)
	writeStringList(&b, "Assumption Comparison", report.Result.AssumptionComparisons)

	b.WriteString("\n## Conclusion\n\n")
	b.WriteString(report.Conclusion + "\n")
	b.WriteString("\n## Next Phase\n\n")
	b.WriteString(report.NextPhase + "\n")
	return b.String()
}
