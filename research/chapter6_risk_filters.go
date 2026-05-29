package research

import (
	"encoding/csv"
	"encoding/json"
	"os"
	"strings"

	"AlgoTrading2026/riskmetrics"
)

type Chapter6RiskFilterSummary struct {
	Allow     int                                `json:"allow"`
	Throttle  int                                `json:"throttle"`
	Block     int                                `json:"block"`
	Decisions []riskmetrics.StrategyRiskDecision `json:"decisions"`
}

func EvaluateChapter6RiskFilters(rows []StrategyRiskMetrics, cfg riskmetrics.RiskFilterConfig) []riskmetrics.StrategyRiskDecision {
	decisions := make([]riskmetrics.StrategyRiskDecision, 0, len(rows))
	for _, row := range rows {
		decisions = append(decisions, riskmetrics.EvaluateStrategyRiskWithConfig(riskmetrics.StrategyRiskMetricInput{
			Strategy:     row.Strategy,
			Sharpe:       row.Sharpe,
			Variance:     row.Variance,
			Expectancy:   row.Expectancy,
			TradesPerDay: row.TradesPerDay,
			RiskGrade:    row.RiskGrade,
		}, cfg))
	}
	return decisions
}

func NewChapter6RiskFilterSummary(decisions []riskmetrics.StrategyRiskDecision) Chapter6RiskFilterSummary {
	summary := Chapter6RiskFilterSummary{Decisions: decisions}
	for _, decision := range decisions {
		switch decision.Decision {
		case riskmetrics.DecisionAllow:
			summary.Allow++
		case riskmetrics.DecisionThrottle:
			summary.Throttle++
		case riskmetrics.DecisionBlock:
			summary.Block++
		}
	}
	return summary
}

func WriteChapter6RiskFiltersCSV(path string, rows []riskmetrics.StrategyRiskDecision) error {
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
		"decision",
		"reasons",
		"sharpe",
		"variance",
		"expectancy",
		"trades_per_day",
		"risk_grade",
	}); err != nil {
		return err
	}

	for _, row := range rows {
		if err := writer.Write([]string{
			row.Strategy,
			string(row.Decision),
			strings.Join(row.Reasons, "|"),
			floatToString(row.Sharpe),
			floatToString(row.Variance),
			floatToString(row.Expectancy),
			floatToString(row.TradesPerDay),
			string(row.RiskGrade),
		}); err != nil {
			return err
		}
	}

	return writer.Error()
}

func WriteChapter6RiskFilterSummaryJSON(path string, summary Chapter6RiskFilterSummary) error {
	if err := ensureDir(path); err != nil {
		return err
	}

	body, err := json.MarshalIndent(summary, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, append(body, '\n'), 0644)
}
