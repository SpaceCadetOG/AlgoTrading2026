package indicators

type MACDResult struct {
	MACD      []float64
	Signal    []float64
	Histogram []float64
}

func MACD(values []float64, fastPeriod int, slowPeriod int, signalPeriod int) MACDResult {
	macd := APO(values, fastPeriod, slowPeriod)
	signal := EMA(macd, signalPeriod)
	histogram := make([]float64, len(values))
	for i := range values {
		histogram[i] = macd[i] - signal[i]
	}

	return MACDResult{
		MACD:      macd,
		Signal:    signal,
		Histogram: histogram,
	}
}
