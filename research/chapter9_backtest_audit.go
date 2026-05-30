package research

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"AlgoTrading2026/backtest"
)

type Chapter9BacktestAudit struct {
	Title               string                       `json:"title"`
	InSampleCount       int                          `json:"inSampleCount"`
	OutOfSampleCount    int                          `json:"outOfSampleCount"`
	InSampleRatio       float64                      `json:"inSampleRatio"`
	Assumptions         backtest.BacktestAssumptions `json:"assumptions"`
	ImplementedConcepts []Chapter9ConceptMapping     `json:"implementedConcepts"`
	MissingConcepts     []string                     `json:"missingConcepts"`
	Conclusion          string                       `json:"conclusion"`
	NextPhase           string                       `json:"nextPhase"`
}

type Chapter9ConceptMapping struct {
	BookConcept string `json:"bookConcept"`
	RepoMapping string `json:"repoMapping"`
	Status      string `json:"status"`
}

func BuildChapter9BacktestAudit(inSampleCount int, outOfSampleCount int, assumptions backtest.BacktestAssumptions) Chapter9BacktestAudit {
	return Chapter9BacktestAudit{
		Title:            "Chapter 9B Backtest Data Splits and Assumptions",
		InSampleCount:    inSampleCount,
		OutOfSampleCount: outOfSampleCount,
		InSampleRatio:    backtest.DefaultInSampleRatio,
		Assumptions:      assumptions,
		ImplementedConcepts: []Chapter9ConceptMapping{
			{BookConcept: "in-sample vs out-of-sample", RepoMapping: "backtest.InSampleOutOfSampleSplit", Status: "implemented"},
			{BookConcept: "correct assumptions", RepoMapping: "backtest.BacktestAssumptions", Status: "documented"},
			{BookConcept: "for-loop backtester", RepoMapping: "backtest.Engine and backtest.Replay", Status: "implemented"},
			{BookConcept: "portfolio cash/holdings/total", RepoMapping: "backtest.PortfolioPoint and EquityCurve", Status: "implemented"},
			{BookConcept: "dual moving average backtest", RepoMapping: "strategies/chapter4.DualMAStrategy", Status: "strategy implemented; Chapter 9 comparison pending"},
		},
		MissingConcepts: []string{
			"paper/forward testing is not enabled",
			"event-driven backtester is not yet complete",
			"simulated clock is not implemented",
			"latency/fill-ratio/partial-fill market simulator assumptions are not fully modeled",
			"HDF5/database/time-series historical data store is not implemented",
		},
		Conclusion: "Chapter 9B formalizes historical data splits and documents current backtest assumptions before event-driven backtesting.",
		NextPhase:  "Chapter 9C should add a simulated clock and event-driven backtester skeleton using Chapter 7 queues without enabling paper or live trading.",
	}
}

func WriteChapter9BacktestAuditJSON(path string, audit Chapter9BacktestAudit) error {
	if err := ensureDir(path); err != nil {
		return err
	}
	body, err := json.MarshalIndent(audit, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, append(body, '\n'), 0644)
}

func WriteChapter9BacktestAuditMarkdown(path string, audit Chapter9BacktestAudit) error {
	if err := ensureDir(path); err != nil {
		return err
	}
	return os.WriteFile(path, []byte(Chapter9BacktestAuditMarkdown(audit)), 0644)
}

func Chapter9BacktestAuditMarkdown(audit Chapter9BacktestAudit) string {
	var b strings.Builder
	b.WriteString("# " + audit.Title + "\n\n")
	fmt.Fprintf(&b, "- in_sample_count: %d\n", audit.InSampleCount)
	fmt.Fprintf(&b, "- out_of_sample_count: %d\n", audit.OutOfSampleCount)
	fmt.Fprintf(&b, "- in_sample_ratio: %.2f\n", audit.InSampleRatio)

	b.WriteString("\n## Implemented Concepts\n\n")
	b.WriteString("| Book Concept | Repo Mapping | Status |\n")
	b.WriteString("|---|---|---|\n")
	for _, row := range audit.ImplementedConcepts {
		fmt.Fprintf(&b, "| %s | %s | %s |\n", row.BookConcept, row.RepoMapping, row.Status)
	}

	b.WriteString("\n## Current Assumptions\n\n")
	writeAssumptionRow(&b, "engine_model", audit.Assumptions.EngineModel)
	writeAssumptionRow(&b, "fill_model", audit.Assumptions.FillModel)
	writeAssumptionRow(&b, "fee_model", audit.Assumptions.FeeModel)
	writeAssumptionRow(&b, "slippage_model", audit.Assumptions.SlippageModel)
	writeAssumptionRow(&b, "fill_ratio", audit.Assumptions.FillRatio)
	writeAssumptionRow(&b, "latency_assumption", audit.Assumptions.LatencyAssumption)
	writeAssumptionRow(&b, "partial_fill_support", audit.Assumptions.PartialFillSupport)
	writeAssumptionRow(&b, "market_impact_support", audit.Assumptions.MarketImpactSupport)
	writeAssumptionRow(&b, "lookahead_bias_guard", audit.Assumptions.LookaheadBiasGuard)
	writeAssumptionRow(&b, "survivorship_bias_note", audit.Assumptions.SurvivorshipBiasNote)
	writeAssumptionRow(&b, "data_storage_note", audit.Assumptions.DataStorageNote)
	writeAssumptionRow(&b, "event_driven_support_status", audit.Assumptions.EventDrivenSupportStatus)
	writeAssumptionRow(&b, "paper_forward_testing_status", audit.Assumptions.PaperForwardTestingStatus)

	writeStringList(&b, "Assumption Gaps", audit.Assumptions.Gaps)
	writeStringList(&b, "Missing Chapter 9 Concepts", audit.MissingConcepts)

	b.WriteString("\n## Conclusion\n\n")
	b.WriteString(audit.Conclusion + "\n")
	b.WriteString("\n## Next Phase\n\n")
	b.WriteString(audit.NextPhase + "\n")
	return b.String()
}

func writeAssumptionRow(b *strings.Builder, key string, value string) {
	fmt.Fprintf(b, "- %s: %s\n", key, value)
}
