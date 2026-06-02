package vwap

import (
	"math"

	"AlgoTrading2026/exchanges"
)

type ReactionType string

const (
	ReactionNone   ReactionType = "NONE"
	ReactionBounce ReactionType = "BOUNCE"
	ReactionBreak  ReactionType = "BREAK"
	ReactionChop   ReactionType = "CHOP"
)

type ReactionStudy struct {
	Timestamp        int64
	Reaction         ReactionType
	Close            float64
	VWAP             float64
	DistanceFromVWAP float64
	FollowThrough5   float64
	FollowThrough10  float64
	FollowThrough20  float64
}

func ClassifyReaction(candles []exchanges.Candle, vwapValues []float64, i int) ReactionType {
	if i <= 0 || i >= len(candles) || i >= len(vwapValues) || i-1 >= len(vwapValues) {
		return ReactionNone
	}
	candle := candles[i]
	value := vwapValues[i]
	if !TouchedVWAP(candle.HighFloat(), candle.LowFloat(), value) {
		return ReactionNone
	}

	close := candle.CloseFloat()
	displacement := math.Abs(close - value)
	if value > 0 && displacement/value <= 0.0005 {
		return ReactionChop
	}

	prevClose := candles[i-1].CloseFloat()
	prevVWAP := vwapValues[i-1]
	prevSide := sideOf(prevClose, prevVWAP)
	currentSide := sideOf(close, value)
	if prevSide == 0 || currentSide == 0 {
		return ReactionChop
	}
	if prevSide == currentSide {
		return ReactionBounce
	}
	return ReactionBreak
}

func ReactionStudies(candles []exchanges.Candle, vwapValues []float64) []ReactionStudy {
	n := len(candles)
	if len(vwapValues) < n {
		n = len(vwapValues)
	}
	out := make([]ReactionStudy, 0, n)
	for i := 0; i < n; i++ {
		reaction := ClassifyReaction(candles, vwapValues, i)
		if reaction == ReactionNone {
			continue
		}
		close := candles[i].CloseFloat()
		value := vwapValues[i]
		out = append(out, ReactionStudy{
			Timestamp:        candles[i].StartTime,
			Reaction:         reaction,
			Close:            close,
			VWAP:             value,
			DistanceFromVWAP: close - value,
			FollowThrough5:   followThrough(candles, i, 5),
			FollowThrough10:  followThrough(candles, i, 10),
			FollowThrough20:  followThrough(candles, i, 20),
		})
	}
	return out
}

func sideOf(price float64, value float64) int {
	switch {
	case value == 0 || price == value:
		return 0
	case price > value:
		return 1
	default:
		return -1
	}
}

func followThrough(candles []exchanges.Candle, i int, lookahead int) float64 {
	if i < 0 || i >= len(candles) || lookahead <= 0 {
		return 0
	}
	j := i + lookahead
	if j >= len(candles) {
		j = len(candles) - 1
	}
	return candles[j].CloseFloat() - candles[i].CloseFloat()
}
