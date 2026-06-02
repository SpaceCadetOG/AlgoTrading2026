package vwap

import "AlgoTrading2026/exchanges"

func SessionVWAP(candles []exchanges.Candle) []float64 {
	out := make([]float64, len(candles))
	var cumulativePV float64
	var cumulativeVolume float64

	for i, candle := range candles {
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

func typicalPrice(candle exchanges.Candle) float64 {
	high := candle.HighFloat()
	low := candle.LowFloat()
	close := candle.CloseFloat()
	if high > 0 && low > 0 && close > 0 {
		return (high + low + close) / 3
	}
	return close
}

func positiveVolume(volume float64) float64 {
	if volume < 0 {
		return 0
	}
	return volume
}
