package riskmetrics

import "testing"

func TestEvaluatePromotionPromoteToResearch(t *testing.T) {
	decision := EvaluatePromotionWithConfig(PromotionInput{
		Rank:       1,
		Strategy:   "clean",
		Decision:   DecisionAllow,
		Score:      20,
		Expectancy: 0.2,
		RiskGrade:  GradeB,
	}, testPromotionConfig())

	if decision.Gate != PromoteToResearch {
		t.Fatalf("gate = %s, want %s, reasons=%v", decision.Gate, PromoteToResearch, decision.Reasons)
	}
	assertPromotionReason(t, decision.Reasons, ReasonAcceptableResearchRisk)
}

func TestEvaluatePromotionTopThrottleKeepThrottled(t *testing.T) {
	decision := EvaluatePromotionWithConfig(PromotionInput{
		Rank:       2,
		Strategy:   "top-throttle",
		Decision:   DecisionThrottle,
		Score:      -5,
		Expectancy: 0.1,
		RiskGrade:  GradeC,
	}, testPromotionConfig())

	if decision.Gate != KeepThrottled {
		t.Fatalf("gate = %s, want %s", decision.Gate, KeepThrottled)
	}
	assertPromotionReason(t, decision.Reasons, ReasonTopRankedThrottle)
}

func TestEvaluatePromotionBlocksNegativeExpectancyPromotion(t *testing.T) {
	decision := EvaluatePromotionWithConfig(PromotionInput{
		Rank:       1,
		Strategy:   "bad-ev",
		Decision:   DecisionAllow,
		Score:      20,
		Expectancy: -0.1,
		RiskGrade:  GradeC,
	}, testPromotionConfig())

	if decision.Gate == PromoteToResearch {
		t.Fatalf("negative expectancy should not promote")
	}
	assertPromotionReason(t, decision.Reasons, ReasonNegativePromotionEV)
}

func TestEvaluatePromotionBlocksPoorGradePromotion(t *testing.T) {
	decision := EvaluatePromotionWithConfig(PromotionInput{
		Rank:       1,
		Strategy:   "poor-grade",
		Decision:   DecisionAllow,
		Score:      20,
		Expectancy: 0.1,
		RiskGrade:  GradeD,
	}, testPromotionConfig())

	if decision.Gate == PromoteToResearch {
		t.Fatalf("D/F grade should not promote")
	}
	assertPromotionReason(t, decision.Reasons, ReasonPoorPromotionRiskGrade)
}

func TestEvaluatePromotionBlockRewriteOrRemove(t *testing.T) {
	rewrite := EvaluatePromotionWithConfig(PromotionInput{
		Rank:       8,
		Strategy:   "rewrite",
		Decision:   DecisionBlock,
		Score:      -80,
		Expectancy: 0.1,
		RiskGrade:  GradeC,
	}, testPromotionConfig())
	if rewrite.Gate != RewriteRequired {
		t.Fatalf("gate = %s, want %s", rewrite.Gate, RewriteRequired)
	}

	remove := EvaluatePromotionWithConfig(PromotionInput{
		Rank:       12,
		Strategy:   "remove",
		Decision:   DecisionBlock,
		Score:      -130,
		Expectancy: -0.1,
		RiskGrade:  GradeF,
	}, testPromotionConfig())
	if remove.Gate != RemoveFromCandidates {
		t.Fatalf("gate = %s, want %s", remove.Gate, RemoveFromCandidates)
	}
}

func testPromotionConfig() PromotionConfig {
	return PromotionConfig{
		TopThrottleRanks: 3,
		RemoveBelowScore: -100,
	}
}

func assertPromotionReason(t *testing.T, reasons []string, want string) {
	t.Helper()
	for _, reason := range reasons {
		if reason == want {
			return
		}
	}
	t.Fatalf("reasons = %v, missing %s", reasons, want)
}
