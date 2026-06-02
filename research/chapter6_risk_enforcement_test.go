package research

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"AlgoTrading2026/backtest"
	"AlgoTrading2026/risk"
)

func TestBuildChapter6RiskEnforcementSummary(t *testing.T) {
	result := backtest.RiskEnforcedResult{
		Strategy: "test",
		Before:   backtest.Report{Metrics: backtest.Metrics{TotalTrades: 10, NetPnL: -5, MaxDrawdownPct: 10}},
		After:    backtest.Report{Metrics: backtest.Metrics{TotalTrades: 5, NetPnL: -1, MaxDrawdownPct: 5}},
		Violations: []risk.RiskViolation{
			{Rule: risk.RuleStopLoss},
			{Rule: risk.RuleMaxNotional},
		},
		ViolationsByRule: map[string]int{risk.RuleStopLoss: 1, risk.RuleMaxNotional: 1},
		ImprovedPnL:      true,
		ImprovedDrawdown: true,
	}

	summary := BuildChapter6RiskEnforcementSummary([]backtest.RiskEnforcedResult{result}, map[string]bool{"test": true})
	if summary.Strategies != 1 || summary.TradesBefore != 10 || summary.TradesAfter != 5 || summary.Violations != 2 {
		t.Fatalf("unexpected summary: %+v", summary)
	}
	if len(summary.StrategiesImproved) != 1 || len(summary.StrategiesRejected) != 1 {
		t.Fatalf("expected improved and rejected strategy: %+v", summary)
	}
}

func TestChapter6RiskEnforcementWriters(t *testing.T) {
	summary := BuildChapter6RiskEnforcementSummary([]backtest.RiskEnforcedResult{{
		Strategy:         "test",
		Before:           backtest.Report{Metrics: backtest.Metrics{TotalTrades: 1}},
		After:            backtest.Report{Metrics: backtest.Metrics{TotalTrades: 1}},
		ViolationsByRule: map[string]int{},
	}}, nil)

	dir := t.TempDir()
	csvPath := filepath.Join(dir, "chapter6_risk_enforcement.csv")
	if err := WriteChapter6RiskEnforcementCSV(csvPath, summary.Rows); err != nil {
		t.Fatalf("write csv: %v", err)
	}
	body, err := os.ReadFile(csvPath)
	if err != nil {
		t.Fatalf("read csv: %v", err)
	}
	if !strings.Contains(string(body), "strategy,trades_before") {
		t.Fatalf("expected csv header, got %s", string(body))
	}

	jsonPath := filepath.Join(dir, "chapter6_risk_enforcement_summary.json")
	if err := WriteChapter6RiskEnforcementSummaryJSON(jsonPath, summary); err != nil {
		t.Fatalf("write json: %v", err)
	}
	var decoded Chapter6RiskEnforcementSummary
	jsonBody, _ := os.ReadFile(jsonPath)
	if err := json.Unmarshal(jsonBody, &decoded); err != nil {
		t.Fatalf("json should decode: %v", err)
	}

	mdPath := filepath.Join(dir, "chapter6_risk_enforcement.md")
	if err := WriteChapter6RiskEnforcementMarkdown(mdPath, summary); err != nil {
		t.Fatalf("write md: %v", err)
	}
	markdown, _ := os.ReadFile(mdPath)
	if !strings.Contains(string(markdown), "No live or paper trading is enabled") {
		t.Fatalf("expected safety conclusion, got %s", string(markdown))
	}
}
