package pairs

import "AlgoTrading2026/exchanges"

type PairCandidate struct {
	SymbolA      string
	SymbolB      string
	VenueA       string
	VenueB       string
	Correlation  float64
	LatestSpread float64
	LatestZScore float64
	Samples      int
}

func BuildPairCandidate(
	venueA string,
	symbolA string,
	candlesA []exchanges.Candle,
	venueB string,
	symbolB string,
	candlesB []exchanges.Candle,
	zPeriod int,
) (PairCandidate, error) {
	aligned, err := AlignByTime(candlesA, candlesB)
	if err != nil {
		return PairCandidate{}, err
	}

	a := make([]float64, len(aligned))
	b := make([]float64, len(aligned))
	for i, point := range aligned {
		a[i] = point.AClose
		b[i] = point.BClose
	}

	aReturns := Returns(a)
	bReturns := Returns(b)
	spread := Spread(a, b)
	zscore := ZScore(spread, zPeriod)

	last := len(aligned) - 1
	return PairCandidate{
		VenueA:       venueA,
		SymbolA:      symbolA,
		VenueB:       venueB,
		SymbolB:      symbolB,
		Correlation:  Pearson(aReturns, bReturns),
		LatestSpread: spread[last],
		LatestZScore: zscore[last],
		Samples:      len(aligned),
	}, nil
}
