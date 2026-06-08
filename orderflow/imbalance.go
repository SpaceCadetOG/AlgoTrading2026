package orderflow

func DetectImbalances(bar FootprintBar, ratioThreshold float64) []Imbalance {
	if ratioThreshold <= 0 {
		ratioThreshold = 3
	}
	out := make([]Imbalance, 0)
	for _, level := range bar.Levels {
		switch {
		case level.BidVolume == 0 && level.AskVolume > 0:
			out = append(out, Imbalance{Price: level.Price, Direction: ImbalanceAsk, Ratio: ratioThreshold})
		case level.AskVolume == 0 && level.BidVolume > 0:
			out = append(out, Imbalance{Price: level.Price, Direction: ImbalanceBid, Ratio: ratioThreshold})
		case level.BidVolume > 0 && level.AskVolume/level.BidVolume >= ratioThreshold:
			out = append(out, Imbalance{Price: level.Price, Direction: ImbalanceAsk, Ratio: level.AskVolume / level.BidVolume})
		case level.AskVolume > 0 && level.BidVolume/level.AskVolume >= ratioThreshold:
			out = append(out, Imbalance{Price: level.Price, Direction: ImbalanceBid, Ratio: level.BidVolume / level.AskVolume})
		}
	}
	return out
}
