package research

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

type TimeAssumptionReport struct {
	Title                     string   `json:"title"`
	Implemented               []string `json:"implemented"`
	CurrentCandleUsage        string   `json:"currentCandleUsage"`
	WhyNoWallClockBacktests   string   `json:"whyNoWallClockBacktests"`
	RemainingGaps             []string `json:"remainingGaps"`
	ReadinessForEventBacktest string   `json:"readinessForEventBacktest"`
}

func BuildChapter9TimeModelReport(candles int) TimeAssumptionReport {
	return TimeAssumptionReport{
		Title: "Chapter 9C Simulated Clock and Value-of-Time Layer",
		Implemented: []string{
			"backtest.SimulatedClock exposes Current, Advance, Set, Reset, and StepTo.",
			"SimulatedClock rejects backward movement unless Reset is used explicitly.",
			"backtest.CandleTimeIterator sorts candles by timestamp and advances the simulated clock candle-by-candle.",
			fmt.Sprintf("The main harness validates the time model against %d candles.", candles),
		},
		CurrentCandleUsage:      "The existing for-loop backtester uses candle StartTime and EndUTC timestamps for replay, fills, equity points, and trade timestamps.",
		WhyNoWallClockBacktests: "Backtests need deterministic simulated time so repeated runs do not depend on machine clock, runtime duration, network timing, or wall-clock scheduling.",
		RemainingGaps: []string{
			"OMS timeout modeling is not implemented.",
			"Latency simulation is not implemented.",
			"Event scheduling is not implemented.",
			"Exchange response delays are not modeled.",
		},
		ReadinessForEventBacktest: "Ready to use as the deterministic time source for a Chapter 9 event-driven backtester skeleton.",
	}
}

func WriteChapter9TimeModelJSON(path string, report TimeAssumptionReport) error {
	if err := ensureDir(path); err != nil {
		return err
	}
	body, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, append(body, '\n'), 0644)
}

func WriteChapter9TimeModelMarkdown(path string, report TimeAssumptionReport) error {
	if err := ensureDir(path); err != nil {
		return err
	}
	return os.WriteFile(path, []byte(Chapter9TimeModelMarkdown(report)), 0644)
}

func Chapter9TimeModelMarkdown(report TimeAssumptionReport) string {
	var b strings.Builder
	b.WriteString("# " + report.Title + "\n\n")
	writeStringList(&b, "Implemented", report.Implemented)
	b.WriteString("\n## Current Candle Timestamp Usage\n\n")
	b.WriteString(report.CurrentCandleUsage + "\n")
	b.WriteString("\n## Why Wall-Clock Time Is Not Used\n\n")
	b.WriteString(report.WhyNoWallClockBacktests + "\n")
	writeStringList(&b, "Remaining Gaps", report.RemainingGaps)
	b.WriteString("\n## Readiness For Event Backtest\n\n")
	b.WriteString(report.ReadinessForEventBacktest + "\n")
	return b.String()
}
