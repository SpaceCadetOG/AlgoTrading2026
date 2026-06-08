package paper

import "math"

type FundingEvent struct {
	Timestamp  int64   `json:"timestamp"`
	PositionID string  `json:"positionId"`
	Symbol     string  `json:"symbol"`
	Side       string  `json:"side"`
	Rate       float64 `json:"rate"`
	Amount     float64 `json:"amount"`
}

func FundingCharge(position PaperPosition, fundingRate float64) float64 {
	notional := math.Abs(position.MarkPrice * position.Quantity)
	if notional == 0 {
		notional = math.Abs(position.EntryPrice * position.Quantity)
	}
	if notional == 0 {
		return 0
	}
	switch normalizeSide(position.Side) {
	case "LONG":
		return -notional * fundingRate
	case "SHORT":
		return notional * fundingRate
	default:
		return 0
	}
}

func FundingHazardBlocked(rate float64, positionSide string, cfg Config) bool {
	if !cfg.EnableFunding {
		return false
	}
	if math.Abs(rate) < cfg.FundingHazardRate {
		return false
	}
	switch normalizeSide(positionSide) {
	case "LONG":
		return rate > 0
	case "SHORT":
		return rate < 0
	default:
		return false
	}
}
