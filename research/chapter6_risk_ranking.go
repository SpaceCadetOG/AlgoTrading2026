package research

import (
	"encoding/csv"
	"encoding/json"
	"os"
	"strconv"
	"strings"

	"AlgoTrading2026/riskmetrics"
)

type Chapter6RiskRankingSummary struct {
	BestStrategy  string                       `json:"bestStrategy"`
	BestScore     float64                      `json:"bestScore"`
	WorstStrategy string                       `json:"worstStrategy"`
	WorstScore    float64                      `json:"worstScore"`
	Rankings      []riskmetrics.RankedStrategy `json:"rankings"`
}

func BuildChapter6RiskRanking(rows []StrategyRiskMetrics, decisions []riskmetrics.StrategyRiskDecision) []riskmetrics.RankedStrategy {
	inputs := make([]riskmetrics.RiskRankingInput, 0, len(rows))
	for i, row := range rows {
		decision := riskmetrics.StrategyRiskDecision{
			Strategy: row.Strategy,
			Decision: riskmetrics.DecisionAllow,
		}
		if i < len(decisions) {
			decision = decisions[i]
		}

		inputs = append(inputs, riskmetrics.RiskRankingInput{
			Strategy:           row.Strategy,
			Sharpe:             row.Sharpe,
			Sortino:            row.Sortino,
			Expectancy:         row.Expectancy,
			Variance:           row.Variance,
			StdDev:             row.StdDev,
			AverageHoldCandles: row.AverageHoldCandles,
			AverageHoldHours:   row.AverageHoldHours,
			TradesPerDay:       row.TradesPerDay,
			RiskGrade:          row.RiskGrade,
			Decision:           decision.Decision,
			Reasons:            decision.Reasons,
		})
	}
	return riskmetrics.RankStrategies(inputs)
}

func NewChapter6RiskRankingSummary(rankings []riskmetrics.RankedStrategy) Chapter6RiskRankingSummary {
	summary := Chapter6RiskRankingSummary{Rankings: rankings}
	if len(rankings) == 0 {
		return summary
	}
	summary.BestStrategy = rankings[0].Strategy
	summary.BestScore = rankings[0].Score
	last := rankings[len(rankings)-1]
	summary.WorstStrategy = last.Strategy
	summary.WorstScore = last.Score
	return summary
}

func WriteChapter6RiskRankingCSV(path string, rows []riskmetrics.RankedStrategy) error {
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
		"rank",
		"strategy",
		"decision",
		"risk_adjusted_score",
		"reasons",
		"sharpe",
		"sortino",
		"expectancy",
		"variance",
		"stddev",
		"average_hold_candles",
		"average_hold_hours",
		"trades_per_day",
		"risk_grade",
	}); err != nil {
		return err
	}

	for _, row := range rows {
		if err := writer.Write([]string{
			strconv.Itoa(row.Rank),
			row.Strategy,
			string(row.Decision),
			floatToString(row.Score),
			strings.Join(row.Reasons, "|"),
			floatToString(row.Sharpe),
			floatToString(row.Sortino),
			floatToString(row.Expectancy),
			floatToString(row.Variance),
			floatToString(row.StdDev),
			floatToString(row.AverageHoldCandles),
			floatToString(row.AverageHoldHours),
			floatToString(row.TradesPerDay),
			string(row.RiskGrade),
		}); err != nil {
			return err
		}
	}

	return writer.Error()
}

func WriteChapter6RiskRankingSummaryJSON(path string, summary Chapter6RiskRankingSummary) error {
	if err := ensureDir(path); err != nil {
		return err
	}

	body, err := json.MarshalIndent(summary, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, append(body, '\n'), 0644)
}
