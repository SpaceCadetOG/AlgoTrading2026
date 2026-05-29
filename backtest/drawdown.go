package backtest

type DrawdownTracker struct {
	PeakEquity         float64
	MaxDrawdown        float64
	MaxDrawdownPct     float64
	CurrentDrawdown    float64
	CurrentDrawdownPct float64
}

func NewDrawdownTracker(startingEquity float64) DrawdownTracker {
	return DrawdownTracker{PeakEquity: startingEquity}
}

func (d *DrawdownTracker) Update(equity float64) {
	if d.PeakEquity == 0 || equity > d.PeakEquity {
		d.PeakEquity = equity
	}

	d.CurrentDrawdown = d.PeakEquity - equity
	if d.PeakEquity > 0 {
		d.CurrentDrawdownPct = d.CurrentDrawdown / d.PeakEquity * 100
	}

	if d.CurrentDrawdown > d.MaxDrawdown {
		d.MaxDrawdown = d.CurrentDrawdown
		d.MaxDrawdownPct = d.CurrentDrawdownPct
	}
}
