package backtest

import (
	"sort"

	"AlgoTrading2026/exchanges"
)

type Replay struct {
	Candles []exchanges.Candle
}

func NewReplay(candles []exchanges.Candle) Replay {
	ordered := append([]exchanges.Candle(nil), candles...)
	sort.SliceStable(ordered, func(i int, j int) bool {
		return ordered[i].StartTime < ordered[j].StartTime
	})

	return Replay{Candles: ordered}
}

func (r Replay) Len() int {
	return len(r.Candles)
}
