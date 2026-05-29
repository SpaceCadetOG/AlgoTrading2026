package riskmetrics

import "sort"

type RiskRankingInput struct {
	Strategy string

	Sharpe     float64
	Sortino    float64
	Expectancy float64

	Variance float64
	StdDev   float64

	AverageHoldCandles float64
	AverageHoldHours   float64

	TradesPerDay float64
	RiskGrade    RiskGrade

	Decision RiskDecision
	Reasons  []string
}

type RankedStrategy struct {
	Rank  int     `json:"rank"`
	Score float64 `json:"riskAdjustedScore"`

	Strategy string       `json:"strategy"`
	Decision RiskDecision `json:"decision"`
	Reasons  []string     `json:"reasons"`

	Sharpe             float64   `json:"sharpe"`
	Sortino            float64   `json:"sortino"`
	Expectancy         float64   `json:"expectancy"`
	Variance           float64   `json:"variance"`
	StdDev             float64   `json:"stdDev"`
	AverageHoldCandles float64   `json:"averageHoldCandles"`
	AverageHoldHours   float64   `json:"averageHoldHours"`
	TradesPerDay       float64   `json:"tradesPerDay"`
	RiskGrade          RiskGrade `json:"riskGrade"`
}

func RiskAdjustedScore(input RiskRankingInput) float64 {
	score := 0.0
	score += input.Sharpe * 25
	score += input.Sortino * 15
	score += input.Expectancy * 20

	score -= input.TradesPerDay * 2
	score -= input.Variance * 10
	score -= input.StdDev * 5
	score -= input.AverageHoldHours * 0.10
	score -= input.AverageHoldCandles * 0.02

	score += gradeScore(input.RiskGrade)

	switch input.Decision {
	case DecisionBlock:
		score -= 100
	case DecisionThrottle:
		score -= 25
	}

	return score
}

func RankStrategies(inputs []RiskRankingInput) []RankedStrategy {
	rows := make([]RankedStrategy, 0, len(inputs))
	for _, input := range inputs {
		rows = append(rows, RankedStrategy{
			Score:              RiskAdjustedScore(input),
			Strategy:           input.Strategy,
			Decision:           input.Decision,
			Reasons:            append([]string(nil), input.Reasons...),
			Sharpe:             input.Sharpe,
			Sortino:            input.Sortino,
			Expectancy:         input.Expectancy,
			Variance:           input.Variance,
			StdDev:             input.StdDev,
			AverageHoldCandles: input.AverageHoldCandles,
			AverageHoldHours:   input.AverageHoldHours,
			TradesPerDay:       input.TradesPerDay,
			RiskGrade:          input.RiskGrade,
		})
	}

	sort.SliceStable(rows, func(i int, j int) bool {
		if rows[i].Score == rows[j].Score {
			return rows[i].Strategy < rows[j].Strategy
		}
		return rows[i].Score > rows[j].Score
	})

	for i := range rows {
		rows[i].Rank = i + 1
	}
	return rows
}

func gradeScore(grade RiskGrade) float64 {
	switch grade {
	case GradeA:
		return 20
	case GradeB:
		return 10
	case GradeC:
		return 0
	case GradeD:
		return -10
	case GradeF:
		return -25
	default:
		return -5
	}
}
