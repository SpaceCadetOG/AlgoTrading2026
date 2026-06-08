package orderflow

type StackedImbalanceSetup struct {
	Venue           string
	VenueSymbol     string
	CanonicalSymbol string

	Timestamp int64

	ImbalanceDirection string
	StackedLevels      int

	ZoneLow  float64
	ZoneHigh float64

	LongContext  bool
	ShortContext bool

	Accepted bool
	Rejected bool
	Retested bool

	FollowThrough5  float64
	FollowThrough10 float64
	FollowThrough20 float64

	BarDelta  float64
	BarVolume float64

	HVNPrice  float64
	HVNVolume float64

	NearHVN           bool
	NearVolumeCluster bool
	NearMultipleHVN   bool
}

type StackedImbalanceInput struct {
	Venue           string
	VenueSymbol     string
	CanonicalSymbol string

	StartTimestamp int64
	EndTimestamp   int64
	High           float64
	Low            float64
	Close          float64
	Levels         []FootprintLevel

	BarDelta  float64
	BarVolume float64

	HVNPrice  float64
	HVNVolume float64

	VolumeClusterCount int
}

type StackedImbalanceZone struct {
	Venue           string
	VenueSymbol     string
	CanonicalSymbol string
	ZoneLow         float64
	ZoneHigh        float64
}

type StackedImbalanceConfig struct {
	ImbalanceRatioThreshold float64
	MinStackedLevels        int
	RetestThresholdPct      float64
	ProximityPct            float64
}

func DefaultStackedImbalanceConfig() StackedImbalanceConfig {
	return StackedImbalanceConfig{
		ImbalanceRatioThreshold: 3,
		MinStackedLevels:        3,
		RetestThresholdPct:      0.001,
		ProximityPct:            0.0025,
	}
}

func DetectStackedImbalanceSetups(inputs []StackedImbalanceInput, zones []StackedImbalanceZone, config StackedImbalanceConfig) []StackedImbalanceSetup {
	if config.ImbalanceRatioThreshold <= 0 {
		config.ImbalanceRatioThreshold = DefaultStackedImbalanceConfig().ImbalanceRatioThreshold
	}
	if config.MinStackedLevels <= 0 {
		config.MinStackedLevels = DefaultStackedImbalanceConfig().MinStackedLevels
	}
	if config.RetestThresholdPct <= 0 {
		config.RetestThresholdPct = DefaultStackedImbalanceConfig().RetestThresholdPct
	}
	if config.ProximityPct <= 0 {
		config.ProximityPct = DefaultStackedImbalanceConfig().ProximityPct
	}
	setups := make([]StackedImbalanceSetup, 0)
	for i, input := range inputs {
		bar := FootprintBar{
			Venue: input.Venue, VenueSymbol: input.VenueSymbol, CanonicalSymbol: input.CanonicalSymbol,
			StartTimestamp: input.StartTimestamp, EndTimestamp: input.EndTimestamp,
			High: input.High, Low: input.Low, Close: input.Close, Levels: input.Levels,
		}
		imbalances := DetectImbalances(bar, config.ImbalanceRatioThreshold)
		stacked := DetectStackedImbalances(imbalances, config.MinStackedLevels)
		for _, stack := range stacked {
			setup := StackedImbalanceSetup{
				Venue:              input.Venue,
				VenueSymbol:        input.VenueSymbol,
				CanonicalSymbol:    input.CanonicalSymbol,
				Timestamp:          input.StartTimestamp,
				ImbalanceDirection: stackedDirection(stack.Direction),
				StackedLevels:      stack.Count,
				ZoneLow:            minFloat(stack.StartPrice, stack.EndPrice),
				ZoneHigh:           maxFloat(stack.StartPrice, stack.EndPrice),
				BarDelta:           input.BarDelta,
				BarVolume:          input.BarVolume,
				HVNPrice:           input.HVNPrice,
				HVNVolume:          input.HVNVolume,
				NearHVN:            stackedZoneNearLevel(stack, input.HVNPrice, config.ProximityPct),
				NearVolumeCluster:  input.VolumeClusterCount > 0 && stackedZoneNearLevel(stack, input.HVNPrice, config.ProximityPct),
				NearMultipleHVN:    stackedZoneNearAnyZone(input, stack, zones, config.ProximityPct),
			}
			setup.LongContext = stack.Direction == ImbalanceAsk
			setup.ShortContext = stack.Direction == ImbalanceBid
			outcomeIndex := nextSameStackedMarket(inputs, i)
			if outcomeIndex >= 0 {
				outcome := inputs[outcomeIndex]
				if setup.LongContext {
					setup.Accepted = outcome.Close >= setup.ZoneHigh
					setup.Rejected = outcome.Close < setup.ZoneLow
				}
				if setup.ShortContext {
					setup.Accepted = outcome.Close <= setup.ZoneLow
					setup.Rejected = outcome.Close > setup.ZoneHigh
				}
			}
			setup.Retested = firstStackedRetest(inputs, i, setup, config.RetestThresholdPct) >= 0
			setup.FollowThrough5 = stackedFollowThrough(inputs, i, 5, setup)
			setup.FollowThrough10 = stackedFollowThrough(inputs, i, 10, setup)
			setup.FollowThrough20 = stackedFollowThrough(inputs, i, 20, setup)
			setups = append(setups, setup)
		}
	}
	return setups
}

func stackedDirection(direction ImbalanceDirection) string {
	if direction == ImbalanceAsk {
		return "BUY"
	}
	if direction == ImbalanceBid {
		return "SELL"
	}
	return ""
}

func stackedZoneNearLevel(stack StackedImbalance, level float64, proximityPct float64) bool {
	if level <= 0 {
		return false
	}
	low := minFloat(stack.StartPrice, stack.EndPrice)
	high := maxFloat(stack.StartPrice, stack.EndPrice)
	buffer := level * proximityPct
	return level >= low-buffer && level <= high+buffer
}

func stackedZoneNearAnyZone(input StackedImbalanceInput, stack StackedImbalance, zones []StackedImbalanceZone, proximityPct float64) bool {
	low := minFloat(stack.StartPrice, stack.EndPrice)
	high := maxFloat(stack.StartPrice, stack.EndPrice)
	mid := (low + high) / 2
	if mid <= 0 {
		return false
	}
	buffer := mid * proximityPct
	for _, zone := range zones {
		if input.Venue != zone.Venue || input.VenueSymbol != zone.VenueSymbol || input.CanonicalSymbol != zone.CanonicalSymbol {
			continue
		}
		if low <= zone.ZoneHigh+buffer && high >= zone.ZoneLow-buffer {
			return true
		}
	}
	return false
}

func firstStackedRetest(inputs []StackedImbalanceInput, start int, setup StackedImbalanceSetup, thresholdPct float64) int {
	mid := (setup.ZoneLow + setup.ZoneHigh) / 2
	if mid <= 0 {
		return -1
	}
	buffer := mid * thresholdPct
	for i := start + 1; i < len(inputs); i++ {
		if !sameStackedMarket(inputs[start], inputs[i]) {
			continue
		}
		if inputs[i].Low <= setup.ZoneHigh+buffer && inputs[i].High >= setup.ZoneLow-buffer {
			return i
		}
	}
	return -1
}

func stackedFollowThrough(inputs []StackedImbalanceInput, start int, horizon int, setup StackedImbalanceSetup) float64 {
	count := 0
	mid := (setup.ZoneLow + setup.ZoneHigh) / 2
	for i := start + 1; i < len(inputs); i++ {
		if !sameStackedMarket(inputs[start], inputs[i]) {
			continue
		}
		count++
		if count == horizon {
			if setup.LongContext {
				return inputs[i].Close - mid
			}
			if setup.ShortContext {
				return mid - inputs[i].Close
			}
			return inputs[i].Close - mid
		}
	}
	return 0
}

func nextSameStackedMarket(inputs []StackedImbalanceInput, start int) int {
	for i := start + 1; i < len(inputs); i++ {
		if sameStackedMarket(inputs[start], inputs[i]) {
			return i
		}
	}
	return -1
}

func sameStackedMarket(a StackedImbalanceInput, b StackedImbalanceInput) bool {
	return a.Venue == b.Venue && a.VenueSymbol == b.VenueSymbol && a.CanonicalSymbol == b.CanonicalSymbol
}

func minFloat(a float64, b float64) float64 {
	if a < b {
		return a
	}
	return b
}

func maxFloat(a float64, b float64) float64 {
	if a > b {
		return a
	}
	return b
}
