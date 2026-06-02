package realism

type PlaceInLineEstimate struct {
	OrderSize              float64             `json:"orderSize"`
	VisibleLiquidity       float64             `json:"visibleLiquidity"`
	EstimatedQueuePosition float64             `json:"estimatedQueuePosition"`
	FillProbability        float64             `json:"fillProbability"`
	Risk                   DislocationSeverity `json:"risk"`
}

func NewPlaceInLineEstimate(orderSize float64, visibleLiquidity float64, estimatedQueuePosition float64) PlaceInLineEstimate {
	estimate := PlaceInLineEstimate{
		OrderSize:              nonNegativeFloat(orderSize),
		VisibleLiquidity:       nonNegativeFloat(visibleLiquidity),
		EstimatedQueuePosition: nonNegativeFloat(estimatedQueuePosition),
	}
	estimate.FillProbability = estimatePlaceInLineFillProbability(estimate)
	estimate.Risk = ClassifyPlaceInLineRisk(estimate)
	return estimate
}

func DefaultPlaceInLineEstimate() PlaceInLineEstimate {
	return NewPlaceInLineEstimate(1, 0, 0)
}

func ClassifyPlaceInLineRisk(estimate PlaceInLineEstimate) DislocationSeverity {
	switch {
	case estimate.VisibleLiquidity <= 0:
		return SeverityHigh
	case estimate.FillProbability >= 0.80:
		return SeverityLow
	case estimate.FillProbability >= 0.35:
		return SeverityMedium
	default:
		return SeverityHigh
	}
}

func estimatePlaceInLineFillProbability(estimate PlaceInLineEstimate) float64 {
	if estimate.VisibleLiquidity <= 0 || estimate.OrderSize <= 0 {
		return 0
	}
	availableAfterQueue := estimate.VisibleLiquidity - estimate.EstimatedQueuePosition
	if availableAfterQueue <= 0 {
		return 0
	}
	probability := availableAfterQueue / estimate.OrderSize
	if probability > 1 {
		return 1
	}
	return probability
}

func nonNegativeFloat(value float64) float64 {
	if value < 0 {
		return 0
	}
	return value
}
