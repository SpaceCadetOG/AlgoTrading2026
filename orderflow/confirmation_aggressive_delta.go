package orderflow

import "sort"

type AggressiveDeltaConfirmation struct {
	Venue           string
	VenueSymbol     string
	CanonicalSymbol string
	Timestamp       int64
	Delta           float64
	AbsoluteDelta   float64
	Volume          float64
	Open            float64
	High            float64
	Low             float64
	Close           float64

	AggressiveBuy       bool
	AggressiveSell      bool
	DeltaStrengthBucket string

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

type AggressiveDeltaInput struct {
	Venue           string
	VenueSymbol     string
	CanonicalSymbol string
	Timestamp       int64
	Delta           float64
	Volume          float64
	Open            float64
	High            float64
	Low             float64
	Close           float64

	HVNPrice           float64
	VolumeClusterCount int
	NearMultipleHVN    bool

	ImbalanceCount        int
	StackedImbalanceCount int
}

type AggressiveDeltaConfig struct {
	DeltaThreshold float64
	Percentile     float64
}

func DefaultAggressiveDeltaConfig() AggressiveDeltaConfig {
	return AggressiveDeltaConfig{Percentile: 0.70}
}

func DetectAggressiveDeltaConfirmations(inputs []AggressiveDeltaInput, config AggressiveDeltaConfig) []AggressiveDeltaConfirmation {
	threshold := config.DeltaThreshold
	if threshold <= 0 {
		percentile := config.Percentile
		if percentile <= 0 || percentile > 1 {
			percentile = DefaultAggressiveDeltaConfig().Percentile
		}
		threshold = aggressiveDeltaPercentile(inputs, percentile)
		if threshold <= 0 {
			threshold = 1
		}
	}
	out := make([]AggressiveDeltaConfirmation, 0)
	for i, input := range inputs {
		absDelta := absFloat(input.Delta)
		buy := input.Delta >= threshold
		sell := input.Delta <= -threshold
		if !buy && !sell {
			continue
		}
		confirmation := AggressiveDeltaConfirmation{
			Venue:                 input.Venue,
			VenueSymbol:           input.VenueSymbol,
			CanonicalSymbol:       input.CanonicalSymbol,
			Timestamp:             input.Timestamp,
			Delta:                 input.Delta,
			AbsoluteDelta:         absDelta,
			Volume:                input.Volume,
			Open:                  input.Open,
			High:                  input.High,
			Low:                   input.Low,
			Close:                 input.Close,
			AggressiveBuy:         buy,
			AggressiveSell:        sell,
			DeltaStrengthBucket:   deltaStrengthBucket(absDelta, threshold),
			NearHVN:               input.HVNPrice > 0,
			NearVolumeCluster:     input.VolumeClusterCount > 0,
			NearMultipleHVN:       input.NearMultipleHVN,
			ImbalanceCount:        input.ImbalanceCount,
			StackedImbalanceCount: input.StackedImbalanceCount,
		}
		next := nextSameAggressiveDeltaMarket(inputs, i)
		if next < 0 {
			confirmation.Neutral = true
		} else {
			future := inputs[next]
			if buy {
				confirmation.Accepted = future.Close > input.Close
				confirmation.Rejected = future.Close < input.Close
				confirmation.Neutral = future.Close == input.Close
			}
			if sell {
				confirmation.Accepted = future.Close < input.Close
				confirmation.Rejected = future.Close > input.Close
				confirmation.Neutral = future.Close == input.Close
			}
		}
		confirmation.FollowThrough5 = aggressiveDeltaFollowThrough(inputs, i, 5, buy, sell)
		confirmation.FollowThrough10 = aggressiveDeltaFollowThrough(inputs, i, 10, buy, sell)
		confirmation.FollowThrough20 = aggressiveDeltaFollowThrough(inputs, i, 20, buy, sell)
		out = append(out, confirmation)
	}
	return out
}

func aggressiveDeltaPercentile(inputs []AggressiveDeltaInput, percentile float64) float64 {
	if len(inputs) == 0 {
		return 0
	}
	values := make([]float64, 0, len(inputs))
	for _, input := range inputs {
		values = append(values, absFloat(input.Delta))
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

func deltaStrengthBucket(absDelta float64, threshold float64) string {
	switch {
	case absDelta >= threshold*3:
		return "extreme"
	case absDelta >= threshold*2:
		return "high"
	case absDelta >= threshold*1.25:
		return "medium"
	default:
		return "low"
	}
}

func nextSameAggressiveDeltaMarket(inputs []AggressiveDeltaInput, start int) int {
	for i := start + 1; i < len(inputs); i++ {
		if sameAggressiveDeltaMarket(inputs[start], inputs[i]) {
			return i
		}
	}
	return -1
}

func aggressiveDeltaFollowThrough(inputs []AggressiveDeltaInput, start int, horizon int, buy bool, sell bool) float64 {
	count := 0
	for i := start + 1; i < len(inputs); i++ {
		if !sameAggressiveDeltaMarket(inputs[start], inputs[i]) {
			continue
		}
		count++
		if count == horizon {
			if buy {
				return inputs[i].Close - inputs[start].Close
			}
			if sell {
				return inputs[start].Close - inputs[i].Close
			}
		}
	}
	return 0
}

func sameAggressiveDeltaMarket(a AggressiveDeltaInput, b AggressiveDeltaInput) bool {
	return a.Venue == b.Venue && a.VenueSymbol == b.VenueSymbol && a.CanonicalSymbol == b.CanonicalSymbol
}

func absFloat(value float64) float64 {
	if value < 0 {
		return -value
	}
	return value
}
