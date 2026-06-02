package l2recorder

import (
	"time"

	"AlgoTrading2026/orderbook"
)

type SnapshotRow struct {
	Timestamp int64
	Venue     string
	Symbol    string

	BestBid   float64
	BestAsk   float64
	Spread    float64
	SpreadPct float64
	Mid       float64

	BidDepth1Pct  float64
	AskDepth1Pct  float64
	Imbalance1Pct float64

	Valid bool
	Error string
}

func SnapshotToRow(snapshot orderbook.OrderBookSnapshot, err error, now time.Time) SnapshotRow {
	timestamp := snapshot.Time
	if timestamp == 0 {
		timestamp = now.UnixMilli()
	}
	row := SnapshotRow{
		Timestamp: timestamp,
		Venue:     snapshot.Venue,
		Symbol:    snapshot.Symbol,
	}
	if err != nil {
		row.Error = err.Error()
		return row
	}
	if validationErr := orderbook.ValidateSnapshot(snapshot); validationErr != nil {
		row.Error = validationErr.Error()
		return row
	}
	row.BestBid = orderbook.BestBid(snapshot).PriceFloat()
	row.BestAsk = orderbook.BestAsk(snapshot).PriceFloat()
	row.Spread = orderbook.Spread(snapshot)
	row.SpreadPct = orderbook.SpreadPct(snapshot)
	row.Mid = orderbook.Mid(snapshot)
	row.BidDepth1Pct = orderbook.BidDepthWithinPct(snapshot, 1)
	row.AskDepth1Pct = orderbook.AskDepthWithinPct(snapshot, 1)
	row.Imbalance1Pct = orderbook.Imbalance(snapshot, 1)
	row.Valid = true
	return row
}

func ErrorRow(venue string, symbol string, err error, now time.Time) SnapshotRow {
	row := SnapshotRow{
		Timestamp: now.UnixMilli(),
		Venue:     venue,
		Symbol:    symbol,
	}
	if err != nil {
		row.Error = err.Error()
	}
	return row
}
