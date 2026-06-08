package orderflow

type UnfinishedBusinessSetup struct {
	Venue           string
	VenueSymbol     string
	CanonicalSymbol string

	Timestamp int64

	Level    float64
	Location string

	UnfinishedHigh bool
	UnfinishedLow  bool

	MagnetContext bool
	LongContext   bool
	ShortContext  bool

	Revisited bool
	Accepted  bool
	Rejected  bool
	Neutral   bool

	FollowThrough5  float64
	FollowThrough10 float64
	FollowThrough20 float64

	BarDelta  float64
	BarVolume float64

	HVNPrice  float64
	HVNVolume float64

	ImbalanceCount        int
	StackedImbalanceCount int
}

type UnfinishedBusinessInput struct {
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

	ImbalanceCount        int
	StackedImbalanceCount int
}

type UnfinishedBusinessConfig struct {
	RevisitTolerancePct float64
}

func DefaultUnfinishedBusinessConfig() UnfinishedBusinessConfig {
	return UnfinishedBusinessConfig{RevisitTolerancePct: 0.0005}
}

func DetectUnfinishedBusinessSetups(inputs []UnfinishedBusinessInput, config UnfinishedBusinessConfig) []UnfinishedBusinessSetup {
	if config.RevisitTolerancePct <= 0 {
		config.RevisitTolerancePct = DefaultUnfinishedBusinessConfig().RevisitTolerancePct
	}
	setups := make([]UnfinishedBusinessSetup, 0)
	for i, input := range inputs {
		bar := FootprintBar{
			Venue: input.Venue, VenueSymbol: input.VenueSymbol, CanonicalSymbol: input.CanonicalSymbol,
			StartTimestamp: input.StartTimestamp, EndTimestamp: input.EndTimestamp,
			High: input.High, Low: input.Low, Close: input.Close, Levels: input.Levels,
		}
		unfinished := DetectUnfinishedBusiness(bar)
		if unfinished.High {
			setup := unfinishedBusinessBase(input, "HIGH", input.High)
			setup.UnfinishedHigh = true
			setup.ShortContext = true
			applyUnfinishedBusinessOutcome(inputs, i, &setup, config)
			setups = append(setups, setup)
		}
		if unfinished.Low {
			setup := unfinishedBusinessBase(input, "LOW", input.Low)
			setup.UnfinishedLow = true
			setup.LongContext = true
			applyUnfinishedBusinessOutcome(inputs, i, &setup, config)
			setups = append(setups, setup)
		}
	}
	return setups
}

func unfinishedBusinessBase(input UnfinishedBusinessInput, location string, level float64) UnfinishedBusinessSetup {
	return UnfinishedBusinessSetup{
		Venue:                 input.Venue,
		VenueSymbol:           input.VenueSymbol,
		CanonicalSymbol:       input.CanonicalSymbol,
		Timestamp:             input.StartTimestamp,
		Level:                 level,
		Location:              location,
		MagnetContext:         true,
		BarDelta:              input.BarDelta,
		BarVolume:             input.BarVolume,
		HVNPrice:              input.HVNPrice,
		HVNVolume:             input.HVNVolume,
		ImbalanceCount:        input.ImbalanceCount,
		StackedImbalanceCount: input.StackedImbalanceCount,
	}
}

func applyUnfinishedBusinessOutcome(inputs []UnfinishedBusinessInput, start int, setup *UnfinishedBusinessSetup, config UnfinishedBusinessConfig) {
	revisitIndex := firstUnfinishedBusinessRevisit(inputs, start, *setup, config.RevisitTolerancePct)
	setup.Revisited = revisitIndex >= 0
	if revisitIndex >= 0 {
		outcome := inputs[revisitIndex]
		switch setup.Location {
		case "HIGH":
			setup.Accepted = outcome.Close > setup.Level
			setup.Rejected = outcome.Close < setup.Level
		case "LOW":
			setup.Accepted = outcome.Close < setup.Level
			setup.Rejected = outcome.Close > setup.Level
		}
	}
	setup.Neutral = !setup.Accepted && !setup.Rejected
	setup.FollowThrough5 = unfinishedBusinessFollowThrough(inputs, start, 5, *setup)
	setup.FollowThrough10 = unfinishedBusinessFollowThrough(inputs, start, 10, *setup)
	setup.FollowThrough20 = unfinishedBusinessFollowThrough(inputs, start, 20, *setup)
}

func firstUnfinishedBusinessRevisit(inputs []UnfinishedBusinessInput, start int, setup UnfinishedBusinessSetup, tolerancePct float64) int {
	if setup.Level <= 0 {
		return -1
	}
	buffer := setup.Level * tolerancePct
	for i := start + 1; i < len(inputs); i++ {
		if !sameUnfinishedBusinessMarket(inputs[start], inputs[i]) {
			continue
		}
		if inputs[i].Low <= setup.Level+buffer && inputs[i].High >= setup.Level-buffer {
			return i
		}
	}
	return -1
}

func unfinishedBusinessFollowThrough(inputs []UnfinishedBusinessInput, start int, horizon int, setup UnfinishedBusinessSetup) float64 {
	count := 0
	for i := start + 1; i < len(inputs); i++ {
		if !sameUnfinishedBusinessMarket(inputs[start], inputs[i]) {
			continue
		}
		count++
		if count == horizon {
			if setup.ShortContext {
				return setup.Level - inputs[i].Close
			}
			if setup.LongContext {
				return inputs[i].Close - setup.Level
			}
			return inputs[i].Close - setup.Level
		}
	}
	return 0
}

func sameUnfinishedBusinessMarket(a UnfinishedBusinessInput, b UnfinishedBusinessInput) bool {
	return a.Venue == b.Venue && a.VenueSymbol == b.VenueSymbol && a.CanonicalSymbol == b.CanonicalSymbol
}
