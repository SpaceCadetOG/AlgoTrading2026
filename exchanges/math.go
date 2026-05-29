package exchanges

import "strconv"

func parseFloat(value string) float64 {
	v, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return 0
	}

	return v
}

func (t Ticker) BidFloat() float64 {
	return parseFloat(t.Bid)
}

func (t Ticker) AskFloat() float64 {
	return parseFloat(t.Ask)
}

func (t Ticker) LastFloat() float64 {
	return parseFloat(t.Last)
}

func (t Ticker) MidPrice() float64 {
	bid := t.BidFloat()
	ask := t.AskFloat()

	if bid == 0 || ask == 0 {
		return t.LastFloat()
	}

	return (bid + ask) / 2
}

func (t Ticker) Spread() float64 {
	return t.AskFloat() - t.BidFloat()
}

func (t Ticker) SpreadPct() float64 {
	mid := t.MidPrice()
	if mid == 0 {
		return 0
	}

	return t.Spread() / mid * 100
}

func (tr Trade) PriceFloat() float64 {
	return parseFloat(tr.Price)
}

func (tr Trade) SizeFloat() float64 {
	return parseFloat(tr.Size)
}

func (tr Trade) Notional() float64 {
	return tr.PriceFloat() * tr.SizeFloat()
}

func (t Ticker) BidSizeFloat() float64 {
	return parseFloat(t.BidSize)
}

func (t Ticker) AskSizeFloat() float64 {
	return parseFloat(t.AskSize)
}

func (t Ticker) BidUSD() float64 {
	return t.BidFloat() * t.BidSizeFloat()
}

func (t Ticker) AskUSD() float64 {
	return t.AskFloat() * t.AskSizeFloat()
}