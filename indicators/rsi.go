package indicators

func RSI(values []float64, period int) []float64 {
	out := make([]float64, len(values))
	if period <= 0 || len(values) == 0 {
		return out
	}

	for i := range values {
		if i < period {
			continue
		}

		gain := 0.0
		loss := 0.0
		for j := i - period + 1; j <= i; j++ {
			change := values[j] - values[j-1]
			if change > 0 {
				gain += change
			} else {
				loss -= change
			}
		}

		avgGain := gain / float64(period)
		avgLoss := loss / float64(period)
		switch {
		case avgLoss == 0 && avgGain == 0:
			out[i] = 50
		case avgLoss == 0:
			out[i] = 100
		default:
			rs := avgGain / avgLoss
			out[i] = 100 - 100/(1+rs)
		}
	}
	return out
}
