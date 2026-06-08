package orderflow

func Delta(bar FootprintBar) float64 {
	return TotalAskVolume(bar) - TotalBidVolume(bar)
}
