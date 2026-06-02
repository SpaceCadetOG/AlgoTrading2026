package orderbook

import "strconv"

type BookLevel struct {
	Price string
	Size  string
}

type OrderBookSnapshot struct {
	Venue      string
	Symbol     string
	Time       int64
	Sequence   string
	Bids       []BookLevel
	Asks       []BookLevel
	IsSnapshot bool
}

func (l BookLevel) PriceFloat() float64 {
	return parseFloat(l.Price)
}

func (l BookLevel) SizeFloat() float64 {
	return parseFloat(l.Size)
}

func (l BookLevel) Notional() float64 {
	return l.PriceFloat() * l.SizeFloat()
}

func parseFloat(value string) float64 {
	parsed, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return 0
	}
	return parsed
}
