package orderbook

func BestBid(snapshot OrderBookSnapshot) BookLevel {
	if len(snapshot.Bids) == 0 {
		return BookLevel{}
	}
	best := snapshot.Bids[0]
	for _, level := range snapshot.Bids[1:] {
		if level.PriceFloat() > best.PriceFloat() {
			best = level
		}
	}
	return best
}

func BestAsk(snapshot OrderBookSnapshot) BookLevel {
	if len(snapshot.Asks) == 0 {
		return BookLevel{}
	}
	best := snapshot.Asks[0]
	for _, level := range snapshot.Asks[1:] {
		price := level.PriceFloat()
		if price > 0 && (best.PriceFloat() == 0 || price < best.PriceFloat()) {
			best = level
		}
	}
	return best
}

func Spread(snapshot OrderBookSnapshot) float64 {
	bid := BestBid(snapshot).PriceFloat()
	ask := BestAsk(snapshot).PriceFloat()
	if bid == 0 || ask == 0 {
		return 0
	}
	return ask - bid
}

func SpreadPct(snapshot OrderBookSnapshot) float64 {
	mid := Mid(snapshot)
	if mid == 0 {
		return 0
	}
	return Spread(snapshot) / mid * 100
}

func Mid(snapshot OrderBookSnapshot) float64 {
	bid := BestBid(snapshot).PriceFloat()
	ask := BestAsk(snapshot).PriceFloat()
	if bid == 0 || ask == 0 {
		return 0
	}
	return (bid + ask) / 2
}

func DepthWithinPct(snapshot OrderBookSnapshot, pct float64) float64 {
	return BidDepthWithinPct(snapshot, pct) + AskDepthWithinPct(snapshot, pct)
}

func BidDepthWithinPct(snapshot OrderBookSnapshot, pct float64) float64 {
	center := Mid(snapshot)
	if center == 0 || pct < 0 {
		return 0
	}
	minPrice := center * (1 - pct/100)
	var total float64
	for _, level := range snapshot.Bids {
		price := level.PriceFloat()
		if price >= minPrice && price <= center {
			total += level.Notional()
		}
	}
	return total
}

func AskDepthWithinPct(snapshot OrderBookSnapshot, pct float64) float64 {
	center := Mid(snapshot)
	if center == 0 || pct < 0 {
		return 0
	}
	maxPrice := center * (1 + pct/100)
	var total float64
	for _, level := range snapshot.Asks {
		price := level.PriceFloat()
		if price >= center && price <= maxPrice {
			total += level.Notional()
		}
	}
	return total
}

func Imbalance(snapshot OrderBookSnapshot, pct float64) float64 {
	bidDepth := BidDepthWithinPct(snapshot, pct)
	askDepth := AskDepthWithinPct(snapshot, pct)
	total := bidDepth + askDepth
	if total == 0 {
		return 0
	}
	return (bidDepth - askDepth) / total
}

func LiquidityNearPrice(snapshot OrderBookSnapshot, center float64, pctBand float64) (float64, float64) {
	if center <= 0 || pctBand < 0 {
		return 0, 0
	}
	lower := center * (1 - pctBand/100)
	upper := center * (1 + pctBand/100)

	var bidLiquidity float64
	for _, level := range snapshot.Bids {
		price := level.PriceFloat()
		if price >= lower && price <= center {
			bidLiquidity += level.Notional()
		}
	}

	var askLiquidity float64
	for _, level := range snapshot.Asks {
		price := level.PriceFloat()
		if price >= center && price <= upper {
			askLiquidity += level.Notional()
		}
	}

	return bidLiquidity, askLiquidity
}
