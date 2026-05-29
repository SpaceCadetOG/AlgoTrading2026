package riskmetrics

import "testing"

func TestEvaluateStrategyRiskAllow(t *testing.T) {
	decision := EvaluateStrategyRiskWithConfig(StrategyRiskMetricInput{
		Strategy:     "clean",
		Sharpe:       0.8,
		Variance:     0.1,
		Expectancy:   0.2,
		TradesPerDay: 1,
		RiskGrade:    GradeB,
	}, testFilterConfig())

	if decision.Decision != DecisionAllow {
		t.Fatalf("decision = %s, want %s, reasons=%v", decision.Decision, DecisionAllow, decision.Reasons)
	}
	if len(decision.Reasons) != 0 {
		t.Fatalf("reasons = %v, want none", decision.Reasons)
	}
}

func TestEvaluateStrategyRiskThrottle(t *testing.T) {
	decision := EvaluateStrategyRiskWithConfig(StrategyRiskMetricInput{
		Strategy:     "soft-risk",
		Sharpe:       -0.1,
		Variance:     0.1,
		Expectancy:   0.1,
		TradesPerDay: 1,
		RiskGrade:    GradeC,
	}, testFilterConfig())

	if decision.Decision != DecisionThrottle {
		t.Fatalf("decision = %s, want %s", decision.Decision, DecisionThrottle)
	}
	assertHasReason(t, decision.Reasons, ReasonLowSharpe)
}

func TestEvaluateStrategyRiskBlock(t *testing.T) {
	decision := EvaluateStrategyRiskWithConfig(StrategyRiskMetricInput{
		Strategy:     "blocked",
		Sharpe:       0.2,
		Variance:     0.1,
		Expectancy:   0.1,
		TradesPerDay: 1,
		RiskGrade:    GradeF,
	}, testFilterConfig())

	if decision.Decision != DecisionBlock {
		t.Fatalf("decision = %s, want %s", decision.Decision, DecisionBlock)
	}
	assertHasReason(t, decision.Reasons, ReasonPoorRiskGrade)
}

func TestEvaluateStrategyRiskMultipleReasons(t *testing.T) {
	decision := EvaluateStrategyRiskWithConfig(StrategyRiskMetricInput{
		Strategy:     "messy",
		Sharpe:       -0.7,
		Variance:     1.5,
		Expectancy:   -0.2,
		TradesPerDay: 7,
		RiskGrade:    GradeD,
	}, testFilterConfig())

	if decision.Decision != DecisionBlock {
		t.Fatalf("decision = %s, want %s", decision.Decision, DecisionBlock)
	}

	for _, reason := range []string{
		ReasonLowSharpe,
		ReasonHighVariance,
		ReasonNegativeExpectancy,
		ReasonExcessiveTradesPerDay,
		ReasonPoorRiskGrade,
	} {
		assertHasReason(t, decision.Reasons, reason)
	}
}

func testFilterConfig() RiskFilterConfig {
	return RiskFilterConfig{
		MinSharpeAllow:       0,
		MinSharpeBlock:       -0.5,
		MaxVarianceAllow:     0.5,
		MaxVarianceBlock:     1,
		MaxTradesPerDayAllow: 3,
		MaxTradesPerDayBlock: 6,
		ThrottleGrades:       []RiskGrade{GradeD},
		BlockGrades:          []RiskGrade{GradeF},
	}
}

func assertHasReason(t *testing.T, reasons []string, want string) {
	t.Helper()
	for _, reason := range reasons {
		if reason == want {
			return
		}
	}
	t.Fatalf("reasons = %v, missing %s", reasons, want)
}
