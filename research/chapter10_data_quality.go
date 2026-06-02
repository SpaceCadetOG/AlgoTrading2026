package research

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"AlgoTrading2026/realism"
)

type Chapter10DataQualityReport struct {
	Title       string                    `json:"title"`
	Venue       string                    `json:"venue"`
	Symbol      string                    `json:"symbol"`
	Interval    string                    `json:"interval"`
	DataQuality realism.DataQualityReport `json:"dataQuality"`
	Strengths   []string                  `json:"strengths"`
	Gaps        []string                  `json:"gaps"`
	Conclusion  []string                  `json:"conclusion"`
	NextPhase   string                    `json:"nextPhase"`
}

func BuildChapter10DataQualityReport(venue string, symbol string, interval string, dataQuality realism.DataQualityReport) Chapter10DataQualityReport {
	return Chapter10DataQualityReport{
		Title:       "Chapter 10C Market Data Quality and Historical/Live Parity Audit",
		Venue:       venue,
		Symbol:      symbol,
		Interval:    interval,
		DataQuality: dataQuality,
		Strengths: []string{
			"normalized candle data is checked before being used for realism-sensitive research",
			"timestamp gaps, duplicates, ordering, invalid prices, invalid volume, and price jumps are measured",
			"historical/live candle schema parity is represented explicitly",
			"quality and parity risks are separate so data issues are not hidden inside strategy metrics",
		},
		Gaps: []string{
			"historical/live parity is documented from adapter behavior, not verified with live recorded samples",
			"market data accuracy is not yet cross-checked against an independent reference feed",
			"venue-specific candle repair rules are not implemented",
			"data quality gates are not yet applied to strategy promotion decisions",
		},
		Conclusion: []string{
			"Chapter 10C audits market data quality and historical/live parity only.",
			"No live or paper trading is enabled.",
			"No real exchange calls are made.",
			"Next phase is continued Chapter 10 market realism research.",
		},
		NextPhase: "Continue Chapter 10 market realism research.",
	}
}

func WriteChapter10DataQualityJSON(path string, report Chapter10DataQualityReport) error {
	if err := ensureDir(path); err != nil {
		return err
	}
	body, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, append(body, '\n'), 0644)
}

func WriteChapter10DataQualityMarkdown(path string, report Chapter10DataQualityReport) error {
	if err := ensureDir(path); err != nil {
		return err
	}
	return os.WriteFile(path, []byte(Chapter10DataQualityMarkdown(report)), 0644)
}

func Chapter10DataQualityMarkdown(report Chapter10DataQualityReport) string {
	var b strings.Builder
	b.WriteString("# " + report.Title + "\n\n")
	fmt.Fprintf(&b, "- venue: %s\n", report.Venue)
	fmt.Fprintf(&b, "- symbol: %s\n", report.Symbol)
	fmt.Fprintf(&b, "- interval: %s\n", report.Interval)

	quality := report.DataQuality.Quality
	b.WriteString("\n## Market Data Quality\n\n")
	fmt.Fprintf(&b, "- total_candles_checked: %d\n", report.DataQuality.TotalCandlesChecked)
	fmt.Fprintf(&b, "- quality_risk: %s\n", report.DataQuality.QualityRisk)
	fmt.Fprintf(&b, "- missing_candles: %d\n", quality.MissingCandles)
	fmt.Fprintf(&b, "- duplicate_timestamps: %d\n", quality.DuplicateTimestamps)
	fmt.Fprintf(&b, "- out_of_order_timestamps: %d\n", quality.OutOfOrderTimestamps)
	fmt.Fprintf(&b, "- invalid_ohlc_values: %d\n", quality.ZeroNegativeOHLC)
	fmt.Fprintf(&b, "- invalid_volume: %d\n", quality.InvalidVolume)
	fmt.Fprintf(&b, "- large_time_gaps: %d\n", quality.LargeTimeGaps)
	fmt.Fprintf(&b, "- suspicious_price_jumps: %d\n", quality.SuspiciousPriceJumps)
	fmt.Fprintf(&b, "- expected_interval_ms: %d\n", quality.ExpectedIntervalMS)
	fmt.Fprintf(&b, "- max_observed_gap_ms: %d\n", quality.MaxObservedGapMS)
	fmt.Fprintf(&b, "- max_price_jump_pct: %.4f\n", quality.MaxPriceJumpPct)

	parity := report.DataQuality.Parity
	b.WriteString("\n## Historical/Live Parity\n\n")
	fmt.Fprintf(&b, "- parity_risk: %s\n", report.DataQuality.ParityRisk)
	fmt.Fprintf(&b, "- historical_schema_documented: %t\n", parity.HistoricalSchemaDocumented)
	fmt.Fprintf(&b, "- live_schema_documented: %t\n", parity.LiveSchemaDocumented)
	fmt.Fprintf(&b, "- timestamp_format_parity: %t\n", parity.TimestampFormatParity)
	fmt.Fprintf(&b, "- symbol_format_parity: %t\n", parity.SymbolFormatParity)
	fmt.Fprintf(&b, "- price_precision_parity: %t\n", parity.PricePrecisionParity)
	fmt.Fprintf(&b, "- volume_precision_parity: %t\n", parity.VolumePrecisionParity)
	fmt.Fprintf(&b, "- candle_interval_parity: %t\n", parity.CandleIntervalParity)

	writeStringList(&b, "Known Venue Differences", parity.KnownVenueDifferences)
	writeStringList(&b, "Warnings", report.DataQuality.Warnings)
	writeStringList(&b, "Recommendations", report.DataQuality.Recommendations)
	writeStringList(&b, "Strengths", report.Strengths)
	writeStringList(&b, "Gaps", report.Gaps)
	writeStringList(&b, "Conclusion", report.Conclusion)

	b.WriteString("\n## Next Phase\n\n")
	b.WriteString(report.NextPhase + "\n")
	return b.String()
}
