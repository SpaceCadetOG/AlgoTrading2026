package riskmetrics

type ExecutionStats struct {
	TradesPerDay   float64
	TradesPerWeek  float64
	TradesPerMonth float64
}

func CalculateExecutionStats(trades int, days float64) ExecutionStats {
	if days <= 0 {
		return ExecutionStats{}
	}

	perDay := float64(trades) / days
	return ExecutionStats{
		TradesPerDay:   perDay,
		TradesPerWeek:  perDay * 7,
		TradesPerMonth: perDay * 30,
	}
}
