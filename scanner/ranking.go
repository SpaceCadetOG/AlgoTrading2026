package scanner

import "sort"

func ScoreInstrument(instrument Instrument) float64 {
	score := 0.0
	if instrument.Volume24h > 0 {
		score += 20
	}
	if instrument.OpenInterest > 0 {
		score += 15
	}
	if instrument.SpreadPct > 0 {
		switch {
		case instrument.SpreadPct < 0.001:
			score += 25
		case instrument.SpreadPct < 0.01:
			score += 15
		default:
			score += 5
		}
	}
	if instrument.L2Available {
		score += 20
	}
	if instrument.CandlesAvailable {
		score += 20
	}
	if instrument.TradesAvailable {
		score += 10
	}
	if instrument.ProfileReady {
		score += 20
	}
	if instrument.VolatilityProxy > 0 && instrument.VolatilityProxy < 0.05 {
		score += 10
	}
	return score
}

func RankInstruments(instruments []Instrument) []Instrument {
	out := append([]Instrument(nil), instruments...)
	for i := range out {
		if out[i].AssetType == "" {
			out[i].AssetType = ClassifyAsset(out[i].Symbol)
		}
		out[i].ResearchScore = ScoreInstrument(out[i])
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].ResearchScore == out[j].ResearchScore {
			if out[i].Venue == out[j].Venue {
				return out[i].Symbol < out[j].Symbol
			}
			return out[i].Venue < out[j].Venue
		}
		return out[i].ResearchScore > out[j].ResearchScore
	})
	for i := range out {
		out[i].Rank = i + 1
	}
	return out
}
