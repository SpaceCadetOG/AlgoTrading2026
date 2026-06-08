package orderflow

func FootprintHVN(bar FootprintBar) FootprintLevel {
	var best FootprintLevel
	var bestVolume float64
	for _, level := range bar.Levels {
		volume := level.BidVolume + level.AskVolume
		if volume > bestVolume {
			best = level
			bestVolume = volume
		}
	}
	return best
}
