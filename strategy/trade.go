package strategy

import "time"

type TradeStatus string

const (
	TradeOpened TradeStatus = "OPEN"
	TradeClosed TradeStatus = "CLOSED"
)

type Trade struct {
	Venue       string
	Symbol      string
	Side        string
	Size        float64
	EntryPrice  float64
	ExitPrice   float64
	OpenedAt    time.Time
	ClosedAt    time.Time
	GrossPnL    float64
	RealizedPnL float64
	Fees        float64
	ExitReason  string
	Status      TradeStatus
}

func NewOpenedTrade(venue string, symbol string, side string, size float64, entryPrice float64, openedAt time.Time) Trade {
	return Trade{
		Venue:      venue,
		Symbol:     symbol,
		Side:       NormalizeSide(side, size),
		Size:       size,
		EntryPrice: entryPrice,
		OpenedAt:   openedAt,
		Status:     TradeOpened,
	}
}

func (t Trade) Close(exitPrice float64, fees float64, reason string, closedAt time.Time) Trade {
	t.ExitPrice = exitPrice
	t.Fees = fees
	t.ExitReason = reason
	t.ClosedAt = closedAt
	t.GrossPnL = UnrealizedPnL(t.Side, t.Size, t.EntryPrice, exitPrice)
	t.RealizedPnL = t.GrossPnL - fees
	t.Status = TradeClosed
	return t
}
