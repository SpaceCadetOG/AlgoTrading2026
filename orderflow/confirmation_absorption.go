package orderflow

import "sort"

type AbsorptionConfirmation struct {
	Venue           string
	VenueSymbol     string
	CanonicalSymbol string

	Timestamp int64

	Open  float64
	High  float64
	Low   float64
	Close float64

	Delta  float64
	Volume float64

	PriceResponse float64
	BarRange      float64
	CloseLocation float64

	BullishAbsorption bool
	BearishAbsorption bool

	Accepted bool
	Rejected bool
	Neutral  bool

	FollowThrough5  float64
	FollowThrough10 float64
	FollowThrough20 float64

	HVNPrice  float64
	HVNVolume float64

	ImbalanceCount        int
	StackedImbalanceCount int

	NearHVN           bool
	NearVolumeCluster bool
	NearMultipleHVN   bool
}

type AbsorptionInput struct {
	Venue           string
	VenueSymbol     string
	CanonicalSymbol string

	Timestamp int64

	Open  float64
	High  float64
	Low   float64
	Close float64

	Delta  float64
	Volume float64

	HVNPrice           float64
	HVNVolume          float64
	VolumeClusterCount int

	NearMultipleHVN bool

	ImbalanceCount        int
	StackedImbalanceCount int
}

type AbsorptionConfig struct {
	DeltaThreshold        float64
	Percentile            float64
	CloseUpperHalf        float64
	CloseLowerHalf        float64
	MeaningfulMoveEpsilon float64
}

const absorptionFallbackDeltaThreshold = 1.0

func DefaultAbsorptionConfig() AbsorptionConfig {
	return AbsorptionConfig{
		DeltaThreshold:        0,
		Percentile:            0.60,
		CloseUpperHalf:        0.50,
		CloseLowerHalf:        0.50,
		MeaningfulMoveEpsilon: 0,
	}
}

func DetectAbsorptionConfirmations(inputs []AbsorptionInput, config AbsorptionConfig) []AbsorptionConfirmation {
	config = normalizeAbsorptionConfig(inputs, config)
	out := make([]AbsorptionConfirmation, 0)
	for i, input := range inputs {
		barRange := input.High - input.Low
		closeLocation := absorptionCloseLocation(input)
		priceResponse := input.Close - input.Open
		weakDownside := input.Close >= input.Open-config.MeaningfulMoveEpsilon || closeLocation >= config.CloseUpperHalf
		weakUpside := input.Close <= input.Open+config.MeaningfulMoveEpsilon || closeLocation <= config.CloseLowerHalf
		bullish := input.Delta <= -config.DeltaThreshold && weakDownside
		bearish := input.Delta >= config.DeltaThreshold && weakUpside
		if !bullish && !bearish {
			continue
		}
		confirmation := AbsorptionConfirmation{
			Venue:                 input.Venue,
			VenueSymbol:           input.VenueSymbol,
			CanonicalSymbol:       input.CanonicalSymbol,
			Timestamp:             input.Timestamp,
			Open:                  input.Open,
			High:                  input.High,
			Low:                   input.Low,
			Close:                 input.Close,
			Delta:                 input.Delta,
			Volume:                input.Volume,
			PriceResponse:         priceResponse,
			BarRange:              barRange,
			CloseLocation:         closeLocation,
			BullishAbsorption:     bullish,
			BearishAbsorption:     bearish,
			HVNPrice:              input.HVNPrice,
			HVNVolume:             input.HVNVolume,
			ImbalanceCount:        input.ImbalanceCount,
			StackedImbalanceCount: input.StackedImbalanceCount,
			NearHVN:               input.HVNPrice > 0,
			NearVolumeCluster:     input.VolumeClusterCount > 0,
			NearMultipleHVN:       input.NearMultipleHVN,
		}
		next := nextSameAbsorptionMarket(inputs, i)
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
		confirmation.FollowThrough5 = absorptionFollowThrough(inputs, i, 5, bullish, bearish)
		confirmation.FollowThrough10 = absorptionFollowThrough(inputs, i, 10, bullish, bearish)
		confirmation.FollowThrough20 = absorptionFollowThrough(inputs, i, 20, bullish, bearish)
		out = append(out, confirmation)
	}
	return out
}

func normalizeAbsorptionConfig(inputs []AbsorptionInput, config AbsorptionConfig) AbsorptionConfig {
	def := DefaultAbsorptionConfig()
	if config.Percentile <= 0 || config.Percentile > 1 {
		config.Percentile = def.Percentile
	}
	if config.CloseUpperHalf <= 0 {
		config.CloseUpperHalf = def.CloseUpperHalf
	}
	if config.CloseLowerHalf <= 0 {
		config.CloseLowerHalf = def.CloseLowerHalf
	}
	if config.DeltaThreshold <= 0 {
		config.DeltaThreshold = absorptionDeltaPercentile(inputs, config.Percentile)
		if config.DeltaThreshold <= 0 {
			config.DeltaThreshold = absorptionFallbackDeltaThreshold
		}
	}
	return config
}

func absorptionDeltaPercentile(inputs []AbsorptionInput, percentile float64) float64 {
	if len(inputs) == 0 {
		return 0
	}
	values := make([]float64, 0, len(inputs))
	for _, input := range inputs {
		value := input.Delta
		if value < 0 {
			value = -value
		}
		values = append(values, value)
	}
	sort.Float64s(values)
	idx := int(float64(len(values)-1) * percentile)
	if idx < 0 {
		idx = 0
	}
	if idx >= len(values) {
		idx = len(values) - 1
	}
	return values[idx]
}

func absorptionCloseLocation(input AbsorptionInput) float64 {
	if input.High <= input.Low {
		return 0.5
	}
	return (input.Close - input.Low) / (input.High - input.Low)
}

func nextSameAbsorptionMarket(inputs []AbsorptionInput, start int) int {
	for i := start + 1; i < len(inputs); i++ {
		if sameAbsorptionMarket(inputs[start], inputs[i]) {
			return i
		}
	}
	return -1
}

func absorptionFollowThrough(inputs []AbsorptionInput, start int, horizon int, bullish bool, bearish bool) float64 {
	count := 0
	for i := start + 1; i < len(inputs); i++ {
		if !sameAbsorptionMarket(inputs[start], inputs[i]) {
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

func sameAbsorptionMarket(a AbsorptionInput, b AbsorptionInput) bool {
	return a.Venue == b.Venue && a.VenueSymbol == b.VenueSymbol && a.CanonicalSymbol == b.CanonicalSymbol
}
