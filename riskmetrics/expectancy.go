package riskmetrics

import "math"

func Expectancy(winRate float64, avgWin float64, avgLoss float64) float64 {
	return winRate*avgWin - (1-winRate)*math.Abs(avgLoss)
}
