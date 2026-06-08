package strategy

type RiskRule struct {
	Name                     string  `json:"name"`
	RiskPerTradePct          float64 `json:"riskPerTradePct"`
	MaxDailyLossPct          float64 `json:"maxDailyLossPct"`
	MaxOpenPositions         int     `json:"maxOpenPositions"`
	MaxTradesPerSymbolPerDay int     `json:"maxTradesPerSymbolPerDay"`
	MinRewardToRisk          float64 `json:"minRewardToRisk"`
	FundingHazardBlock       bool    `json:"fundingHazardBlock"`
}

func DefaultBookRiskRule() RiskRule {
	return RiskRule{
		Name:                     "Fixed Fractional Book Risk",
		RiskPerTradePct:          0.50,
		MaxDailyLossPct:          2.00,
		MaxOpenPositions:         3,
		MaxTradesPerSymbolPerDay: 2,
		MinRewardToRisk:          1.25,
		FundingHazardBlock:       true,
	}
}
