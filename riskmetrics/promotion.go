package riskmetrics

type PromotionGate string

const (
	PromoteToResearch    PromotionGate = "PROMOTE_TO_RESEARCH"
	KeepThrottled        PromotionGate = "KEEP_THROTTLED"
	RewriteRequired      PromotionGate = "REWRITE_REQUIRED"
	RemoveFromCandidates PromotionGate = "REMOVE_FROM_CANDIDATES"
)

const (
	ReasonTopRankedThrottle       = "top_ranked_throttle"
	ReasonBlockedByRiskFilter     = "blocked_by_risk_filter"
	ReasonNegativePromotionEV     = "negative_expectancy"
	ReasonPoorPromotionRiskGrade  = "poor_risk_grade"
	ReasonLowRiskAdjustedRank     = "low_risk_adjusted_rank"
	ReasonAcceptableResearchRisk  = "acceptable_research_risk"
	ReasonSevereRiskAdjustedScore = "severe_risk_adjusted_score"
)

type PromotionConfig struct {
	TopThrottleRanks int
	RemoveBelowScore float64
}

type PromotionInput struct {
	Rank       int
	Strategy   string
	Decision   RiskDecision
	Score      float64
	Reasons    []string
	Expectancy float64
	RiskGrade  RiskGrade
}

type PromotionDecision struct {
	Rank     int           `json:"rank"`
	Strategy string        `json:"strategy"`
	Gate     PromotionGate `json:"gate"`
	Reasons  []string      `json:"reasons"`

	FilterDecision RiskDecision `json:"filterDecision"`
	Score          float64      `json:"riskAdjustedScore"`
	Expectancy     float64      `json:"expectancy"`
	RiskGrade      RiskGrade    `json:"riskGrade"`
}

func DefaultPromotionConfig() PromotionConfig {
	return PromotionConfig{
		TopThrottleRanks: 5,
		RemoveBelowScore: -125,
	}
}

func EvaluatePromotion(input PromotionInput) PromotionDecision {
	return EvaluatePromotionWithConfig(input, DefaultPromotionConfig())
}

func EvaluatePromotionWithConfig(input PromotionInput, cfg PromotionConfig) PromotionDecision {
	reasons := make([]string, 0, 4)
	reasons = append(reasons, input.Reasons...)

	negativeExpectancy := input.Expectancy < 0
	poorGrade := input.RiskGrade == GradeD || input.RiskGrade == GradeF

	if negativeExpectancy {
		reasons = appendReason(reasons, ReasonNegativePromotionEV)
	}
	if poorGrade {
		reasons = appendReason(reasons, ReasonPoorPromotionRiskGrade)
	}

	gate := PromoteToResearch
	switch {
	case input.Decision == DecisionBlock:
		reasons = appendReason(reasons, ReasonBlockedByRiskFilter)
		if input.Score <= cfg.RemoveBelowScore || input.RiskGrade == GradeF {
			gate = RemoveFromCandidates
			reasons = appendReason(reasons, ReasonSevereRiskAdjustedScore)
		} else {
			gate = RewriteRequired
		}
	case input.Decision == DecisionThrottle:
		if input.Rank > 0 && input.Rank <= cfg.TopThrottleRanks {
			gate = KeepThrottled
			reasons = appendReason(reasons, ReasonTopRankedThrottle)
		} else {
			gate = RewriteRequired
			reasons = appendReason(reasons, ReasonLowRiskAdjustedRank)
		}
	case negativeExpectancy || poorGrade:
		gate = RewriteRequired
	default:
		reasons = appendReason(reasons, ReasonAcceptableResearchRisk)
	}

	if gate == PromoteToResearch && (negativeExpectancy || poorGrade) {
		gate = RewriteRequired
	}

	return PromotionDecision{
		Rank:           input.Rank,
		Strategy:       input.Strategy,
		Gate:           gate,
		Reasons:        reasons,
		FilterDecision: input.Decision,
		Score:          input.Score,
		Expectancy:     input.Expectancy,
		RiskGrade:      input.RiskGrade,
	}
}

func EvaluatePromotions(inputs []PromotionInput, cfg PromotionConfig) []PromotionDecision {
	out := make([]PromotionDecision, 0, len(inputs))
	for _, input := range inputs {
		out = append(out, EvaluatePromotionWithConfig(input, cfg))
	}
	return out
}
