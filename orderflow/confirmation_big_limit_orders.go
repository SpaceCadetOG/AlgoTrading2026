package orderflow

type BigLimitOrderConfirmation struct {
	Venue           string
	VenueSymbol     string
	CanonicalSymbol string

	Timestamp int64

	Delta  float64
	Volume float64

	AggressiveBuyVolume  float64
	AggressiveSellVolume float64

	PriceResponse float64

	BullishAbsorption bool
	BearishAbsorption bool

	Accepted bool
	Rejected bool

	FollowThrough5  float64
	FollowThrough10 float64
	FollowThrough20 float64

	NearHVN           bool
	NearVolumeCluster bool
	NearMultipleHVN   bool

	ImbalanceCount        int
	StackedImbalanceCount int
}

type BigLimitOrderInput struct {
	Venue           string
	VenueSymbol     string
	CanonicalSymbol string

	Timestamp int64
	High      float64
	Low       float64
	Close     float64

	Delta  float64
	Volume float64

	AggressiveBuyVolume  float64
	AggressiveSellVolume float64

	NearHVN           bool
	NearVolumeCluster bool
	NearMultipleHVN   bool

	ImbalanceCount        int
	StackedImbalanceCount int
}

type BigLimitOrderConfig struct {
	MinAbsDelta       float64
	CloseNearHighPct  float64
	CloseNearLowPct   float64
	WeakMoveThreshold float64
}

func DefaultBigLimitOrderConfig() BigLimitOrderConfig {
	return BigLimitOrderConfig{
		MinAbsDelta:       1,
		CloseNearHighPct:  0.75,
		CloseNearLowPct:   0.25,
		WeakMoveThreshold: 0,
	}
}

func DetectBigLimitOrderConfirmations(inputs []BigLimitOrderInput, config BigLimitOrderConfig) []BigLimitOrderConfirmation {
	if config.MinAbsDelta <= 0 {
		config.MinAbsDelta = DefaultBigLimitOrderConfig().MinAbsDelta
	}
	if config.CloseNearHighPct <= 0 {
		config.CloseNearHighPct = DefaultBigLimitOrderConfig().CloseNearHighPct
	}
	if config.CloseNearLowPct <= 0 {
		config.CloseNearLowPct = DefaultBigLimitOrderConfig().CloseNearLowPct
	}
	out := make([]BigLimitOrderConfirmation, 0)
	for i, input := range inputs {
		priceResponse := absorptionPriceResponse(input)
		bullish := input.Delta <= -config.MinAbsDelta && priceResponse >= config.CloseNearHighPct
		bearish := input.Delta >= config.MinAbsDelta && priceResponse <= config.CloseNearLowPct
		if !bullish && !bearish {
			continue
		}
		confirmation := BigLimitOrderConfirmation{
			Venue:                 input.Venue,
			VenueSymbol:           input.VenueSymbol,
			CanonicalSymbol:       input.CanonicalSymbol,
			Timestamp:             input.Timestamp,
			Delta:                 input.Delta,
			Volume:                input.Volume,
			AggressiveBuyVolume:   input.AggressiveBuyVolume,
			AggressiveSellVolume:  input.AggressiveSellVolume,
			PriceResponse:         priceResponse,
			BullishAbsorption:     bullish,
			BearishAbsorption:     bearish,
			NearHVN:               input.NearHVN,
			NearVolumeCluster:     input.NearVolumeCluster,
			NearMultipleHVN:       input.NearMultipleHVN,
			ImbalanceCount:        input.ImbalanceCount,
			StackedImbalanceCount: input.StackedImbalanceCount,
		}
		next := nextSameBigLimitMarket(inputs, i)
		if next >= 0 {
			future := inputs[next]
			if bullish {
				confirmation.Accepted = future.Close > input.Close
				confirmation.Rejected = future.Close <= input.Close
			}
			if bearish {
				confirmation.Accepted = future.Close < input.Close
				confirmation.Rejected = future.Close >= input.Close
			}
		}
		confirmation.FollowThrough5 = bigLimitFollowThrough(inputs, i, 5, bullish, bearish)
		confirmation.FollowThrough10 = bigLimitFollowThrough(inputs, i, 10, bullish, bearish)
		confirmation.FollowThrough20 = bigLimitFollowThrough(inputs, i, 20, bullish, bearish)
		out = append(out, confirmation)
	}
	return out
}

func absorptionPriceResponse(input BigLimitOrderInput) float64 {
	if input.High <= input.Low {
		return 0.5
	}
	return (input.Close - input.Low) / (input.High - input.Low)
}

func nextSameBigLimitMarket(inputs []BigLimitOrderInput, start int) int {
	for i := start + 1; i < len(inputs); i++ {
		if sameBigLimitMarket(inputs[start], inputs[i]) {
			return i
		}
	}
	return -1
}

func bigLimitFollowThrough(inputs []BigLimitOrderInput, start int, horizon int, bullish bool, bearish bool) float64 {
	count := 0
	for i := start + 1; i < len(inputs); i++ {
		if !sameBigLimitMarket(inputs[start], inputs[i]) {
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
			return inputs[i].Close - inputs[start].Close
		}
	}
	return 0
}

func sameBigLimitMarket(a BigLimitOrderInput, b BigLimitOrderInput) bool {
	return a.Venue == b.Venue && a.VenueSymbol == b.VenueSymbol && a.CanonicalSymbol == b.CanonicalSymbol
}
