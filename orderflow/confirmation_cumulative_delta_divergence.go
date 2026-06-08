package orderflow

type CumulativeDeltaDivergenceConfirmation struct {
	Venue           string
	VenueSymbol     string
	CanonicalSymbol string
	Timestamp       int64

	PriceHigh            float64
	PriceLow             float64
	PriceClose           float64
	BarDelta             float64
	CumulativeDelta      float64
	PriorPriceHigh       float64
	PriorPriceLow        float64
	PriorCumulativeDelta float64

	BullishDivergence bool
	BearishDivergence bool

	Accepted bool
	Rejected bool
	Neutral  bool

	FollowThrough5  float64
	FollowThrough10 float64
	FollowThrough20 float64

	NearHVN           bool
	NearVolumeCluster bool
	NearMultipleHVN   bool

	ImbalanceCount        int
	StackedImbalanceCount int
}

type CumulativeDeltaDivergenceInput struct {
	Venue           string
	VenueSymbol     string
	CanonicalSymbol string
	Timestamp       int64

	High  float64
	Low   float64
	Close float64
	Delta float64

	HVNPrice           float64
	VolumeClusterCount int
	NearMultipleHVN    bool

	ImbalanceCount        int
	StackedImbalanceCount int
}

type CumulativeDeltaDivergenceConfig struct {
	Window int
}

func DefaultCumulativeDeltaDivergenceConfig() CumulativeDeltaDivergenceConfig {
	return CumulativeDeltaDivergenceConfig{Window: 5}
}

func DetectCumulativeDeltaDivergences(inputs []CumulativeDeltaDivergenceInput, config CumulativeDeltaDivergenceConfig) []CumulativeDeltaDivergenceConfirmation {
	if config.Window <= 0 {
		config = DefaultCumulativeDeltaDivergenceConfig()
	}
	cumulative := make([]float64, len(inputs))
	runningByMarket := map[string]float64{}
	for i, input := range inputs {
		key := cumulativeDeltaMarketKey(input)
		runningByMarket[key] += input.Delta
		cumulative[i] = runningByMarket[key]
	}
	out := make([]CumulativeDeltaDivergenceConfirmation, 0)
	for i, input := range inputs {
		priorIdx := nthPriorSameCumulativeMarket(inputs, i, config.Window)
		if priorIdx < 0 {
			continue
		}
		prior := inputs[priorIdx]
		bullish := input.Low < prior.Low && cumulative[i] > cumulative[priorIdx]
		bearish := input.High > prior.High && cumulative[i] < cumulative[priorIdx]
		if !bullish && !bearish {
			continue
		}
		confirmation := CumulativeDeltaDivergenceConfirmation{
			Venue:                 input.Venue,
			VenueSymbol:           input.VenueSymbol,
			CanonicalSymbol:       input.CanonicalSymbol,
			Timestamp:             input.Timestamp,
			PriceHigh:             input.High,
			PriceLow:              input.Low,
			PriceClose:            input.Close,
			BarDelta:              input.Delta,
			CumulativeDelta:       cumulative[i],
			PriorPriceHigh:        prior.High,
			PriorPriceLow:         prior.Low,
			PriorCumulativeDelta:  cumulative[priorIdx],
			BullishDivergence:     bullish,
			BearishDivergence:     bearish,
			NearHVN:               input.HVNPrice > 0,
			NearVolumeCluster:     input.VolumeClusterCount > 0,
			NearMultipleHVN:       input.NearMultipleHVN,
			ImbalanceCount:        input.ImbalanceCount,
			StackedImbalanceCount: input.StackedImbalanceCount,
		}
		next := nextSameCumulativeMarket(inputs, i)
		if next < 0 {
			confirmation.Neutral = true
		} else {
			future := inputs[next]
			if bullish {
				confirmation.Accepted = future.Close > input.Close
				confirmation.Rejected = future.Close < input.Close
				confirmation.Neutral = future.Close == input.Close
			}
			if bearish {
				confirmation.Accepted = future.Close < input.Close
				confirmation.Rejected = future.Close > input.Close
				confirmation.Neutral = future.Close == input.Close
			}
		}
		confirmation.FollowThrough5 = cumulativeDeltaFollowThrough(inputs, i, 5, bullish, bearish)
		confirmation.FollowThrough10 = cumulativeDeltaFollowThrough(inputs, i, 10, bullish, bearish)
		confirmation.FollowThrough20 = cumulativeDeltaFollowThrough(inputs, i, 20, bullish, bearish)
		out = append(out, confirmation)
	}
	return out
}

func nthPriorSameCumulativeMarket(inputs []CumulativeDeltaDivergenceInput, start int, window int) int {
	count := 0
	for i := start - 1; i >= 0; i-- {
		if sameCumulativeMarket(inputs[start], inputs[i]) {
			count++
			if count == window {
				return i
			}
		}
	}
	return -1
}

func nextSameCumulativeMarket(inputs []CumulativeDeltaDivergenceInput, start int) int {
	for i := start + 1; i < len(inputs); i++ {
		if sameCumulativeMarket(inputs[start], inputs[i]) {
			return i
		}
	}
	return -1
}

func cumulativeDeltaFollowThrough(inputs []CumulativeDeltaDivergenceInput, start int, horizon int, bullish bool, bearish bool) float64 {
	count := 0
	for i := start + 1; i < len(inputs); i++ {
		if !sameCumulativeMarket(inputs[start], inputs[i]) {
			continue
		}
		count++
		if count == horizon {
			if bullish {
				return inputs[i].Close - inputs[start].Close
			}
			if bearish {
				return inputs[start].Close - inputs[i].Close
			}
		}
	}
	return 0
}

func sameCumulativeMarket(a CumulativeDeltaDivergenceInput, b CumulativeDeltaDivergenceInput) bool {
	return a.Venue == b.Venue && a.VenueSymbol == b.VenueSymbol && a.CanonicalSymbol == b.CanonicalSymbol
}

func cumulativeDeltaMarketKey(input CumulativeDeltaDivergenceInput) string {
	return input.Venue + "|" + input.VenueSymbol + "|" + input.CanonicalSymbol
}
