package vwap

import (
	"math"

	"AlgoTrading2026/exchanges"
)

type MagnetStats struct {
	DistanceBandPct        float64 `json:"distanceBandPct"`
	Events                 int     `json:"events"`
	Returns                int     `json:"returns"`
	ReturnProbability      float64 `json:"returnProbability"`
	AverageCandlesToReturn float64 `json:"averageCandlesToReturn"`
}

func MagnetStudy(candles []exchanges.Candle, vwapValues []float64, bands []float64, lookahead int) []MagnetStats {
	out := make([]MagnetStats, 0, len(bands))
	for _, band := range bands {
		out = append(out, magnetStatsForBand(candles, vwapValues, band, lookahead))
	}
	return out
}

func DefaultMagnetStudy(candles []exchanges.Candle, vwapValues []float64) []MagnetStats {
	return MagnetStudy(candles, vwapValues, []float64{0.25, 0.50, 1.00, 2.00}, 20)
}

func magnetStatsForBand(candles []exchanges.Candle, vwapValues []float64, bandPct float64, lookahead int) MagnetStats {
	n := len(candles)
	if len(vwapValues) < n {
		n = len(vwapValues)
	}
	stats := MagnetStats{DistanceBandPct: bandPct}
	var totalReturnCandles int

	for i := 0; i < n; i++ {
		value := vwapValues[i]
		if value <= 0 {
			continue
		}
		close := candles[i].CloseFloat()
		distancePct := math.Abs(close-value) / value * 100
		if distancePct < bandPct {
			continue
		}
		stats.Events++
		if candlesToReturn, ok := returnsToVWAP(candles, vwapValues, i, lookahead); ok {
			stats.Returns++
			totalReturnCandles += candlesToReturn
		}
	}

	if stats.Events > 0 {
		stats.ReturnProbability = float64(stats.Returns) / float64(stats.Events)
	}
	if stats.Returns > 0 {
		stats.AverageCandlesToReturn = float64(totalReturnCandles) / float64(stats.Returns)
	}
	return stats
}

func returnsToVWAP(candles []exchanges.Candle, vwapValues []float64, i int, lookahead int) (int, bool) {
	if i < 0 || i >= len(candles) || i >= len(vwapValues) {
		return 0, false
	}
	end := i + lookahead
	if end >= len(candles) {
		end = len(candles) - 1
	}
	for j := i + 1; j <= end; j++ {
		if j >= len(vwapValues) {
			break
		}
		if TouchedVWAP(candles[j].HighFloat(), candles[j].LowFloat(), vwapValues[j]) {
			return j - i, true
		}
	}
	return 0, false
}
