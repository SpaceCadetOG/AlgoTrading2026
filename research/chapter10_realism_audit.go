package research

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"AlgoTrading2026/backtest"
	"AlgoTrading2026/realism"
)

type Chapter10RealismAudit struct {
	Title                     string                        `json:"title"`
	Dislocation               realism.SimulationDislocation `json:"dislocation"`
	Assumptions               realism.RealismAssumptions    `json:"assumptions"`
	Grade                     realism.RealismGrade          `json:"grade"`
	RealismStrengths          []string                      `json:"realismStrengths"`
	RealismGaps               []string                      `json:"realismGaps"`
	ValidationFindings        []string                      `json:"validationFindings"`
	EventDrivenClassification string                        `json:"eventDrivenClassification"`
	BeforePaperLive           []string                      `json:"beforePaperLive"`
	Conclusion                []string                      `json:"conclusion"`
	NextPhase                 string                        `json:"nextPhase"`
}

func BuildChapter10RealismAudit(comparison backtest.BacktesterComparisonResult) Chapter10RealismAudit {
	forLoopPnL := comparison.EventDrivenFinalPnL - comparison.PnLDifference
	assumptions := realism.DefaultRealismAssumptions()
	dislocation := realism.NewSimulationDislocation(
		"chapter9_backtester_comparison",
		comparison.Symbol,
		comparison.Candles,
		forLoopPnL,
		comparison.EventDrivenFinalPnL,
	)

	return Chapter10RealismAudit{
		Title:       "Chapter 10A Backtester Realism and Simulation Dislocation Audit",
		Dislocation: dislocation,
		Assumptions: assumptions,
		Grade:       realism.GradeRealism(assumptions, dislocation.Severity),
		RealismStrengths: append(assumptions.Strengths(),
			"for-loop and event-driven backtesters are separated",
			"Chapter 9 validation identifies crossed-book artificial profit",
			"simulated clock exists for deterministic event replay",
		),
		RealismGaps: assumptions.Missing(),
		ValidationFindings: []string{
			"event-driven PnL is dominated by deterministic crossed-book spread capture",
			"event-driven fills are full fills at submitted prices",
			"event-driven fills do not currently include fees, slippage, latency, queue position, partial fill ratio, or market impact",
			"OrderBook accumulates liquidity and does not consume or expire filled resting liquidity",
			"OMS does not yet enforce idempotency against duplicate fill responses",
		},
		EventDrivenClassification: "system-validation only; not strategy-performance PnL",
		BeforePaperLive: []string{
			"replace artificial crossed-book generation for performance tests",
			"consume, expire, or reset order-book liquidity",
			"add latency and latency variance modeling",
			"add place-in-line and partial-fill modeling",
			"add market impact modeling",
			"calibrate fees and slippage assumptions",
			"add historical market data accuracy checks",
			"add OMS response idempotency",
		},
		Conclusion: []string{
			"Chapter 10A measures simulation dislocation risk only.",
			"No live or paper trading is enabled.",
			"Current for-loop backtests are useful for strategy research.",
			"Current event-driven backtester is useful for OMS/system validation, not performance estimates.",
			"Next phase should model latency, market impact, and place-in-line before any forward-testing shell.",
		},
		NextPhase: "Model latency, market impact, and place-in-line before any forward-testing shell.",
	}
}

func WriteChapter10RealismAuditJSON(path string, audit Chapter10RealismAudit) error {
	if err := ensureDir(path); err != nil {
		return err
	}
	body, err := json.MarshalIndent(audit, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, append(body, '\n'), 0644)
}

func WriteChapter10RealismAuditMarkdown(path string, audit Chapter10RealismAudit) error {
	if err := ensureDir(path); err != nil {
		return err
	}
	return os.WriteFile(path, []byte(Chapter10RealismAuditMarkdown(audit)), 0644)
}

func Chapter10RealismAuditMarkdown(audit Chapter10RealismAudit) string {
	var b strings.Builder
	b.WriteString("# " + audit.Title + "\n\n")
	b.WriteString("## Simulation Dislocation\n\n")
	fmt.Fprintf(&b, "- strategy: %s\n", audit.Dislocation.Strategy)
	fmt.Fprintf(&b, "- symbol: %s\n", audit.Dislocation.Symbol)
	fmt.Fprintf(&b, "- candles: %d\n", audit.Dislocation.Candles)
	fmt.Fprintf(&b, "- for_loop_pnl: %.2f\n", audit.Dislocation.ForLoopPnL)
	fmt.Fprintf(&b, "- event_driven_pnl: %.2f\n", audit.Dislocation.EventDrivenPnL)
	fmt.Fprintf(&b, "- pnl_difference: %.2f\n", audit.Dislocation.PnLDifference)
	fmt.Fprintf(&b, "- pnl_difference_pct: %.2f\n", audit.Dislocation.PnLDifferencePct)
	fmt.Fprintf(&b, "- bias_direction: %s\n", audit.Dislocation.BiasDirection)
	fmt.Fprintf(&b, "- severity: %s\n", audit.Dislocation.Severity)
	fmt.Fprintf(&b, "- realism_grade: %s\n", audit.Grade)

	b.WriteString("\n## Event-Driven Classification\n\n")
	b.WriteString(audit.EventDrivenClassification + "\n")

	writeStringList(&b, "Realism Strengths", audit.RealismStrengths)
	writeStringList(&b, "Realism Gaps", audit.RealismGaps)
	writeStringList(&b, "Validation Findings", audit.ValidationFindings)
	writeStringList(&b, "Must Be Solved Before Paper/Live Trading", audit.BeforePaperLive)
	writeStringList(&b, "Conclusion", audit.Conclusion)

	b.WriteString("\n## Next Phase\n\n")
	b.WriteString(audit.NextPhase + "\n")
	return b.String()
}
