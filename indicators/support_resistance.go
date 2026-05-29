package indicators

type SupportResistanceResult struct {
	Support    []float64
	Resistance []float64
}

func RollingSupportResistance(highs []float64, lows []float64, period int) SupportResistanceResult {
	n := min(len(highs), len(lows))
	support := make([]float64, n)
	resistance := make([]float64, n)
	if period <= 0 {
		return SupportResistanceResult{Support: support, Resistance: resistance}
	}

	for i := 0; i < n; i++ {
		start := i - period + 1
		if start < 0 {
			start = 0
		}
		low := lows[start]
		high := highs[start]
		for j := start + 1; j <= i; j++ {
			if lows[j] < low {
				low = lows[j]
			}
			if highs[j] > high {
				high = highs[j]
			}
		}
		support[i] = low
		resistance[i] = high
	}

	return SupportResistanceResult{Support: support, Resistance: resistance}
}
