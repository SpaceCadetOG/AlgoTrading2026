package marketdata

import "AlgoTrading2026/exchanges"

func NewTickerEvent(t exchanges.Ticker) MarketEvent {
	return MarketEvent{
		Type:   EventTicker,
		Venue:  t.Venue,
		Symbol: t.Symbol,
		Time:   t.Time,
		Ticker: t,
	}
}

func NewTradeEvent(t exchanges.Trade) MarketEvent {
	return MarketEvent{
		Type:   EventTrade,
		Venue:  t.Venue,
		Symbol: t.Symbol,
		Time:   t.Time,
		Trade:  t,
	}
}

func NewCandleEvent(c exchanges.Candle) MarketEvent {
	return MarketEvent{
		Type:   EventCandle,
		Venue:  c.Venue,
		Symbol: c.Symbol,
		Time:   c.EndTime,
		Candle: c,
	}
}
