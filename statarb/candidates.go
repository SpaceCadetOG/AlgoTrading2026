package statarb

import (
	"math"

	"AlgoTrading2026/exchanges"
	"AlgoTrading2026/pairs"
)

type Candidate struct {
	VenueA  string
	SymbolA string
	VenueB  string
	SymbolB string

	Samples int

	Correlation float64
	HedgeRatio  float64

	LatestSpread float64
	LatestZScore float64

	AbsZScore float64

	Score float64

	Notes []string
}

func BuildCandidate(
	venueA string,
	symbolA string,
	candlesA []exchanges.Candle,
	venueB string,
	symbolB string,
	candlesB []exchanges.Candle,
	zPeriod int,
) (Candidate, error) {
	aligned, err := pairs.AlignByTime(candlesA, candlesB)
	if err != nil {
		return Candidate{}, err
	}

	x := make([]float64, len(aligned))
	y := make([]float64, len(aligned))
	for i, point := range aligned {
		x[i] = point.AClose
		y[i] = point.BClose
	}

	corr := pairs.Pearson(pairs.Returns(x), pairs.Returns(y))
	beta := HedgeRatio(x, y)
	spread := NormalizedSpread(x, y, beta)
	zscores := RollingZScore(spread, zPeriod)
	last := len(aligned) - 1
	absZ := math.Abs(zscores[last])

	candidate := Candidate{
		VenueA:       venueA,
		SymbolA:      symbolA,
		VenueB:       venueB,
		SymbolB:      symbolB,
		Samples:      len(aligned),
		Correlation:  corr,
		HedgeRatio:   beta,
		LatestSpread: spread[last],
		LatestZScore: zscores[last],
		AbsZScore:    absZ,
		Score:        math.Abs(corr) * absZ,
	}
	candidate.Notes = candidateNotes(candidate)
	return candidate, nil
}

func candidateNotes(candidate Candidate) []string {
	notes := make([]string, 0, 4)
	if candidate.Samples < 50 {
		notes = append(notes, "insufficient_samples")
	}
	if math.Abs(candidate.Correlation) < 0.5 {
		notes = append(notes, "weak_correlation")
	}
	if candidate.AbsZScore >= 3 {
		notes = append(notes, "extreme_divergence")
	} else if candidate.AbsZScore >= 2 {
		notes = append(notes, "interesting_divergence")
	}
	return notes
}
