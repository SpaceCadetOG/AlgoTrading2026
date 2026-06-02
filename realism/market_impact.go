package realism

type MarketImpactModel struct {
	OrderSize          float64             `json:"orderSize"`
	AverageLiquidity   float64             `json:"averageLiquidity"`
	ParticipationRate  float64             `json:"participationRate"`
	EstimatedImpactBps float64             `json:"estimatedImpactBps"`
	Risk               DislocationSeverity `json:"risk"`
}

func NewMarketImpactModel(orderSize float64, averageLiquidity float64) MarketImpactModel {
	model := MarketImpactModel{
		OrderSize:        nonNegativeFloat(orderSize),
		AverageLiquidity: nonNegativeFloat(averageLiquidity),
	}
	if model.AverageLiquidity > 0 {
		model.ParticipationRate = model.OrderSize / model.AverageLiquidity
	}
	model.EstimatedImpactBps = estimateImpactBps(model.ParticipationRate)
	model.Risk = ClassifyMarketImpactRisk(model)
	return model
}

func DefaultMarketImpactModel() MarketImpactModel {
	return NewMarketImpactModel(1, 0)
}

func ClassifyMarketImpactRisk(model MarketImpactModel) DislocationSeverity {
	switch {
	case model.AverageLiquidity <= 0:
		return SeverityHigh
	case model.ParticipationRate <= 0.01:
		return SeverityLow
	case model.ParticipationRate <= 0.10:
		return SeverityMedium
	default:
		return SeverityHigh
	}
}

func estimateImpactBps(participationRate float64) float64 {
	if participationRate <= 0 {
		return 0
	}
	return participationRate * 10000
}
