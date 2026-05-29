package strategy

import (
	"math"
	"strconv"
	"strings"
)

func SafeFloat(value string) float64 {
	parsed, err := strconv.ParseFloat(strings.TrimSpace(value), 64)
	if err != nil || math.IsNaN(parsed) || math.IsInf(parsed, 0) {
		return 0
	}

	return parsed
}

func UnrealizedPnL(side string, size float64, entryPrice float64, markPrice float64) float64 {
	absSize := math.Abs(size)
	switch NormalizeSide(side, size) {
	case "LONG":
		return (markPrice - entryPrice) * absSize
	case "SHORT":
		return (entryPrice - markPrice) * absSize
	default:
		return 0
	}
}

func ExposureUSD(size float64, price float64) float64 {
	return math.Abs(size) * price
}

func ROEPct(pnl float64, exposureUSD float64, leverage float64) float64 {
	if exposureUSD == 0 {
		return 0
	}
	margin := exposureUSD
	if leverage > 0 {
		margin = exposureUSD / leverage
	}
	if margin == 0 {
		return 0
	}

	return (pnl / margin) * 100
}

func SumExposure(positions []PositionState) float64 {
	total := 0.0
	for _, position := range positions {
		total += position.ExposureUSD
	}

	return total
}

func SumOpenPnL(positions []PositionState) float64 {
	total := 0.0
	for _, position := range positions {
		total += position.UnrealizedPnL
	}

	return total
}
