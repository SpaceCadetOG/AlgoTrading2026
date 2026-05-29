package labels

type TPSLConfig struct {
	Horizon       int
	TakeProfitPct float64
	StopLossPct   float64
}

// TPSLLabel uses a close-only path approximation. It does not inspect intrabar high/low,
// so it cannot identify take-profit or stop-loss hits that occur inside a candle.
func TPSLLabel(close []float64, cfg TPSLConfig) []int {
	out := make([]int, len(close))
	if cfg.Horizon <= 0 {
		return out
	}

	for i := range close {
		entry := close[i]
		if entry == 0 {
			continue
		}
		tp := entry * (1 + cfg.TakeProfitPct)
		sl := entry * (1 - cfg.StopLossPct)
		end := i + cfg.Horizon
		if end >= len(close) {
			end = len(close) - 1
		}
		for j := i + 1; j <= end; j++ {
			switch {
			case close[j] >= tp:
				out[i] = 1
				j = end + 1
			case close[j] <= sl:
				out[i] = -1
				j = end + 1
			}
		}
	}
	return out
}
