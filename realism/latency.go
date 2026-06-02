package realism

import "time"

type LatencyModel struct {
	SignalLatency           time.Duration       `json:"signalLatency"`
	StrategyLatency         time.Duration       `json:"strategyLatency"`
	GatewayLatency          time.Duration       `json:"gatewayLatency"`
	ExchangeResponseLatency time.Duration       `json:"exchangeResponseLatency"`
	LatencyVariance         time.Duration       `json:"latencyVariance"`
	TotalExpectedLatency    time.Duration       `json:"totalExpectedLatency"`
	Risk                    DislocationSeverity `json:"risk"`
}

func NewLatencyModel(signalLatency time.Duration, strategyLatency time.Duration, gatewayLatency time.Duration, exchangeResponseLatency time.Duration, latencyVariance time.Duration) LatencyModel {
	model := LatencyModel{
		SignalLatency:           nonNegativeDuration(signalLatency),
		StrategyLatency:         nonNegativeDuration(strategyLatency),
		GatewayLatency:          nonNegativeDuration(gatewayLatency),
		ExchangeResponseLatency: nonNegativeDuration(exchangeResponseLatency),
		LatencyVariance:         nonNegativeDuration(latencyVariance),
	}
	model.TotalExpectedLatency = model.SignalLatency + model.StrategyLatency + model.GatewayLatency + model.ExchangeResponseLatency
	model.Risk = ClassifyLatencyRisk(model)
	return model
}

func DefaultLatencyModel() LatencyModel {
	return NewLatencyModel(0, 0, 0, 0, 0)
}

func ClassifyLatencyRisk(model LatencyModel) DislocationSeverity {
	total := model.TotalExpectedLatency
	if total == 0 && model.LatencyVariance == 0 {
		return SeverityHigh
	}
	switch {
	case total <= 100*time.Millisecond && model.LatencyVariance <= 25*time.Millisecond:
		return SeverityLow
	case total <= 500*time.Millisecond && model.LatencyVariance <= 150*time.Millisecond:
		return SeverityMedium
	default:
		return SeverityHigh
	}
}

func nonNegativeDuration(value time.Duration) time.Duration {
	if value < 0 {
		return 0
	}
	return value
}
