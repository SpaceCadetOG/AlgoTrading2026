package priceaction

import (
	"time"

	"AlgoTrading2026/exchanges"
)

type HighLowFeature struct {
	PreviousDailyHigh  float64
	PreviousDailyLow   float64
	PreviousWeeklyHigh float64
	PreviousWeeklyLow  float64
	NearPreviousHigh   bool
	NearPreviousLow    bool
	StrongHigh         bool
	StrongLow          bool
	WeakHigh           bool
	WeakLow            bool
}

func HighsLows(candles []exchanges.Candle, rejections []RejectionFeature, aggression []AggressionFeature) []HighLowFeature {
	out := make([]HighLowFeature, len(candles))
	daily := completedDailyLevels(candles)
	weekly := completedWeeklyLevels(candles)
	for i, candle := range candles {
		day := utcDay(candle.StartUTC())
		week := utcWeek(candle.StartUTC())
		prevDaily := previousDailyLevel(daily, day)
		prevWeekly := previousWeeklyLevel(weekly, week)
		row := HighLowFeature{
			PreviousDailyHigh:  prevDaily.High,
			PreviousDailyLow:   prevDaily.Low,
			PreviousWeeklyHigh: prevWeekly.High,
			PreviousWeeklyLow:  prevWeekly.Low,
		}
		row.NearPreviousHigh = nearLevel(candle.CloseFloat(), row.PreviousDailyHigh) || nearLevel(candle.CloseFloat(), row.PreviousWeeklyHigh) || candle.HighFloat() >= row.PreviousDailyHigh && row.PreviousDailyHigh > 0
		row.NearPreviousLow = nearLevel(candle.CloseFloat(), row.PreviousDailyLow) || nearLevel(candle.CloseFloat(), row.PreviousWeeklyLow) || candle.LowFloat() <= row.PreviousDailyLow && row.PreviousDailyLow > 0
		if i < len(rejections) && i < len(aggression) {
			row.StrongHigh = rejections[i].Detected && rejections[i].Direction == RejectionBearish && nextOppositeAggression(aggression, i, AggressionBearish)
			row.StrongLow = rejections[i].Detected && rejections[i].Direction == RejectionBullish && nextOppositeAggression(aggression, i, AggressionBullish)
			row.WeakHigh = row.NearPreviousHigh && !row.StrongHigh
			row.WeakLow = row.NearPreviousLow && !row.StrongLow
		}
		out[i] = row
	}
	return out
}

type periodLevel struct {
	Start time.Time
	High  float64
	Low   float64
}

func completedDailyLevels(candles []exchanges.Candle) []periodLevel {
	return completedLevels(candles, utcDay)
}

func completedWeeklyLevels(candles []exchanges.Candle) []periodLevel {
	return completedLevels(candles, utcWeek)
}

func completedLevels(candles []exchanges.Candle, bucket func(time.Time) time.Time) []periodLevel {
	var out []periodLevel
	var current periodLevel
	for i, candle := range candles {
		start := bucket(candle.StartUTC())
		if i == 0 || !start.Equal(current.Start) {
			if i > 0 {
				out = append(out, current)
			}
			current = periodLevel{Start: start, High: candle.HighFloat(), Low: candle.LowFloat()}
			continue
		}
		if candle.HighFloat() > current.High {
			current.High = candle.HighFloat()
		}
		if candle.LowFloat() < current.Low {
			current.Low = candle.LowFloat()
		}
	}
	if len(candles) > 0 {
		out = append(out, current)
	}
	return out
}

func previousDailyLevel(levels []periodLevel, day time.Time) periodLevel {
	var previous periodLevel
	for _, level := range levels {
		if level.Start.Before(day) {
			previous = level
		}
	}
	return previous
}

func previousWeeklyLevel(levels []periodLevel, week time.Time) periodLevel {
	var previous periodLevel
	for _, level := range levels {
		if level.Start.Before(week) {
			previous = level
		}
	}
	return previous
}

func utcWeek(t time.Time) time.Time {
	day := utcDay(t)
	weekday := int(day.Weekday())
	if weekday == 0 {
		weekday = 7
	}
	return day.AddDate(0, 0, -(weekday - 1))
}

func nearLevel(price float64, level float64) bool {
	if price <= 0 || level <= 0 {
		return false
	}
	return abs(price-level)/level <= 0.001
}

func nextOppositeAggression(aggression []AggressionFeature, index int, direction AggressionDirection) bool {
	end := index + 3
	if end >= len(aggression) {
		end = len(aggression) - 1
	}
	for i := index + 1; i <= end; i++ {
		if aggression[i].Direction == direction {
			return true
		}
	}
	return false
}
