package labels

const (
	AcceptanceAccepted = "accepted"
	AcceptanceRejected = "rejected"
	AcceptanceNeutral  = "neutral"

	DirectionalFavorable   = "favorable"
	DirectionalUnfavorable = "unfavorable"
	DirectionalFlat        = "flat"

	TripleBarrierUpperHit    = "upper_hit"
	TripleBarrierLowerHit    = "lower_hit"
	TripleBarrierTimeExpired = "time_expired"
)

type SetupLabelRow struct {
	FollowThrough5     float64
	FollowThrough10    float64
	FollowThrough20    float64
	FT5Bps             float64
	FT10Bps            float64
	FT20Bps            float64
	FT5ATR             float64
	FT10ATR            float64
	FT20ATR            float64
	LabelAcceptance    string
	LabelDirectional20 string
	LabelTripleBarrier string
}

func Bps(value float64, price float64) float64 {
	if price == 0 {
		return 0
	}
	return value / price * 10000
}

func ATRMultiple(value float64, atr float64) float64 {
	if atr == 0 {
		return 0
	}
	return value / atr
}

func AcceptanceLabel(accepted bool, rejected bool) string {
	switch {
	case accepted:
		return AcceptanceAccepted
	case rejected:
		return AcceptanceRejected
	default:
		return AcceptanceNeutral
	}
}

func DirectionalLabel20(followThrough20 float64) string {
	switch {
	case followThrough20 > 0:
		return DirectionalFavorable
	case followThrough20 < 0:
		return DirectionalUnfavorable
	default:
		return DirectionalFlat
	}
}

func TripleBarrierLabel(high, low, close []float64, index int, horizon int, atr float64, direction string) string {
	if index < 0 || index >= len(close) || horizon <= 0 || atr <= 0 {
		return TripleBarrierTimeExpired
	}
	entry := close[index]
	upper := entry + atr
	lower := entry - atr
	end := index + horizon
	if end >= len(close) {
		end = len(close) - 1
	}
	for i := index + 1; i <= end; i++ {
		switch direction {
		case "short_context":
			if low[i] <= lower {
				return TripleBarrierUpperHit
			}
			if high[i] >= upper {
				return TripleBarrierLowerHit
			}
		default:
			if high[i] >= upper {
				return TripleBarrierUpperHit
			}
			if low[i] <= lower {
				return TripleBarrierLowerHit
			}
		}
	}
	return TripleBarrierTimeExpired
}
