package volatility

type Regime string

const (
	RegimeLow    Regime = "low"
	RegimeNormal Regime = "normal"
	RegimeHigh   Regime = "high"
)

func ClassifyRegime(vol []float64, lookback int, lowMult float64, highMult float64) []Regime {
	out := make([]Regime, len(vol))
	for i := range out {
		out[i] = RegimeNormal
	}
	if lookback <= 0 {
		return out
	}

	avg := rollingMean(vol, lookback)
	for i, value := range vol {
		if avg[i] == 0 {
			continue
		}
		switch {
		case value < avg[i]*lowMult:
			out[i] = RegimeLow
		case value > avg[i]*highMult:
			out[i] = RegimeHigh
		default:
			out[i] = RegimeNormal
		}
	}
	return out
}

func DefaultRegime(vol []float64) []Regime {
	return ClassifyRegime(vol, 50, 0.75, 1.25)
}
