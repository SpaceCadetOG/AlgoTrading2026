package research

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"AlgoTrading2026/l2recorder"
)

func WriteL2SnapshotAnalysisJSON(path string, analysis l2recorder.SnapshotAnalysis) error {
	if err := ensureDir(path); err != nil {
		return err
	}
	body, err := json.MarshalIndent(analysis, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, append(body, '\n'), 0644)
}

func WriteL2SnapshotAnalysisMarkdown(path string, analysis l2recorder.SnapshotAnalysis) error {
	if err := ensureDir(path); err != nil {
		return err
	}
	return os.WriteFile(path, []byte(L2SnapshotAnalysisMarkdown(analysis)), 0644)
}

func L2SnapshotAnalysisMarkdown(analysis l2recorder.SnapshotAnalysis) string {
	var b strings.Builder
	b.WriteString("# Historical L2 Snapshot Analysis\n\n")

	quality := analysis.DatasetQuality
	b.WriteString("## Dataset Quality\n\n")
	fmt.Fprintf(&b, "- totalRows: %d\n", quality.TotalRows)
	fmt.Fprintf(&b, "- validRows: %d\n", quality.ValidRows)
	fmt.Fprintf(&b, "- invalidRows: %d\n", quality.InvalidRows)
	fmt.Fprintf(&b, "- errorRows: %d\n", quality.ErrorRows)
	fmt.Fprintf(&b, "- validPct: %.2f\n", quality.ValidPct)
	fmt.Fprintf(&b, "- errorPct: %.2f\n", quality.ErrorPct)

	b.WriteString("\n## Venue Summaries\n\n")
	b.WriteString("| Venue | Snapshots | Valid | Invalid | Avg Spread % | Min Spread % | Max Spread % | Avg Imbalance | Min Imbalance | Max Imbalance | Avg Bid Depth | Avg Ask Depth | Avg Mid |\n")
	b.WriteString("|---|---:|---:|---:|---:|---:|---:|---:|---:|---:|---:|---:|---:|\n")
	for _, row := range analysis.VenueSummaries {
		fmt.Fprintf(
			&b,
			"| %s | %d | %d | %d | %.8f | %.8f | %.8f | %.8f | %.8f | %.8f | %.2f | %.2f | %.2f |\n",
			row.Venue,
			row.SnapshotCount,
			row.ValidSnapshotCount,
			row.InvalidSnapshotCount,
			row.AverageSpreadPct,
			row.MinSpreadPct,
			row.MaxSpreadPct,
			row.AverageImbalance1Pct,
			row.MinImbalance1Pct,
			row.MaxImbalance1Pct,
			row.AverageBidDepth1Pct,
			row.AverageAskDepth1Pct,
			row.AverageMid,
		)
	}

	cross := analysis.CrossVenue
	b.WriteString("\n## Cross Venue Comparison\n\n")
	fmt.Fprintf(&b, "- comparableGroups: %d\n", cross.ComparableGroups)
	fmt.Fprintf(&b, "- averageMidDifference: %.8f\n", cross.AverageMidDifference)
	fmt.Fprintf(&b, "- maxMidDifference: %.8f\n", cross.MaxMidDifference)
	fmt.Fprintf(&b, "- averageSpreadDifference: %.8f\n", cross.AverageSpreadDifference)
	fmt.Fprintf(&b, "- widestSpreadVenue: %s\n", cross.WidestSpreadVenue)
	fmt.Fprintf(&b, "- tightestSpreadVenue: %s\n", cross.TightestSpreadVenue)
	fmt.Fprintf(&b, "- deepestLiquidityVenue: %s\n", cross.DeepestLiquidityVenue)

	b.WriteString("\n## Liquidity Rankings\n\n")
	writeRankings(&b, "Average Bid Depth", analysis.LiquidityRanking.AverageBidDepth)
	writeRankings(&b, "Average Ask Depth", analysis.LiquidityRanking.AverageAskDepth)
	writeRankings(&b, "Combined Depth", analysis.LiquidityRanking.CombinedDepth)
	writeRankings(&b, "Tightest Spreads", analysis.LiquidityRanking.TightestSpreads)
	writeRankings(&b, "Highest Bid Pressure", analysis.LiquidityRanking.HighestBidPressure)
	writeRankings(&b, "Highest Ask Pressure", analysis.LiquidityRanking.HighestAskPressure)

	b.WriteString("\n## Findings\n\n")
	for _, finding := range L2SnapshotFindings(analysis) {
		b.WriteString("- " + finding + "\n")
	}

	writeStringList(&b, "Notes", analysis.Notes)
	return b.String()
}

func L2SnapshotFindings(analysis l2recorder.SnapshotAnalysis) []string {
	findings := make([]string, 0)
	cross := analysis.CrossVenue
	if cross.TightestSpreadVenue != "" {
		findings = append(findings, fmt.Sprintf("Tightest spread venue: %s.", cross.TightestSpreadVenue))
	}
	if cross.WidestSpreadVenue != "" {
		findings = append(findings, fmt.Sprintf("Widest spread venue: %s.", cross.WidestSpreadVenue))
	}
	if cross.DeepestLiquidityVenue != "" {
		findings = append(findings, fmt.Sprintf("Deepest liquidity venue: %s.", cross.DeepestLiquidityVenue))
	}
	if len(analysis.LiquidityRanking.HighestBidPressure) > 0 {
		top := analysis.LiquidityRanking.HighestBidPressure[0]
		findings = append(findings, fmt.Sprintf("Highest bid pressure venue: %s (avg imbalance %.4f).", top.Venue, top.Value))
	}
	if len(analysis.LiquidityRanking.HighestAskPressure) > 0 {
		top := analysis.LiquidityRanking.HighestAskPressure[0]
		findings = append(findings, fmt.Sprintf("Highest ask pressure venue: %s (avg imbalance %.4f).", top.Venue, top.Value))
	}
	if analysis.DatasetQuality.ErrorRows > 0 {
		findings = append(findings, fmt.Sprintf("Data quality concern: %d rows contain fetch or validation errors.", analysis.DatasetQuality.ErrorRows))
	}
	if analysis.DatasetQuality.TotalRows > 0 && analysis.DatasetQuality.ValidPct >= 95 {
		findings = append(findings, "Recorder output looks stable enough for a longer pilot collection.")
	}
	return findings
}

func writeRankings(b *strings.Builder, title string, rows []l2recorder.VenueRanking) {
	b.WriteString("### " + title + "\n\n")
	if len(rows) == 0 {
		b.WriteString("No valid venues.\n\n")
		return
	}
	b.WriteString("| Rank | Venue | Value |\n")
	b.WriteString("|---:|---|---:|\n")
	for i, row := range rows {
		fmt.Fprintf(b, "| %d | %s | %.8f |\n", i+1, row.Venue, row.Value)
	}
	b.WriteString("\n")
}
