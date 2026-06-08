package backtest

import "AlgoTrading2026/exchanges"

type Side string

const (
	SideLong  Side = "LONG"
	SideShort Side = "SHORT"
)

type SameBarPolicy string

const (
	SameBarStopFirst   SameBarPolicy = "stop_first"
	SameBarTargetFirst SameBarPolicy = "target_first"
)

type BarExit struct {
	Hit    bool
	Reason string
	Price  float64
}

func EvaluateBarExit(side Side, candle exchanges.Candle, stop float64, target float64, policy SameBarPolicy) BarExit {
	if stop <= 0 && target <= 0 {
		return BarExit{}
	}
	high := candle.HighFloat()
	low := candle.LowFloat()
	if high <= 0 || low <= 0 {
		return BarExit{}
	}

	stopHit := false
	targetHit := false
	switch side {
	case SideLong:
		stopHit = stop > 0 && low <= stop
		targetHit = target > 0 && high >= target
	case SideShort:
		stopHit = stop > 0 && high >= stop
		targetHit = target > 0 && low <= target
	default:
		return BarExit{}
	}

	if stopHit && targetHit {
		if policy == SameBarTargetFirst {
			return BarExit{Hit: true, Reason: "target", Price: target}
		}
		return BarExit{Hit: true, Reason: "stop", Price: stop}
	}
	if stopHit {
		return BarExit{Hit: true, Reason: "stop", Price: stop}
	}
	if targetHit {
		return BarExit{Hit: true, Reason: "target", Price: target}
	}
	return BarExit{}
}
