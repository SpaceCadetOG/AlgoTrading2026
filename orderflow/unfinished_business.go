package orderflow

func DetectUnfinishedBusiness(bar FootprintBar) UnfinishedBusiness {
	var out UnfinishedBusiness
	for _, level := range bar.Levels {
		if level.Price == bar.High && level.BidVolume > 0 && level.AskVolume > 0 {
			out.High = true
		}
		if level.Price == bar.Low && level.BidVolume > 0 && level.AskVolume > 0 {
			out.Low = true
		}
	}
	return out
}
