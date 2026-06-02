package priceaction

import "AlgoTrading2026/exchanges"

func ABCDSetups(candles []exchanges.Candle) []StrategySetup {
	out := directABCDSetups(candles)
	pivots := swingPivots(candles)
	for i := 0; i+3 < len(pivots); i++ {
		a, b, c, d := pivots[i], pivots[i+1], pivots[i+2], pivots[i+3]
		switch {
		case a.Kind == "high" && b.Kind == "low" && c.Kind == "high" && d.Kind == "low":
			ab := a.Price - b.Price
			if ab <= 0 || c.Price >= a.Price {
				continue
			}
			retrace := (c.Price - b.Price) / ab
			projectedD := c.Price - ab
			if retrace >= 0.50 && nearLevel(d.Price, projectedD) {
				out = append(out, setupAt(candles, d.Index, "abcd", StudyLong, projectedD, projectedD, true, invalidatedByClose(candles[d.Index], StudyLong, projectedD), "bullish AB=CD projection near D"))
			}
		case a.Kind == "low" && b.Kind == "high" && c.Kind == "low" && d.Kind == "high":
			ab := b.Price - a.Price
			if ab <= 0 || c.Price <= a.Price {
				continue
			}
			retrace := (b.Price - c.Price) / ab
			projectedD := c.Price + ab
			if retrace >= 0.50 && nearLevel(d.Price, projectedD) {
				out = append(out, setupAt(candles, d.Index, "abcd", StudyShort, projectedD, projectedD, true, invalidatedByClose(candles[d.Index], StudyShort, projectedD), "bearish AB=CD projection near D"))
			}
		}
	}
	return out
}

func directABCDSetups(candles []exchanges.Candle) []StrategySetup {
	out := make([]StrategySetup, 0)
	for i := 0; i+3 < len(candles); i++ {
		a, b, c, d := candles[i], candles[i+1], candles[i+2], candles[i+3]
		abDown := a.HighFloat() - b.LowFloat()
		if abDown > 0 && c.HighFloat() < a.HighFloat() {
			retrace := (c.HighFloat() - b.LowFloat()) / abDown
			projectedD := c.HighFloat() - abDown
			if retrace >= 0.50 && nearLevel(d.LowFloat(), projectedD) {
				out = append(out, setupAt(candles, i+3, "abcd", StudyLong, projectedD, projectedD, true, invalidatedByClose(d, StudyLong, projectedD), "bullish AB=CD projection near D"))
			}
		}
		abUp := b.HighFloat() - a.LowFloat()
		if abUp > 0 && c.LowFloat() > a.LowFloat() {
			retrace := (b.HighFloat() - c.LowFloat()) / abUp
			projectedD := c.LowFloat() + abUp
			if retrace >= 0.50 && nearLevel(d.HighFloat(), projectedD) {
				out = append(out, setupAt(candles, i+3, "abcd", StudyShort, projectedD, projectedD, true, invalidatedByClose(d, StudyShort, projectedD), "bearish AB=CD projection near D"))
			}
		}
	}
	return out
}

type swingPivot struct {
	Index int
	Price float64
	Kind  string
}

func swingPivots(candles []exchanges.Candle) []swingPivot {
	out := make([]swingPivot, 0)
	if len(candles) >= 2 {
		switch {
		case candles[0].HighFloat() > candles[1].HighFloat():
			out = appendPivot(out, swingPivot{Index: 0, Price: candles[0].HighFloat(), Kind: "high"})
		case candles[0].LowFloat() < candles[1].LowFloat():
			out = appendPivot(out, swingPivot{Index: 0, Price: candles[0].LowFloat(), Kind: "low"})
		}
	}
	for i := 1; i+1 < len(candles); i++ {
		if candles[i].HighFloat() > candles[i-1].HighFloat() && candles[i].HighFloat() > candles[i+1].HighFloat() {
			out = appendPivot(out, swingPivot{Index: i, Price: candles[i].HighFloat(), Kind: "high"})
		}
		if candles[i].LowFloat() < candles[i-1].LowFloat() && candles[i].LowFloat() < candles[i+1].LowFloat() {
			out = appendPivot(out, swingPivot{Index: i, Price: candles[i].LowFloat(), Kind: "low"})
		}
	}
	last := len(candles) - 1
	if last >= 1 {
		switch {
		case candles[last].HighFloat() > candles[last-1].HighFloat():
			out = appendPivot(out, swingPivot{Index: last, Price: candles[last].HighFloat(), Kind: "high"})
		case candles[last].LowFloat() < candles[last-1].LowFloat():
			out = appendPivot(out, swingPivot{Index: last, Price: candles[last].LowFloat(), Kind: "low"})
		}
	}
	return out
}

func appendPivot(pivots []swingPivot, pivot swingPivot) []swingPivot {
	if len(pivots) == 0 || pivots[len(pivots)-1].Kind != pivot.Kind {
		return append(pivots, pivot)
	}
	last := pivots[len(pivots)-1]
	if (pivot.Kind == "high" && pivot.Price > last.Price) || (pivot.Kind == "low" && pivot.Price < last.Price) {
		pivots[len(pivots)-1] = pivot
	}
	return pivots
}
