package priceaction

import "AlgoTrading2026/exchanges"

type Phase3Direction string

const (
	Phase3LongContext   Phase3Direction = "long_context"
	Phase3ShortContext  Phase3Direction = "short_context"
	Phase3BreakoutWatch Phase3Direction = "breakout_watch"
)

type Phase3Setup struct {
	Timestamp       int64
	Strategy        string
	Direction       Phase3Direction
	Level           float64
	LevelType       string
	Confirmation    bool
	Invalidated     bool
	FollowThrough5  float64
	FollowThrough10 float64
	FollowThrough20 float64
	Notes           string
}

func HighLowRetestSetups(candles []exchanges.Candle, confirmationCandles int) []Phase3Setup {
	if confirmationCandles <= 0 {
		confirmationCandles = 2
	}
	rejections := Rejections(candles)
	aggression := Aggression(candles, 20)
	levels := HighsLows(candles, rejections, aggression)

	out := make([]Phase3Setup, 0)
	states := map[string]*retestState{
		"daily_high_retest":  {Strategy: "daily_high_retest", LevelType: "previous_daily_high", Direction: Phase3LongContext, HighBreak: true},
		"daily_low_retest":   {Strategy: "daily_low_retest", LevelType: "previous_daily_low", Direction: Phase3ShortContext},
		"weekly_high_retest": {Strategy: "weekly_high_retest", LevelType: "previous_weekly_high", Direction: Phase3LongContext, HighBreak: true},
		"weekly_low_retest":  {Strategy: "weekly_low_retest", LevelType: "previous_weekly_low", Direction: Phase3ShortContext},
	}

	for i := range candles {
		if i >= len(levels) {
			break
		}
		row := levels[i]
		out = append(out, updateRetestState(candles, i, states["daily_high_retest"], row.PreviousDailyHigh, confirmationCandles)...)
		out = append(out, updateRetestState(candles, i, states["daily_low_retest"], row.PreviousDailyLow, confirmationCandles)...)
		out = append(out, updateRetestState(candles, i, states["weekly_high_retest"], row.PreviousWeeklyHigh, confirmationCandles)...)
		out = append(out, updateRetestState(candles, i, states["weekly_low_retest"], row.PreviousWeeklyLow, confirmationCandles)...)
	}
	return out
}

type retestState struct {
	Strategy       string
	LevelType      string
	Direction      Phase3Direction
	HighBreak      bool
	Level          float64
	ConfirmCount   int
	Confirmed      bool
	EmittedAtLevel bool
}

func updateRetestState(candles []exchanges.Candle, i int, state *retestState, level float64, confirmationCandles int) []Phase3Setup {
	if state == nil || level <= 0 {
		return nil
	}
	candle := candles[i]
	if state.Level != level {
		state.Level = level
		state.ConfirmCount = 0
		state.Confirmed = false
		state.EmittedAtLevel = false
	}

	close := candle.CloseFloat()
	if state.HighBreak {
		if close > level {
			state.ConfirmCount++
		} else if !state.Confirmed {
			state.ConfirmCount = 0
		}
	} else {
		if close < level {
			state.ConfirmCount++
		} else if !state.Confirmed {
			state.ConfirmCount = 0
		}
	}
	if state.ConfirmCount >= confirmationCandles {
		state.Confirmed = true
	}
	if !state.Confirmed || state.EmittedAtLevel || state.ConfirmCount <= confirmationCandles {
		return nil
	}
	if !touchesLevel(candle, level) {
		return nil
	}
	if state.HighBreak && close < level {
		return nil
	}
	if !state.HighBreak && close > level {
		return nil
	}

	state.EmittedAtLevel = true
	return []Phase3Setup{phase3SetupAt(
		candles,
		i,
		state.Strategy,
		state.Direction,
		level,
		state.LevelType,
		true,
		phase3Invalidated(candle, state.Direction, level),
		"breach accepted beyond prior daily/weekly level, then retested",
	)}
}

func phase3SetupAt(candles []exchanges.Candle, i int, strategy string, direction Phase3Direction, level float64, levelType string, confirmation bool, invalidated bool, notes string) Phase3Setup {
	return Phase3Setup{
		Timestamp:       candles[i].StartTime,
		Strategy:        strategy,
		Direction:       direction,
		Level:           level,
		LevelType:       levelType,
		Confirmation:    confirmation,
		Invalidated:     invalidated,
		FollowThrough5:  phase3FollowThrough(candles, i, 5, direction),
		FollowThrough10: phase3FollowThrough(candles, i, 10, direction),
		FollowThrough20: phase3FollowThrough(candles, i, 20, direction),
		Notes:           notes,
	}
}

func phase3Invalidated(candle exchanges.Candle, direction Phase3Direction, level float64) bool {
	if level <= 0 {
		return false
	}
	switch direction {
	case Phase3LongContext:
		return candle.CloseFloat() < level
	case Phase3ShortContext:
		return candle.CloseFloat() > level
	default:
		return false
	}
}

func phase3FollowThrough(candles []exchanges.Candle, i int, lookahead int, direction Phase3Direction) float64 {
	if i < 0 || i >= len(candles) || lookahead <= 0 {
		return 0
	}
	j := i + lookahead
	if j >= len(candles) {
		j = len(candles) - 1
	}
	move := candles[j].CloseFloat() - candles[i].CloseFloat()
	switch direction {
	case Phase3ShortContext:
		return -move
	case Phase3BreakoutWatch:
		return move
	default:
		return move
	}
}
