package labels

type LabelRow struct {
	Time         int64
	FutureReturn float64
	Direction    int
	TPSL         int
}

func FutureReturn(close []float64, horizon int) []float64 {
	out := make([]float64, len(close))
	if horizon <= 0 {
		return out
	}

	for i := 0; i+horizon < len(close); i++ {
		if close[i] == 0 {
			continue
		}
		out[i] = (close[i+horizon] - close[i]) / close[i]
	}
	return out
}
