package research

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"AlgoTrading2026/realism"
)

type Chapter10RealismModelsReport struct {
	Title      string                        `json:"title"`
	Model      realism.AggregateRealismModel `json:"model"`
	Strengths  []string                      `json:"strengths"`
	Gaps       []string                      `json:"gaps"`
	Conclusion []string                      `json:"conclusion"`
	NextPhase  string                        `json:"nextPhase"`
}

func BuildChapter10RealismModelsReport(model realism.AggregateRealismModel) Chapter10RealismModelsReport {
	return Chapter10RealismModelsReport{
		Title: "Chapter 10B Simulation Realism Models",
		Model: model,
		Strengths: []string{
			"latency dimensions are now modeled explicitly",
			"place-in-line estimation is represented as queue position and fill probability",
			"market impact is represented by participation rate and estimated impact bps",
			"fill assumptions explicitly flag deterministic fills, stale book risk, and crossed-book artificiality",
			"aggregate realism risk is deterministic and testable",
		},
		Gaps: []string{
			"models are not yet applied to production strategy behavior",
			"models are not yet applied to event-driven fill decisions",
			"latency model is an assumption model, not calibrated from observed exchange data",
			"place-in-line model is an estimate, not venue-specific queue reconstruction",
			"market impact model is a simple participation-rate estimate",
			"historical market data quality and historical/live parity audit remain pending",
		},
		Conclusion: []string{
			"Chapter 10B models realism assumptions only.",
			"No live or paper trading is enabled.",
			"These models are not yet used to change strategy behavior.",
			"Next phase is Chapter 10C market data quality and historical/live parity audit.",
		},
		NextPhase: "Chapter 10C market data quality and historical/live parity audit.",
	}
}

func WriteChapter10RealismModelsJSON(path string, report Chapter10RealismModelsReport) error {
	if err := ensureDir(path); err != nil {
		return err
	}
	body, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, append(body, '\n'), 0644)
}

func WriteChapter10RealismModelsMarkdown(path string, report Chapter10RealismModelsReport) error {
	if err := ensureDir(path); err != nil {
		return err
	}
	return os.WriteFile(path, []byte(Chapter10RealismModelsMarkdown(report)), 0644)
}

func Chapter10RealismModelsMarkdown(report Chapter10RealismModelsReport) string {
	var b strings.Builder
	b.WriteString("# " + report.Title + "\n\n")

	b.WriteString("## Aggregate Risk\n\n")
	fmt.Fprintf(&b, "- overall_risk: %s\n", report.Model.OverallRisk)
	fmt.Fprintf(&b, "- warnings: %d\n", len(report.Model.Warnings))

	b.WriteString("\n## Latency Model\n\n")
	fmt.Fprintf(&b, "- signal_latency: %s\n", report.Model.Latency.SignalLatency)
	fmt.Fprintf(&b, "- strategy_latency: %s\n", report.Model.Latency.StrategyLatency)
	fmt.Fprintf(&b, "- gateway_latency: %s\n", report.Model.Latency.GatewayLatency)
	fmt.Fprintf(&b, "- exchange_response_latency: %s\n", report.Model.Latency.ExchangeResponseLatency)
	fmt.Fprintf(&b, "- latency_variance: %s\n", report.Model.Latency.LatencyVariance)
	fmt.Fprintf(&b, "- total_expected_latency: %s\n", report.Model.Latency.TotalExpectedLatency)
	fmt.Fprintf(&b, "- latency_risk: %s\n", report.Model.Latency.Risk)

	b.WriteString("\n## Place-In-Line Estimate\n\n")
	fmt.Fprintf(&b, "- order_size: %.4f\n", report.Model.PlaceInLine.OrderSize)
	fmt.Fprintf(&b, "- visible_liquidity: %.4f\n", report.Model.PlaceInLine.VisibleLiquidity)
	fmt.Fprintf(&b, "- estimated_queue_position: %.4f\n", report.Model.PlaceInLine.EstimatedQueuePosition)
	fmt.Fprintf(&b, "- fill_probability: %.4f\n", report.Model.PlaceInLine.FillProbability)
	fmt.Fprintf(&b, "- place_in_line_risk: %s\n", report.Model.PlaceInLine.Risk)

	b.WriteString("\n## Market Impact Model\n\n")
	fmt.Fprintf(&b, "- order_size: %.4f\n", report.Model.MarketImpact.OrderSize)
	fmt.Fprintf(&b, "- average_liquidity: %.4f\n", report.Model.MarketImpact.AverageLiquidity)
	fmt.Fprintf(&b, "- participation_rate: %.4f\n", report.Model.MarketImpact.ParticipationRate)
	fmt.Fprintf(&b, "- estimated_impact_bps: %.4f\n", report.Model.MarketImpact.EstimatedImpactBps)
	fmt.Fprintf(&b, "- impact_risk: %s\n", report.Model.MarketImpact.Risk)

	b.WriteString("\n## Fill Assumptions\n\n")
	fmt.Fprintf(&b, "- fill_ratio: %.4f\n", report.Model.FillAssumption.FillRatio)
	fmt.Fprintf(&b, "- partial_fill_supported: %t\n", report.Model.FillAssumption.PartialFillSupported)
	fmt.Fprintf(&b, "- cancel_amend_supported: %t\n", report.Model.FillAssumption.CancelAmendSupported)
	fmt.Fprintf(&b, "- stale_book_risk: %t\n", report.Model.FillAssumption.StaleBookRisk)
	fmt.Fprintf(&b, "- crossed_book_artificiality: %t\n", report.Model.FillAssumption.CrossedBookArtificiality)
	fmt.Fprintf(&b, "- deterministic_fill_warning: %t\n", report.Model.FillAssumption.DeterministicFillWarning)
	fmt.Fprintf(&b, "- fill_risk: %s\n", report.Model.FillAssumption.Risk)

	writeStringList(&b, "Warnings", report.Model.Warnings)
	writeStringList(&b, "Strengths", report.Strengths)
	writeStringList(&b, "Gaps", report.Gaps)
	writeStringList(&b, "Conclusion", report.Conclusion)

	b.WriteString("\n## Next Phase\n\n")
	b.WriteString(report.NextPhase + "\n")
	return b.String()
}
