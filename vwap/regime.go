package vwap

type Regime string

const (
	RegimeAboveRising  Regime = "above_rising"
	RegimeAboveFalling Regime = "above_falling"
	RegimeBelowRising  Regime = "below_rising"
	RegimeBelowFalling Regime = "below_falling"
	RegimeNeutral      Regime = "neutral"
)

func Regimes(close []float64, vwapValues []float64, slopes []float64) []Regime {
	n := len(close)
	if len(vwapValues) < n {
		n = len(vwapValues)
	}
	if len(slopes) < n {
		n = len(slopes)
	}
	out := make([]Regime, n)
	for i := 0; i < n; i++ {
		out[i] = Classify(close[i], vwapValues[i], slopes[i])
	}
	return out
}

func Classify(close float64, vwapValue float64, slope float64) Regime {
	if close == 0 || vwapValue == 0 || slope == 0 {
		return RegimeNeutral
	}
	switch {
	case close > vwapValue && slope > 0:
		return RegimeAboveRising
	case close > vwapValue && slope < 0:
		return RegimeAboveFalling
	case close < vwapValue && slope > 0:
		return RegimeBelowRising
	case close < vwapValue && slope < 0:
		return RegimeBelowFalling
	default:
		return RegimeNeutral
	}
}
