package marketdata

type EventType string

const (
	EventTicker EventType = "ticker"
	EventTrade  EventType = "trade"
	EventCandle EventType = "candle"
)

type MarketEvent struct {
	Type   EventType
	Venue  string
	Symbol string
	Time   int64

	Ticker any
	Trade  any
	Candle any
}