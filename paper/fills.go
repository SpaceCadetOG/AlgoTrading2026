package paper

import (
	"fmt"
	"math"

	"AlgoTrading2026/orderbook"
)

type FillResult struct {
	AveragePrice float64
	FilledQty    float64
	UnfilledQty  float64
	PartialFill  bool
	FeePaid      float64
	SlippageBps  float64
}

func SimulateEntryFill(snapshot orderbook.OrderBookSnapshot, side string, quantity float64, thinSession bool, cfg Config) (FillResult, error) {
	return simulateFill(snapshot, side, quantity, thinSession, cfg, false, false, false)
}

func SimulateExitFill(snapshot orderbook.OrderBookSnapshot, side string, quantity float64, thinSession bool, stopExit bool, forceExit bool, cfg Config) (FillResult, error) {
	return simulateFill(snapshot, side, quantity, thinSession, cfg, stopExit, forceExit, true)
}

func simulateFill(snapshot orderbook.OrderBookSnapshot, side string, quantity float64, thinSession bool, cfg Config, stopExit bool, forceExit bool, exitFill bool) (FillResult, error) {
	if quantity <= 0 {
		return FillResult{}, fmt.Errorf("quantity must be positive")
	}

	levels := entryLevels(snapshot, side)
	if len(levels) == 0 {
		return FillResult{}, fmt.Errorf("no order book levels available")
	}

	remaining := quantity
	notional := 0.0
	filled := 0.0
	for _, level := range levels {
		price := level.PriceFloat()
		size := level.SizeFloat()
		if price <= 0 || size <= 0 {
			continue
		}
		take := math.Min(size, remaining)
		notional += take * price
		filled += take
		remaining -= take
		if remaining <= 0 {
			break
		}
	}

	if filled == 0 {
		return FillResult{}, fmt.Errorf("no visible depth available")
	}

	avg := notional / filled
	slip := baseSlippageBps(thinSession, stopExit, forceExit)
	if remaining > 0 && !cfg.AllowPartialFills {
		last := levels[len(levels)-1].PriceFloat()
		penalized := applyDirectionalBps(last, side, slip+12)
		notional += remaining * penalized
		filled += remaining
		remaining = 0
		avg = notional / filled
		slip += 12
	}
	avg = applyDirectionalBps(avg, side, slip)
	feeBps := cfg.TakerFeeBps
	if exitFill && !stopExit && !forceExit {
		feeBps = cfg.MakerFeeBps
	}
	if stopExit || forceExit {
		feeBps = cfg.TakerFeeBps
	}
	fee := avg * filled * feeBps / 10000

	return FillResult{
		AveragePrice: avg,
		FilledQty:    filled,
		UnfilledQty:  remaining,
		PartialFill:  remaining > 0,
		FeePaid:      fee,
		SlippageBps:  slip,
	}, nil
}

func entryLevels(snapshot orderbook.OrderBookSnapshot, side string) []orderbook.BookLevel {
	if stringsEqualFold(side, "LONG") || stringsEqualFold(side, "BUY") {
		return append([]orderbook.BookLevel(nil), snapshot.Asks...)
	}
	return append([]orderbook.BookLevel(nil), snapshot.Bids...)
}

func baseSlippageBps(thinSession bool, stopExit bool, forceExit bool) float64 {
	slip := 1.0
	if thinSession {
		slip = 4.0
	}
	if stopExit {
		slip += 3.0
	}
	if forceExit {
		slip += 5.0
	}
	return slip
}

func applyDirectionalBps(price float64, side string, bps float64) float64 {
	if price <= 0 {
		return 0
	}
	if stringsEqualFold(side, "LONG") || stringsEqualFold(side, "BUY") {
		return price * (1 + bps/10000)
	}
	return price * (1 - bps/10000)
}
