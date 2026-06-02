package priceaction

import "AlgoTrading2026/exchanges"

type StudyDirection string

const (
	StudyLong  StudyDirection = "long"
	StudyShort StudyDirection = "short"
)

type StrategySetup struct {
	Timestamp       int64
	Strategy        string
	Direction       StudyDirection
	Level           float64
	EntryZone       float64
	Confirmation    bool
	Invalidated     bool
	FollowThrough5  float64
	FollowThrough10 float64
	FollowThrough20 float64
	Notes           string
}

func SRFlipSetups(candles []exchanges.Candle) []StrategySetup {
	rejections := Rejections(candles)
	flips := SupportResistanceFlips(candles, rejections)
	out := make([]StrategySetup, 0)
	for i, flip := range flips {
		if !flip.Detected || i >= len(candles) {
			continue
		}
		direction := StudyShort
		if flip.Direction == FlipResistanceToSupport {
			direction = StudyLong
		}
		out = append(out, setupAt(candles, i, "sr_flip", direction, flip.Level, flip.Level, true, invalidatedByClose(candles[i], direction, flip.Level), "breached level retested as flipped support/resistance"))
	}
	return out
}

func setupAt(candles []exchanges.Candle, i int, strategy string, direction StudyDirection, level float64, entryZone float64, confirmation bool, invalidated bool, notes string) StrategySetup {
	return StrategySetup{
		Timestamp:       candles[i].StartTime,
		Strategy:        strategy,
		Direction:       direction,
		Level:           level,
		EntryZone:       entryZone,
		Confirmation:    confirmation,
		Invalidated:     invalidated,
		FollowThrough5:  followThroughByDirection(candles, i, 5, direction),
		FollowThrough10: followThroughByDirection(candles, i, 10, direction),
		FollowThrough20: followThroughByDirection(candles, i, 20, direction),
		Notes:           notes,
	}
}

func invalidatedByClose(candle exchanges.Candle, direction StudyDirection, level float64) bool {
	if level == 0 {
		return false
	}
	switch direction {
	case StudyLong:
		return candle.CloseFloat() < level
	case StudyShort:
		return candle.CloseFloat() > level
	default:
		return false
	}
}

func followThroughByDirection(candles []exchanges.Candle, i int, lookahead int, direction StudyDirection) float64 {
	if i < 0 || i >= len(candles) || lookahead <= 0 {
		return 0
	}
	j := i + lookahead
	if j >= len(candles) {
		j = len(candles) - 1
	}
	move := candles[j].CloseFloat() - candles[i].CloseFloat()
	if direction == StudyShort {
		return -move
	}
	return move
}

func touchesLevel(candle exchanges.Candle, level float64) bool {
	if level <= 0 {
		return false
	}
	tolerance := level * 0.001
	return candle.LowFloat() <= level+tolerance && candle.HighFloat() >= level-tolerance
}
