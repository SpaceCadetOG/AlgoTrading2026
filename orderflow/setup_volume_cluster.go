package orderflow

type VolumeClusterSetup struct {
	Venue           string
	VenueSymbol     string
	CanonicalSymbol string

	StartTimestamp int64

	ClusterPrice  float64
	ClusterVolume float64

	LongContext  bool
	ShortContext bool

	Retested bool

	Accepted bool
	Rejected bool

	FollowThrough5  float64
	FollowThrough10 float64
	FollowThrough20 float64

	Volume float64
	Delta  float64

	ImbalanceCount        int
	StackedImbalanceCount int
}

type VolumeClusterInput struct {
	Venue                 string
	VenueSymbol           string
	CanonicalSymbol       string
	StartTimestamp        int64
	High                  float64
	Low                   float64
	Close                 float64
	Volume                float64
	Delta                 float64
	HVNPrice              float64
	HVNVolume             float64
	VolumeClusterCount    int
	ImbalanceCount        int
	StackedImbalanceCount int
	LevelCount            int
}

type VolumeClusterSetupConfig struct {
	ClusterVolumeMultiplier float64
	RetestThresholdPct      float64
}

func DefaultVolumeClusterSetupConfig() VolumeClusterSetupConfig {
	return VolumeClusterSetupConfig{
		ClusterVolumeMultiplier: 2.0,
		RetestThresholdPct:      0.001,
	}
}

func DetectVolumeClusterSetups(inputs []VolumeClusterInput, config VolumeClusterSetupConfig) []VolumeClusterSetup {
	if config.ClusterVolumeMultiplier <= 0 {
		config.ClusterVolumeMultiplier = DefaultVolumeClusterSetupConfig().ClusterVolumeMultiplier
	}
	if config.RetestThresholdPct <= 0 {
		config.RetestThresholdPct = DefaultVolumeClusterSetupConfig().RetestThresholdPct
	}
	setups := make([]VolumeClusterSetup, 0)
	for i, row := range inputs {
		if !isClusterCandidate(row, config.ClusterVolumeMultiplier) {
			continue
		}
		setup := VolumeClusterSetup{
			Venue:                 row.Venue,
			VenueSymbol:           row.VenueSymbol,
			CanonicalSymbol:       row.CanonicalSymbol,
			StartTimestamp:        row.StartTimestamp,
			ClusterPrice:          row.HVNPrice,
			ClusterVolume:         row.HVNVolume,
			Volume:                row.Volume,
			Delta:                 row.Delta,
			ImbalanceCount:        row.ImbalanceCount,
			StackedImbalanceCount: row.StackedImbalanceCount,
		}
		contextIndex := firstLaterSameMarket(inputs, i)
		if contextIndex >= 0 {
			next := inputs[contextIndex]
			if next.Close > row.HVNPrice {
				setup.LongContext = true
			} else if next.Close < row.HVNPrice {
				setup.ShortContext = true
			}
		}
		retestIndex := firstRetest(inputs, i, row, config.RetestThresholdPct)
		setup.Retested = retestIndex >= 0
		outcomeIndex := retestIndex
		if outcomeIndex < 0 {
			outcomeIndex = firstLaterSameMarket(inputs, i)
		}
		if outcomeIndex >= 0 {
			outcome := inputs[outcomeIndex]
			if setup.LongContext {
				setup.Accepted = outcome.Close >= row.HVNPrice
				setup.Rejected = outcome.Close < row.HVNPrice
			}
			if setup.ShortContext {
				setup.Accepted = outcome.Close <= row.HVNPrice
				setup.Rejected = outcome.Close > row.HVNPrice
			}
		}
		setup.FollowThrough5 = followThrough(inputs, i, 5, row.HVNPrice, setup.LongContext, setup.ShortContext)
		setup.FollowThrough10 = followThrough(inputs, i, 10, row.HVNPrice, setup.LongContext, setup.ShortContext)
		setup.FollowThrough20 = followThrough(inputs, i, 20, row.HVNPrice, setup.LongContext, setup.ShortContext)
		setups = append(setups, setup)
	}
	return setups
}

func isClusterCandidate(row VolumeClusterInput, multiplier float64) bool {
	if row.VolumeClusterCount <= 0 || row.HVNPrice <= 0 || row.HVNVolume <= 0 || row.Volume <= 0 || row.LevelCount <= 0 {
		return false
	}
	averageLevelVolume := row.Volume / float64(row.LevelCount)
	return row.HVNVolume >= multiplier*averageLevelVolume
}

func firstLaterSameMarket(inputs []VolumeClusterInput, start int) int {
	for i := start + 1; i < len(inputs); i++ {
		if sameMarket(inputs[start], inputs[i]) {
			return i
		}
	}
	return -1
}

func firstRetest(inputs []VolumeClusterInput, start int, row VolumeClusterInput, thresholdPct float64) int {
	for i := start + 1; i < len(inputs); i++ {
		if !sameMarket(row, inputs[i]) {
			continue
		}
		if withinRetest(inputs[i], row.HVNPrice, thresholdPct) {
			return i
		}
	}
	return -1
}

func withinRetest(row VolumeClusterInput, level float64, thresholdPct float64) bool {
	if level <= 0 {
		return false
	}
	buffer := level * thresholdPct
	return row.Low <= level+buffer && row.High >= level-buffer
}

func followThrough(inputs []VolumeClusterInput, start int, horizon int, level float64, longContext bool, shortContext bool) float64 {
	count := 0
	for i := start + 1; i < len(inputs); i++ {
		if !sameMarket(inputs[start], inputs[i]) {
			continue
		}
		count++
		if count == horizon {
			if longContext {
				return inputs[i].Close - level
			}
			if shortContext {
				return level - inputs[i].Close
			}
			return inputs[i].Close - level
		}
	}
	return 0
}

func sameMarket(a VolumeClusterInput, b VolumeClusterInput) bool {
	return a.Venue == b.Venue && a.VenueSymbol == b.VenueSymbol && a.CanonicalSymbol == b.CanonicalSymbol
}
