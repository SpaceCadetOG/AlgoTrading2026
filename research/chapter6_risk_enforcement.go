package research

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"

	"AlgoTrading2026/backtest"
)

type Chapter6RiskEnforcementRow struct {
	Strategy         string         `json:"strategy"`
	TradesBefore     int            `json:"tradesBefore"`
	TradesAfter      int            `json:"tradesAfter"`
	PnLBefore        float64        `json:"pnlBefore"`
	PnLAfter         float64        `json:"pnlAfter"`
	DrawdownBefore   float64        `json:"drawdownBefore"`
	DrawdownAfter    float64        `json:"drawdownAfter"`
	Violations       int            `json:"violations"`
	ViolationsByRule map[string]int `json:"violationsByRule"`
	ImprovedPnL      bool           `json:"improvedPnl"`
	ImprovedDrawdown bool           `json:"improvedDrawdown"`
	StillRejected    bool           `json:"stillRejected"`
}

type Chapter6RiskEnforcementSummary struct {
	Strategies         int                          `json:"strategies"`
	TradesBefore       int                          `json:"tradesBefore"`
	TradesAfter        int                          `json:"tradesAfter"`
	Violations         int                          `json:"violations"`
	ViolationsByRule   map[string]int               `json:"violationsByRule"`
	StrategiesImproved []string                     `json:"strategiesImproved"`
	StrategiesRejected []string                     `json:"strategiesStillRejected"`
	Rows               []Chapter6RiskEnforcementRow `json:"rows"`
	Conclusion         []string                     `json:"conclusion"`
	NextStep           string                       `json:"nextStep"`
}

func BuildChapter6RiskEnforcementSummary(results []backtest.RiskEnforcedResult, rejectedStrategies map[string]bool) Chapter6RiskEnforcementSummary {
	summary := Chapter6RiskEnforcementSummary{
		ViolationsByRule: make(map[string]int),
		Conclusion: []string{
			"Risk controls are enforced in backtest/research only.",
			"No live or paper trading is enabled.",
			"Strategy signal logic is unchanged.",
			"Next step is continued book-aligned risk and realism research.",
		},
		NextStep: "Continue book-aligned risk and realism research.",
	}

	for _, result := range results {
		row := Chapter6RiskEnforcementRow{
			Strategy:         result.Strategy,
			TradesBefore:     result.Before.Metrics.TotalTrades,
			TradesAfter:      result.After.Metrics.TotalTrades,
			PnLBefore:        result.Before.Metrics.NetPnL,
			PnLAfter:         result.After.Metrics.NetPnL,
			DrawdownBefore:   result.Before.Metrics.MaxDrawdownPct,
			DrawdownAfter:    result.After.Metrics.MaxDrawdownPct,
			Violations:       len(result.Violations),
			ViolationsByRule: result.ViolationsByRule,
			ImprovedPnL:      result.ImprovedPnL,
			ImprovedDrawdown: result.ImprovedDrawdown,
			StillRejected:    rejectedStrategies[result.Strategy],
		}
		summary.Rows = append(summary.Rows, row)
		summary.Strategies++
		summary.TradesBefore += row.TradesBefore
		summary.TradesAfter += row.TradesAfter
		summary.Violations += row.Violations
		if row.ImprovedPnL || row.ImprovedDrawdown {
			summary.StrategiesImproved = append(summary.StrategiesImproved, row.Strategy)
		}
		if row.StillRejected {
			summary.StrategiesRejected = append(summary.StrategiesRejected, row.Strategy)
		}
		for rule, count := range row.ViolationsByRule {
			summary.ViolationsByRule[rule] += count
		}
	}

	sort.Strings(summary.StrategiesImproved)
	sort.Strings(summary.StrategiesRejected)
	return summary
}

func WriteChapter6RiskEnforcementCSV(path string, rows []Chapter6RiskEnforcementRow) error {
	if err := ensureDir(path); err != nil {
		return err
	}
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	if err := writer.Write([]string{
		"strategy",
		"trades_before",
		"trades_after",
		"pnl_before",
		"pnl_after",
		"drawdown_before",
		"drawdown_after",
		"violations",
		"violations_by_rule",
		"improved_pnl",
		"improved_drawdown",
		"still_rejected",
	}); err != nil {
		return err
	}

	for _, row := range rows {
		if err := writer.Write([]string{
			row.Strategy,
			strconv.Itoa(row.TradesBefore),
			strconv.Itoa(row.TradesAfter),
			formatEnforcementFloat(row.PnLBefore),
			formatEnforcementFloat(row.PnLAfter),
			formatEnforcementFloat(row.DrawdownBefore),
			formatEnforcementFloat(row.DrawdownAfter),
			strconv.Itoa(row.Violations),
			formatRuleCounts(row.ViolationsByRule),
			strconv.FormatBool(row.ImprovedPnL),
			strconv.FormatBool(row.ImprovedDrawdown),
			strconv.FormatBool(row.StillRejected),
		}); err != nil {
			return err
		}
	}
	return nil
}

func WriteChapter6RiskEnforcementSummaryJSON(path string, summary Chapter6RiskEnforcementSummary) error {
	if err := ensureDir(path); err != nil {
		return err
	}
	body, err := json.MarshalIndent(summary, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, append(body, '\n'), 0644)
}

func WriteChapter6RiskEnforcementMarkdown(path string, summary Chapter6RiskEnforcementSummary) error {
	if err := ensureDir(path); err != nil {
		return err
	}
	return os.WriteFile(path, []byte(Chapter6RiskEnforcementMarkdown(summary)), 0644)
}

func Chapter6RiskEnforcementMarkdown(summary Chapter6RiskEnforcementSummary) string {
	var b strings.Builder
	b.WriteString("# Chapter 6B/6C Risk Controls Enforcement\n\n")
	fmt.Fprintf(&b, "- strategies: %d\n", summary.Strategies)
	fmt.Fprintf(&b, "- trades_before: %d\n", summary.TradesBefore)
	fmt.Fprintf(&b, "- trades_after: %d\n", summary.TradesAfter)
	fmt.Fprintf(&b, "- violations: %d\n", summary.Violations)

	b.WriteString("\n## Strategy Comparison\n\n")
	b.WriteString("| Strategy | Trades Before | Trades After | PnL Before | PnL After | DD Before | DD After | Violations | Still Rejected |\n")
	b.WriteString("|---|---:|---:|---:|---:|---:|---:|---:|---|\n")
	for _, row := range summary.Rows {
		fmt.Fprintf(
			&b,
			"| %s | %d | %d | %.2f | %.2f | %.2f | %.2f | %d | %t |\n",
			row.Strategy,
			row.TradesBefore,
			row.TradesAfter,
			row.PnLBefore,
			row.PnLAfter,
			row.DrawdownBefore,
			row.DrawdownAfter,
			row.Violations,
			row.StillRejected,
		)
	}

	writeStringList(&b, "Strategies Improved", summary.StrategiesImproved)
	writeStringList(&b, "Strategies Still Rejected", summary.StrategiesRejected)
	writeStringList(&b, "Conclusion", summary.Conclusion)

	b.WriteString("\n## Violations By Rule\n\n")
	if len(summary.ViolationsByRule) == 0 {
		b.WriteString("_None._\n")
	} else {
		for _, rule := range sortedRuleKeys(summary.ViolationsByRule) {
			fmt.Fprintf(&b, "- %s: %d\n", rule, summary.ViolationsByRule[rule])
		}
	}

	b.WriteString("\n## Next Step\n\n")
	b.WriteString(summary.NextStep + "\n")
	return b.String()
}

func formatRuleCounts(counts map[string]int) string {
	if len(counts) == 0 {
		return ""
	}
	parts := make([]string, 0, len(counts))
	for _, rule := range sortedRuleKeys(counts) {
		parts = append(parts, fmt.Sprintf("%s:%d", rule, counts[rule]))
	}
	return strings.Join(parts, ";")
}

func sortedRuleKeys(counts map[string]int) []string {
	keys := make([]string, 0, len(counts))
	for key := range counts {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func formatEnforcementFloat(value float64) string {
	return strconv.FormatFloat(value, 'f', 6, 64)
}
