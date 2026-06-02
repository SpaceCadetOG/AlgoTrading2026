package features

import "math"

const (
	VolatilityLow    = "low"
	VolatilityNormal = "normal"
	VolatilityHigh   = "high"
)

type SetupFeatureRow struct {
	SetupType string
	SetupID   string
	Timestamp int64
	Direction string

	Price      float64
	POC        float64
	VAH        float64
	VAL        float64
	NearestHVN float64
	NearestLVN float64

	ProfileShape         string
	ShapeConfidence      float64
	ProfileVolume        float64
	VWAP                 float64
	VWAPSlope            float64
	DistanceToVWAPPct    float64
	DistanceToPOCPct     float64
	DistanceToVAHPct     float64
	DistanceToVALPct     float64
	DistanceToHVNPct     float64
	DistanceToLVNPct     float64
	TimeInValueArea      bool
	ATR14                float64
	ATRPct               float64
	VolatilityState      string
	Session              string
	DailyOpenDistancePct float64

	VWAPAlignment   string
	POCConfluence   bool
	HVNConfluence   bool
	VAHVALRejection bool
	Accepted        bool
	Rejected        bool
	Invalidated     bool
}

func DistancePct(price float64, level float64) float64 {
	if level == 0 {
		return 0
	}
	return (price - level) / level
}

func ATR(high, low, close []float64, period int) []float64 {
	out := make([]float64, len(close))
	if period <= 0 || len(close) == 0 {
		return out
	}
	tr := make([]float64, len(close))
	for i := range close {
		if i == 0 {
			tr[i] = high[i] - low[i]
			continue
		}
		a := high[i] - low[i]
		b := math.Abs(high[i] - close[i-1])
		c := math.Abs(low[i] - close[i-1])
		tr[i] = math.Max(a, math.Max(b, c))
	}
	for i := range tr {
		if i+1 < period {
			continue
		}
		var sum float64
		for j := i + 1 - period; j <= i; j++ {
			sum += tr[j]
		}
		out[i] = sum / float64(period)
	}
	return out
}

func VolatilityState(atrPct float64) string {
	switch {
	case atrPct <= 0:
		return VolatilityNormal
	case atrPct < 0.003:
		return VolatilityLow
	case atrPct > 0.01:
		return VolatilityHigh
	default:
		return VolatilityNormal
	}
}

func InValueArea(price float64, vah float64, val float64) bool {
	if vah == 0 || val == 0 {
		return false
	}
	if val > vah {
		val, vah = vah, val
	}
	return price >= val && price <= vah
}
