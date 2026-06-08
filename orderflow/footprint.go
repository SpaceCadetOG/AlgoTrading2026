package orderflow

func TotalBidVolume(bar FootprintBar) float64 {
	var total float64
	for _, level := range bar.Levels {
		total += level.BidVolume
	}
	return total
}

func TotalAskVolume(bar FootprintBar) float64 {
	var total float64
	for _, level := range bar.Levels {
		total += level.AskVolume
	}
	return total
}

func TotalVolume(bar FootprintBar) float64 {
	return TotalBidVolume(bar) + TotalAskVolume(bar)
}
