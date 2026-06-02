package vwap

type EMAAlignment string

const (
	EMAAlignmentBullish EMAAlignment = "bullish_alignment"
	EMAAlignmentBearish EMAAlignment = "bearish_alignment"
	EMAAlignmentMixed   EMAAlignment = "mixed"
)

type EMARelationship struct {
	PriceVsEMA9  float64
	PriceVsEMA20 float64
	EMA9VsEMA20  float64
	Alignment    EMAAlignment
}

func EMARelationships(close []float64) []EMARelationship {
	ema9 := EMA9(close)
	ema20 := EMA20(close)
	out := make([]EMARelationship, len(close))
	for i, price := range close {
		out[i] = EMARelationship{
			PriceVsEMA9:  price - ema9[i],
			PriceVsEMA20: price - ema20[i],
			EMA9VsEMA20:  ema9[i] - ema20[i],
			Alignment:    ClassifyEMAAlignment(price, ema9[i], ema20[i]),
		}
	}
	return out
}

func ClassifyEMAAlignment(price float64, ema9 float64, ema20 float64) EMAAlignment {
	switch {
	case price > ema9 && ema9 > ema20:
		return EMAAlignmentBullish
	case price < ema9 && ema9 < ema20:
		return EMAAlignmentBearish
	default:
		return EMAAlignmentMixed
	}
}
