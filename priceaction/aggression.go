package priceaction

import (
	"math"

	"AlgoTrading2026/exchanges"
)

type AggressionDirection string

const (
	AggressionNone    AggressionDirection = "none"
	AggressionBullish AggressionDirection = "bullish"
	AggressionBearish AggressionDirection = "bearish"
)

type CandleMetrics struct {
	RangePct     float64
	BodyPct      float64
	UpperWickPct float64
	LowerWickPct float64
}

type AggressionFeature struct {
	Direction AggressionDirection
	Score     float64
}

func Metrics(candle exchanges.Candle) CandleMetrics {
	open := candle.OpenFloat()
	high := candle.HighFloat()
	low := candle.LowFloat()
	close := candle.CloseFloat()
	candleRange := high - low
	if candleRange <= 0 {
		return CandleMetrics{}
	}
	body := math.Abs(close - open)
	upper := high - math.Max(open, close)
	lower := math.Min(open, close) - low
	rangePct := 0.0
	if open > 0 {
		rangePct = candleRange / open * 100
	}
	return CandleMetrics{
		RangePct:     rangePct,
		BodyPct:      body / candleRange,
		UpperWickPct: math.Max(upper, 0) / candleRange,
		LowerWickPct: math.Max(lower, 0) / candleRange,
	}
}

func Aggression(candles []exchanges.Candle, lookback int) []AggressionFeature {
	if lookback <= 0 {
		lookback = 20
	}
	out := make([]AggressionFeature, len(candles))
	for i, candle := range candles {
		avgRange := averageRange(candles, i, lookback)
		avgVolume := averageVolume(candles, i, lookback)
		candleRange := candle.Range()
		volume := candle.VolumeFloat()
		metrics := Metrics(candle)
		rangeRatio := safeRatio(candleRange, avgRange)
		volumeRatio := safeRatio(volume, avgVolume)
		score := rangeRatio*0.45 + volumeRatio*0.35 + metrics.BodyPct*0.20

		direction := AggressionNone
		if score >= 1.25 && metrics.BodyPct >= 0.55 {
			switch {
			case candle.IsBullish():
				direction = AggressionBullish
			case candle.IsBearish():
				direction = AggressionBearish
			}
		}
		out[i] = AggressionFeature{Direction: direction, Score: score}
	}
	return out
}

func averageRange(candles []exchanges.Candle, index int, lookback int) float64 {
	start := index - lookback
	if start < 0 {
		start = 0
	}
	var sum float64
	var count int
	for i := start; i < index; i++ {
		if candles[i].Range() > 0 {
			sum += candles[i].Range()
			count++
		}
	}
	if count == 0 && index >= 0 && index < len(candles) {
		return candles[index].Range()
	}
	return safeAverage(sum, count)
}

func averageVolume(candles []exchanges.Candle, index int, lookback int) float64 {
	start := index - lookback
	if start < 0 {
		start = 0
	}
	var sum float64
	var count int
	for i := start; i < index; i++ {
		if candles[i].VolumeFloat() > 0 {
			sum += candles[i].VolumeFloat()
			count++
		}
	}
	if count == 0 && index >= 0 && index < len(candles) {
		return candles[index].VolumeFloat()
	}
	return safeAverage(sum, count)
}

func safeRatio(value float64, base float64) float64 {
	if base == 0 {
		return 0
	}
	return value / base
}

func safeAverage(sum float64, count int) float64 {
	if count == 0 {
		return 0
	}
	return sum / float64(count)
}
