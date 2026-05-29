package backtest

type EquityCurve struct {
	Points []PortfolioPoint
}

func (e *EquityCurve) Append(point PortfolioPoint) {
	e.Points = append(e.Points, point)
}

func (e EquityCurve) Len() int {
	return len(e.Points)
}

func (e EquityCurve) EndingEquity(fallback float64) float64 {
	if len(e.Points) == 0 {
		return fallback
	}
	return e.Points[len(e.Points)-1].TotalEquity
}

func (e EquityCurve) Drawdown(startingEquity float64) DrawdownTracker {
	tracker := NewDrawdownTracker(startingEquity)
	for _, point := range e.Points {
		tracker.Update(point.TotalEquity)
	}
	return tracker
}
