package volumeprofile

func PointOfControl(profile VolumeProfile) float64 {
	if len(profile.Bins) == 0 {
		return 0
	}
	maxBin := profile.Bins[0]
	for _, bin := range profile.Bins[1:] {
		if bin.Volume > maxBin.Volume {
			maxBin = bin
		}
	}
	return maxBin.Mid
}

func POCIndex(bins []PriceBin) int {
	if len(bins) == 0 {
		return -1
	}
	index := 0
	for i := 1; i < len(bins); i++ {
		if bins[i].Volume > bins[index].Volume {
			index = i
		}
	}
	return index
}
