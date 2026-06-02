package priceaction

import (
	"time"

	"AlgoTrading2026/exchanges"
)

type OpenFeature struct {
	SessionOpenLevel float64
	DailyOpenLevel   float64
}

func Opens(candles []exchanges.Candle) []OpenFeature {
	out := make([]OpenFeature, len(candles))
	var currentDay time.Time
	var dailyOpen float64
	for i, candle := range candles {
		day := utcDay(candle.StartUTC())
		if i == 0 || !day.Equal(currentDay) {
			currentDay = day
			dailyOpen = candle.OpenFloat()
		}
		out[i] = OpenFeature{
			SessionOpenLevel: dailyOpen,
			DailyOpenLevel:   dailyOpen,
		}
	}
	return out
}

func utcDay(t time.Time) time.Time {
	y, m, d := t.UTC().Date()
	return time.Date(y, m, d, 0, 0, 0, 0, time.UTC)
}
