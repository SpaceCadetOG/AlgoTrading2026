package volatility

import "math"

func LogReturns(close []float64) []float64 {
	out := make([]float64, len(close))
	for i := 1; i < len(close); i++ {
		if close[i-1] <= 0 || close[i] <= 0 {
			continue
		}
		out[i] = math.Log(close[i] / close[i-1])
	}
	return out
}

func RealizedVolatility(close []float64, period int) []float64 {
	return RollingStdDev(LogReturns(close), period)
}
