package vwap

import "AlgoTrading2026/exchanges"

type Interaction struct {
	Timestamp        int64
	Close            float64
	High             float64
	Low              float64
	Volume           float64
	SessionVWAP      float64
	DistanceFromVWAP float64
	AboveVWAP        bool
	BelowVWAP        bool
	TouchedVWAP      bool
	CrossedVWAP      bool
	Reaction         ReactionType
}

func Interactions(candles []exchanges.Candle, vwapValues []float64) []Interaction {
	n := len(candles)
	if len(vwapValues) < n {
		n = len(vwapValues)
	}
	out := make([]Interaction, 0, n)
	for i := 0; i < n; i++ {
		candle := candles[i]
		value := vwapValues[i]
		close := candle.CloseFloat()
		high := candle.HighFloat()
		low := candle.LowFloat()
		interaction := Interaction{
			Timestamp:        candle.StartTime,
			Close:            close,
			High:             high,
			Low:              low,
			Volume:           candle.VolumeFloat(),
			SessionVWAP:      value,
			DistanceFromVWAP: close - value,
			AboveVWAP:        close > value,
			BelowVWAP:        close < value,
			TouchedVWAP:      TouchedVWAP(high, low, value),
			CrossedVWAP:      CrossedVWAP(candles, vwapValues, i),
		}
		interaction.Reaction = ClassifyReaction(candles, vwapValues, i)
		out = append(out, interaction)
	}
	return out
}

func TouchedVWAP(high float64, low float64, value float64) bool {
	return value > 0 && low <= value && high >= value
}

func CrossedVWAP(candles []exchanges.Candle, vwapValues []float64, i int) bool {
	if i <= 0 || i >= len(candles) || i >= len(vwapValues) || i-1 >= len(vwapValues) {
		return false
	}
	prevClose := candles[i-1].CloseFloat()
	close := candles[i].CloseFloat()
	prevVWAP := vwapValues[i-1]
	currentVWAP := vwapValues[i]
	if prevVWAP == 0 || currentVWAP == 0 {
		return false
	}
	return (prevClose < prevVWAP && close > currentVWAP) || (prevClose > prevVWAP && close < currentVWAP)
}
