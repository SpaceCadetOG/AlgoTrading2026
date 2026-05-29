package riskmetrics

import "testing"

func TestRiskAdjustedScorePenalizesBlockDecision(t *testing.T) {
	base := RiskRankingInput{
		Strategy:     "base",
		Sharpe:       1,
		Sortino:      1,
		Expectancy:   0.5,
		TradesPerDay: 1,
		RiskGrade:    GradeA,
		Decision:     DecisionAllow,
	}
	blocked := base
	blocked.Decision = DecisionBlock

	if RiskAdjustedScore(blocked) >= RiskAdjustedScore(base) {
		t.Fatalf("blocked score should be lower than allow score")
	}
}

func TestRiskAdjustedScorePenalizesThrottleDecision(t *testing.T) {
	base := RiskRankingInput{
		Strategy:     "base",
		Sharpe:       1,
		Sortino:      1,
		Expectancy:   0.5,
		TradesPerDay: 1,
		RiskGrade:    GradeA,
		Decision:     DecisionAllow,
	}
	throttled := base
	throttled.Decision = DecisionThrottle

	if RiskAdjustedScore(throttled) >= RiskAdjustedScore(base) {
		t.Fatalf("throttled score should be lower than allow score")
	}
	if RiskAdjustedScore(throttled) <= RiskAdjustedScore(RiskRankingInput{
		Strategy:     "blocked",
		Sharpe:       1,
		Sortino:      1,
		Expectancy:   0.5,
		TradesPerDay: 1,
		RiskGrade:    GradeA,
		Decision:     DecisionBlock,
	}) {
		t.Fatalf("throttle penalty should be less severe than block penalty")
	}
}

func TestRiskAdjustedScoreUsesVarianceStdDevExecutionAndHoldTime(t *testing.T) {
	clean := RiskRankingInput{
		Strategy:           "clean",
		Sharpe:             0.5,
		Sortino:            0.5,
		Expectancy:         0.2,
		Variance:           0.1,
		StdDev:             0.1,
		AverageHoldCandles: 4,
		AverageHoldHours:   1,
		TradesPerDay:       1,
		RiskGrade:          GradeB,
		Decision:           DecisionAllow,
	}
	noisy := clean
	noisy.Variance = 2
	noisy.StdDev = 2
	noisy.TradesPerDay = 8
	noisy.AverageHoldCandles = 100
	noisy.AverageHoldHours = 25

	if RiskAdjustedScore(noisy) >= RiskAdjustedScore(clean) {
		t.Fatalf("noisy score should be lower than clean score")
	}
}

func TestRankStrategiesSortsBestToWorst(t *testing.T) {
	rows := RankStrategies([]RiskRankingInput{
		{Strategy: "blocked", Sharpe: 2, Sortino: 2, Expectancy: 1, RiskGrade: GradeA, Decision: DecisionBlock},
		{Strategy: "best", Sharpe: 1, Sortino: 1, Expectancy: 0.5, RiskGrade: GradeA, Decision: DecisionAllow},
		{Strategy: "middle", Sharpe: 0.4, Sortino: 0.4, Expectancy: 0.1, RiskGrade: GradeC, Decision: DecisionThrottle},
	})

	if rows[0].Strategy != "best" {
		t.Fatalf("top strategy = %s, want best", rows[0].Strategy)
	}
	if rows[0].Rank != 1 || rows[1].Rank != 2 || rows[2].Rank != 3 {
		t.Fatalf("unexpected ranks: %+v", rows)
	}
}
