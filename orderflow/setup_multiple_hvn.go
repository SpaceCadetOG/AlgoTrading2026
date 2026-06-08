package orderflow

type HVNZone struct {
	LowerPrice       float64
	UpperPrice       float64
	HVNCount         int
	AverageHVNVolume float64
	StartTimestamp   int64
	EndTimestamp     int64
}

type MultipleHVNSetup struct {
	Venue           string
	VenueSymbol     string
	CanonicalSymbol string

	Timestamp int64

	ZoneLow  float64
	ZoneHigh float64

	HVNCount int

	AverageHVNVolume float64

	LongContext  bool
	ShortContext bool

	Retested bool

	Accepted bool
	Rejected bool

	FollowThrough5  float64
	FollowThrough10 float64
	FollowThrough20 float64

	Delta float64

	ImbalanceCount        int
	StackedImbalanceCount int
}

type MultipleHVNInput struct {
	Venue                 string
	VenueSymbol           string
	CanonicalSymbol       string
	StartTimestamp        int64
	High                  float64
	Low                   float64
	Close                 float64
	HVNPrice              float64
	HVNVolume             float64
	Delta                 float64
	ImbalanceCount        int
	StackedImbalanceCount int
}

type MultipleHVNConfig struct {
	ProximityPct       float64
	MaxBarsApart       int
	RetestThresholdPct float64
}

func DefaultMultipleHVNConfig() MultipleHVNConfig {
	return MultipleHVNConfig{
		ProximityPct:       0.0025,
		MaxBarsApart:       5,
		RetestThresholdPct: 0.001,
	}
}

func DetectMultipleHVNSetups(inputs []MultipleHVNInput, config MultipleHVNConfig) []MultipleHVNSetup {
	if config.ProximityPct <= 0 {
		config.ProximityPct = DefaultMultipleHVNConfig().ProximityPct
	}
	if config.MaxBarsApart <= 0 {
		config.MaxBarsApart = DefaultMultipleHVNConfig().MaxBarsApart
	}
	if config.RetestThresholdPct <= 0 {
		config.RetestThresholdPct = DefaultMultipleHVNConfig().RetestThresholdPct
	}
	setups := make([]MultipleHVNSetup, 0)
	seen := map[string]bool{}
	consumed := map[int]bool{}
	for i, row := range inputs {
		if consumed[i] {
			continue
		}
		if row.HVNPrice <= 0 || row.HVNVolume <= 0 {
			continue
		}
		zone, members := buildHVNZone(inputs, i, config)
		if zone.HVNCount < 2 {
			continue
		}
		for _, member := range members {
			consumed[member] = true
		}
		key := row.Venue + "|" + row.VenueSymbol + "|" + row.CanonicalSymbol + "|" + intKey(zone.StartTimestamp) + "|" + intKey(zone.EndTimestamp)
		if seen[key] {
			continue
		}
		seen[key] = true
		lastMemberIndex := members[len(members)-1]
		setup := MultipleHVNSetup{
			Venue:                 row.Venue,
			VenueSymbol:           row.VenueSymbol,
			CanonicalSymbol:       row.CanonicalSymbol,
			Timestamp:             zone.StartTimestamp,
			ZoneLow:               zone.LowerPrice,
			ZoneHigh:              zone.UpperPrice,
			HVNCount:              zone.HVNCount,
			AverageHVNVolume:      zone.AverageHVNVolume,
			Delta:                 averageZoneDelta(inputs, members),
			ImbalanceCount:        sumZoneImbalances(inputs, members),
			StackedImbalanceCount: sumZoneStackedImbalances(inputs, members),
		}
		contextIndex := firstLaterMultipleHVNSameMarket(inputs, lastMemberIndex)
		if contextIndex >= 0 {
			next := inputs[contextIndex]
			if next.Close > zone.UpperPrice {
				setup.LongContext = true
			} else if next.Close < zone.LowerPrice {
				setup.ShortContext = true
			}
		}
		retestIndex := firstHVNZoneRetest(inputs, lastMemberIndex, row, zone, config.RetestThresholdPct)
		setup.Retested = retestIndex >= 0
		outcomeIndex := retestIndex
		if outcomeIndex < 0 {
			outcomeIndex = contextIndex
		}
		if outcomeIndex >= 0 {
			outcome := inputs[outcomeIndex]
			if setup.LongContext {
				setup.Accepted = outcome.Close >= zone.LowerPrice
				setup.Rejected = outcome.Close < zone.LowerPrice
			}
			if setup.ShortContext {
				setup.Accepted = outcome.Close <= zone.UpperPrice
				setup.Rejected = outcome.Close > zone.UpperPrice
			}
		}
		mid := (zone.LowerPrice + zone.UpperPrice) / 2
		setup.FollowThrough5 = multipleHVNFollowThrough(inputs, lastMemberIndex, 5, mid, setup.LongContext, setup.ShortContext)
		setup.FollowThrough10 = multipleHVNFollowThrough(inputs, lastMemberIndex, 10, mid, setup.LongContext, setup.ShortContext)
		setup.FollowThrough20 = multipleHVNFollowThrough(inputs, lastMemberIndex, 20, mid, setup.LongContext, setup.ShortContext)
		setups = append(setups, setup)
	}
	return setups
}

func buildHVNZone(inputs []MultipleHVNInput, start int, config MultipleHVNConfig) (HVNZone, []int) {
	row := inputs[start]
	zone := HVNZone{
		LowerPrice:     row.HVNPrice,
		UpperPrice:     row.HVNPrice,
		StartTimestamp: row.StartTimestamp,
		EndTimestamp:   row.StartTimestamp,
	}
	members := []int{start}
	totalVolume := row.HVNVolume
	for i := start + 1; i < len(inputs) && len(members) < config.MaxBarsApart+1; i++ {
		if !sameMultipleHVNMarket(row, inputs[i]) {
			continue
		}
		if !hvnNearZone(inputs[i].HVNPrice, zone, config.ProximityPct) {
			continue
		}
		members = append(members, i)
		if inputs[i].HVNPrice < zone.LowerPrice {
			zone.LowerPrice = inputs[i].HVNPrice
		}
		if inputs[i].HVNPrice > zone.UpperPrice {
			zone.UpperPrice = inputs[i].HVNPrice
		}
		if inputs[i].StartTimestamp > zone.EndTimestamp {
			zone.EndTimestamp = inputs[i].StartTimestamp
		}
		totalVolume += inputs[i].HVNVolume
	}
	zone.HVNCount = len(members)
	if zone.HVNCount > 0 {
		zone.AverageHVNVolume = totalVolume / float64(zone.HVNCount)
	}
	return zone, members
}

func hvnNearZone(price float64, zone HVNZone, proximityPct float64) bool {
	if price <= 0 || zone.LowerPrice <= 0 || zone.UpperPrice <= 0 {
		return false
	}
	mid := (zone.LowerPrice + zone.UpperPrice) / 2
	if mid <= 0 {
		return false
	}
	buffer := mid * proximityPct
	return price >= zone.LowerPrice-buffer && price <= zone.UpperPrice+buffer
}

func firstLaterMultipleHVNSameMarket(inputs []MultipleHVNInput, start int) int {
	for i := start + 1; i < len(inputs); i++ {
		if sameMultipleHVNMarket(inputs[start], inputs[i]) {
			return i
		}
	}
	return -1
}

func firstHVNZoneRetest(inputs []MultipleHVNInput, start int, row MultipleHVNInput, zone HVNZone, thresholdPct float64) int {
	for i := start + 1; i < len(inputs); i++ {
		if !sameMultipleHVNMarket(row, inputs[i]) {
			continue
		}
		if multipleHVNWithinZoneRetest(inputs[i], zone, thresholdPct) {
			return i
		}
	}
	return -1
}

func multipleHVNWithinZoneRetest(row MultipleHVNInput, zone HVNZone, thresholdPct float64) bool {
	mid := (zone.LowerPrice + zone.UpperPrice) / 2
	if mid <= 0 {
		return false
	}
	buffer := mid * thresholdPct
	return row.Low <= zone.UpperPrice+buffer && row.High >= zone.LowerPrice-buffer
}

func multipleHVNFollowThrough(inputs []MultipleHVNInput, start int, horizon int, level float64, longContext bool, shortContext bool) float64 {
	count := 0
	for i := start + 1; i < len(inputs); i++ {
		if !sameMultipleHVNMarket(inputs[start], inputs[i]) {
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

func sameMultipleHVNMarket(a MultipleHVNInput, b MultipleHVNInput) bool {
	return a.Venue == b.Venue && a.VenueSymbol == b.VenueSymbol && a.CanonicalSymbol == b.CanonicalSymbol
}

func averageZoneDelta(inputs []MultipleHVNInput, members []int) float64 {
	if len(members) == 0 {
		return 0
	}
	var total float64
	for _, idx := range members {
		total += inputs[idx].Delta
	}
	return total / float64(len(members))
}

func sumZoneImbalances(inputs []MultipleHVNInput, members []int) int {
	total := 0
	for _, idx := range members {
		total += inputs[idx].ImbalanceCount
	}
	return total
}

func sumZoneStackedImbalances(inputs []MultipleHVNInput, members []int) int {
	total := 0
	for _, idx := range members {
		total += inputs[idx].StackedImbalanceCount
	}
	return total
}
