package volumeprofile

func HighVolumeNodes(bins []PriceBin) []PriceBin {
	out := make([]PriceBin, 0)
	for i := 1; i < len(bins)-1; i++ {
		if bins[i].Volume > bins[i-1].Volume && bins[i].Volume > bins[i+1].Volume {
			out = append(out, bins[i])
		}
	}
	return out
}

func LowVolumeNodes(bins []PriceBin) []PriceBin {
	out := make([]PriceBin, 0)
	for i := 1; i < len(bins)-1; i++ {
		if bins[i].Volume < bins[i-1].Volume && bins[i].Volume < bins[i+1].Volume {
			out = append(out, bins[i])
		}
	}
	return out
}

func IsNode(bin PriceBin, nodes []PriceBin) bool {
	for _, node := range nodes {
		if bin.Low == node.Low && bin.High == node.High {
			return true
		}
	}
	return false
}
