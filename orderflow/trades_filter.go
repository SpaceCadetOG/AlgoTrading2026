package orderflow

func TradesFilter(bar FootprintBar, threshold float64) []LargeTrade {
	out := make([]LargeTrade, 0)
	for _, level := range bar.Levels {
		volume := level.BidVolume + level.AskVolume
		if volume >= threshold {
			out = append(out, LargeTrade{Price: level.Price, Volume: volume})
		}
	}
	return out
}
