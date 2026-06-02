package vwap

import "AlgoTrading2026/exchanges"

func AnchoredVWAP(candles []exchanges.Candle, anchorTime int64) []float64 {
	out := make([]float64, len(candles))
	var cumulativePV float64
	var cumulativeVolume float64
	anchored := anchorTime <= 0

	for i, candle := range candles {
		if !anchored && candle.StartTime >= anchorTime {
			anchored = true
			cumulativePV = 0
			cumulativeVolume = 0
		}
		if !anchored {
			continue
		}

		price := typicalPrice(candle)
		volume := positiveVolume(candle.VolumeFloat())
		cumulativePV += price * volume
		cumulativeVolume += volume
		if cumulativeVolume > 0 {
			out[i] = cumulativePV / cumulativeVolume
		}
	}

	return out
}
