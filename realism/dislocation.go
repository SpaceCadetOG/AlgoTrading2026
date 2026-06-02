package realism

import "math"

type BiasDirection string

const (
	BiasOptimistic  BiasDirection = "optimistic"
	BiasPessimistic BiasDirection = "pessimistic"
	BiasMixed       BiasDirection = "mixed"
	BiasUnknown     BiasDirection = "unknown"
)

type DislocationSeverity string

const (
	SeverityLow    DislocationSeverity = "low"
	SeverityMedium DislocationSeverity = "medium"
	SeverityHigh   DislocationSeverity = "high"
)

type SimulationDislocation struct {
	Strategy         string              `json:"strategy"`
	Symbol           string              `json:"symbol"`
	Candles          int                 `json:"candles"`
	ForLoopPnL       float64             `json:"forLoopPnL"`
	EventDrivenPnL   float64             `json:"eventDrivenPnL"`
	PnLDifference    float64             `json:"pnlDifference"`
	PnLDifferencePct float64             `json:"pnlDifferencePct"`
	BiasDirection    BiasDirection       `json:"biasDirection"`
	Severity         DislocationSeverity `json:"severity"`
}

func NewSimulationDislocation(strategy string, symbol string, candles int, forLoopPnL float64, eventDrivenPnL float64) SimulationDislocation {
	diff := eventDrivenPnL - forLoopPnL
	pct := pnlDifferencePct(diff, forLoopPnL)
	return SimulationDislocation{
		Strategy:         strategy,
		Symbol:           symbol,
		Candles:          candles,
		ForLoopPnL:       forLoopPnL,
		EventDrivenPnL:   eventDrivenPnL,
		PnLDifference:    diff,
		PnLDifferencePct: pct,
		BiasDirection:    classifyBias(diff),
		Severity:         classifySeverity(diff, pct),
	}
}

func pnlDifferencePct(diff float64, baseline float64) float64 {
	denom := math.Abs(baseline)
	if denom < 1 {
		if diff == 0 {
			return 0
		}
		return math.Copysign(10000, diff)
	}
	return diff / denom * 100
}

func classifyBias(diff float64) BiasDirection {
	switch {
	case diff > 0:
		return BiasOptimistic
	case diff < 0:
		return BiasPessimistic
	default:
		return BiasUnknown
	}
}

func classifySeverity(diff float64, pct float64) DislocationSeverity {
	absDiff := math.Abs(diff)
	absPct := math.Abs(pct)
	if math.IsInf(absPct, 0) || absDiff >= 1000 || absPct >= 100 {
		return SeverityHigh
	}
	if absDiff >= 100 || absPct >= 25 {
		return SeverityMedium
	}
	return SeverityLow
}
