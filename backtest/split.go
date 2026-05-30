package backtest

import (
	"sort"

	"AlgoTrading2026/exchanges"
)

const DefaultInSampleRatio = 0.8

func InSampleOutOfSampleSplit(candles []exchanges.Candle, inSampleRatio float64) ([]exchanges.Candle, []exchanges.Candle) {
	if len(candles) == 0 {
		return nil, nil
	}
	if inSampleRatio <= 0 || inSampleRatio >= 1 {
		inSampleRatio = DefaultInSampleRatio
	}

	ordered := append([]exchanges.Candle(nil), candles...)
	sort.SliceStable(ordered, func(i int, j int) bool {
		return ordered[i].StartTime < ordered[j].StartTime
	})

	if len(ordered) == 1 {
		return ordered, nil
	}

	splitIndex := int(float64(len(ordered)) * inSampleRatio)
	if splitIndex < 1 {
		splitIndex = 1
	}
	if splitIndex >= len(ordered) {
		splitIndex = len(ordered) - 1
	}

	inSample := append([]exchanges.Candle(nil), ordered[:splitIndex]...)
	outOfSample := append([]exchanges.Candle(nil), ordered[splitIndex:]...)
	return inSample, outOfSample
}
