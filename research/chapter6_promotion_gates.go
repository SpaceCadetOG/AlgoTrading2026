package research

import (
	"encoding/csv"
	"encoding/json"
	"os"
	"strconv"
	"strings"

	"AlgoTrading2026/riskmetrics"
)

type Chapter6PromotionGatesSummary struct {
	PromoteToResearch    int                             `json:"promoteToResearch"`
	KeepThrottled        int                             `json:"keepThrottled"`
	RewriteRequired      int                             `json:"rewriteRequired"`
	RemoveFromCandidates int                             `json:"removeFromCandidates"`
	Decisions            []riskmetrics.PromotionDecision `json:"decisions"`
}

func BuildChapter6PromotionGates(rankings []riskmetrics.RankedStrategy, cfg riskmetrics.PromotionConfig) []riskmetrics.PromotionDecision {
	inputs := make([]riskmetrics.PromotionInput, 0, len(rankings))
	for _, row := range rankings {
		inputs = append(inputs, riskmetrics.PromotionInput{
			Rank:       row.Rank,
			Strategy:   row.Strategy,
			Decision:   row.Decision,
			Score:      row.Score,
			Reasons:    row.Reasons,
			Expectancy: row.Expectancy,
			RiskGrade:  row.RiskGrade,
		})
	}
	return riskmetrics.EvaluatePromotions(inputs, cfg)
}

func NewChapter6PromotionGatesSummary(decisions []riskmetrics.PromotionDecision) Chapter6PromotionGatesSummary {
	summary := Chapter6PromotionGatesSummary{Decisions: decisions}
	for _, decision := range decisions {
		switch decision.Gate {
		case riskmetrics.PromoteToResearch:
			summary.PromoteToResearch++
		case riskmetrics.KeepThrottled:
			summary.KeepThrottled++
		case riskmetrics.RewriteRequired:
			summary.RewriteRequired++
		case riskmetrics.RemoveFromCandidates:
			summary.RemoveFromCandidates++
		}
	}
	return summary
}

func WriteChapter6PromotionGatesCSV(path string, rows []riskmetrics.PromotionDecision) error {
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
		"gate",
		"filter_decision",
		"risk_adjusted_score",
		"expectancy",
		"risk_grade",
		"reasons",
	}); err != nil {
		return err
	}

	for _, row := range rows {
		if err := writer.Write([]string{
			strconv.Itoa(row.Rank),
			row.Strategy,
			string(row.Gate),
			string(row.FilterDecision),
			floatToString(row.Score),
			floatToString(row.Expectancy),
			string(row.RiskGrade),
			strings.Join(row.Reasons, "|"),
		}); err != nil {
			return err
		}
	}

	return writer.Error()
}

func WriteChapter6PromotionGatesSummaryJSON(path string, summary Chapter6PromotionGatesSummary) error {
	if err := ensureDir(path); err != nil {
		return err
	}

	body, err := json.MarshalIndent(summary, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, append(body, '\n'), 0644)
}
