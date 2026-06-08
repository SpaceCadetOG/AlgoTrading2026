package orderflow

func DetectVolumeClusters(bar FootprintBar, multiplier float64) []VolumeCluster {
	if multiplier <= 0 {
		multiplier = 1.5
	}
	if len(bar.Levels) == 0 {
		return nil
	}
	avg := TotalVolume(bar) / float64(len(bar.Levels))
	out := make([]VolumeCluster, 0)
	for _, level := range bar.Levels {
		volume := level.BidVolume + level.AskVolume
		if volume >= avg*multiplier {
			out = append(out, VolumeCluster{Price: level.Price, Volume: volume})
		}
	}
	return out
}
