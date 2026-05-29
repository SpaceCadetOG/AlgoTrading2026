package riskmetrics

import "math"

type RiskDecision string

const (
	DecisionAllow    RiskDecision = "ALLOW"
	DecisionThrottle RiskDecision = "THROTTLE"
	DecisionBlock    RiskDecision = "BLOCK"
)

const (
	ReasonLowSharpe             = "low_sharpe"
	ReasonHighVariance          = "high_variance"
	ReasonNegativeExpectancy    = "negative_expectancy"
	ReasonExcessiveTradesPerDay = "excessive_trades_per_day"
	ReasonPoorRiskGrade         = "poor_risk_grade"
)

type RiskFilterConfig struct {
	MinSharpeAllow float64
	MinSharpeBlock float64

	MaxVarianceAllow float64
	MaxVarianceBlock float64

	MaxTradesPerDayAllow float64
	MaxTradesPerDayBlock float64

	ThrottleGrades []RiskGrade
	BlockGrades    []RiskGrade

	BlockNegativeExpectancy bool
}

type StrategyRiskMetricInput struct {
	Strategy string

	Sharpe       float64
	Variance     float64
	Expectancy   float64
	TradesPerDay float64
	RiskGrade    RiskGrade
}

type StrategyRiskDecision struct {
	Strategy string       `json:"strategy"`
	Decision RiskDecision `json:"decision"`
	Reasons  []string     `json:"reasons"`

	Sharpe       float64   `json:"sharpe"`
	Variance     float64   `json:"variance"`
	Expectancy   float64   `json:"expectancy"`
	TradesPerDay float64   `json:"tradesPerDay"`
	RiskGrade    RiskGrade `json:"riskGrade"`
}

func DefaultRiskFilterConfig() RiskFilterConfig {
	return RiskFilterConfig{
		MinSharpeAllow:       0,
		MinSharpeBlock:       -0.50,
		MaxVarianceAllow:     0.50,
		MaxVarianceBlock:     1.00,
		MaxTradesPerDayAllow: 3,
		MaxTradesPerDayBlock: 6,
		ThrottleGrades:       []RiskGrade{GradeD},
		BlockGrades:          []RiskGrade{GradeF},
	}
}

func EvaluateStrategyRisk(metrics StrategyRiskMetricInput) StrategyRiskDecision {
	return EvaluateStrategyRiskWithConfig(metrics, DefaultRiskFilterConfig())
}

func EvaluateStrategyRiskWithConfig(metrics StrategyRiskMetricInput, cfg RiskFilterConfig) StrategyRiskDecision {
	reasons := make([]string, 0, 5)
	block := false

	if isFinite(metrics.Sharpe) && metrics.Sharpe < cfg.MinSharpeAllow {
		reasons = appendReason(reasons, ReasonLowSharpe)
		if metrics.Sharpe < cfg.MinSharpeBlock {
			block = true
		}
	}

	if isFinite(metrics.Variance) && metrics.Variance > cfg.MaxVarianceAllow {
		reasons = appendReason(reasons, ReasonHighVariance)
		if metrics.Variance > cfg.MaxVarianceBlock {
			block = true
		}
	}

	if isFinite(metrics.Expectancy) && metrics.Expectancy < 0 {
		reasons = appendReason(reasons, ReasonNegativeExpectancy)
		if cfg.BlockNegativeExpectancy {
			block = true
		}
	}

	if isFinite(metrics.TradesPerDay) && metrics.TradesPerDay > cfg.MaxTradesPerDayAllow {
		reasons = appendReason(reasons, ReasonExcessiveTradesPerDay)
		if metrics.TradesPerDay > cfg.MaxTradesPerDayBlock {
			block = true
		}
	}

	if containsGrade(cfg.BlockGrades, metrics.RiskGrade) {
		reasons = appendReason(reasons, ReasonPoorRiskGrade)
		block = true
	} else if containsGrade(cfg.ThrottleGrades, metrics.RiskGrade) {
		reasons = appendReason(reasons, ReasonPoorRiskGrade)
	}

	decision := DecisionAllow
	if block {
		decision = DecisionBlock
	} else if len(reasons) > 0 {
		decision = DecisionThrottle
	}

	return StrategyRiskDecision{
		Strategy:     metrics.Strategy,
		Decision:     decision,
		Reasons:      reasons,
		Sharpe:       metrics.Sharpe,
		Variance:     metrics.Variance,
		Expectancy:   metrics.Expectancy,
		TradesPerDay: metrics.TradesPerDay,
		RiskGrade:    metrics.RiskGrade,
	}
}

func containsGrade(grades []RiskGrade, grade RiskGrade) bool {
	for _, candidate := range grades {
		if candidate == grade {
			return true
		}
	}
	return false
}

func appendReason(reasons []string, reason string) []string {
	for _, existing := range reasons {
		if existing == reason {
			return reasons
		}
	}
	return append(reasons, reason)
}

func isFinite(value float64) bool {
	return !math.IsNaN(value) && !math.IsInf(value, 0)
}
