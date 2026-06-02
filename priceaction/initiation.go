package priceaction

import "AlgoTrading2026/exchanges"

type InitiationDirection string

const (
	InitiationNone InitiationDirection = "none"
	InitiationUp   InitiationDirection = "up"
	InitiationDown InitiationDirection = "down"
)

type InitiationFeature struct {
	Detected   bool
	Direction  InitiationDirection
	StartPrice float64
	OpenDrive  bool
}

func Initiations(candles []exchanges.Candle, aggression []AggressionFeature, sideways []SidewaysFeature) []InitiationFeature {
	out := make([]InitiationFeature, len(candles))
	for i, candle := range candles {
		if i == 0 || i >= len(aggression) || i >= len(sideways) {
			continue
		}
		zone := latestSideways(sideways, i-1, 12)
		if !zone.Detected {
			continue
		}
		switch {
		case aggression[i].Direction == AggressionBullish && candle.CloseFloat() > zone.High:
			out[i] = InitiationFeature{Detected: true, Direction: InitiationUp, StartPrice: candle.OpenFloat(), OpenDrive: nearSessionOpen(candles, i)}
		case aggression[i].Direction == AggressionBearish && candle.CloseFloat() < zone.Low:
			out[i] = InitiationFeature{Detected: true, Direction: InitiationDown, StartPrice: candle.OpenFloat(), OpenDrive: nearSessionOpen(candles, i)}
		}
	}
	return out
}

func latestSideways(sideways []SidewaysFeature, index int, maxLookback int) SidewaysFeature {
	start := index - maxLookback
	if start < 0 {
		start = 0
	}
	for i := index; i >= start; i-- {
		if sideways[i].Detected {
			return sideways[i]
		}
	}
	return SidewaysFeature{}
}

func nearSessionOpen(candles []exchanges.Candle, index int) bool {
	if index < 0 || index >= len(candles) {
		return false
	}
	day := candles[index].StartUTC().YearDay()
	var dayIndex int
	for i := index; i >= 0; i-- {
		if candles[i].StartUTC().YearDay() != day {
			break
		}
		dayIndex++
	}
	return dayIndex <= 4
}
