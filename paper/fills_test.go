package paper

import (
	"math"
	"testing"

	"AlgoTrading2026/orderbook"
)

func TestSimulateEntryFillPartialAndDepthVWAP(t *testing.T) {
	cfg := DefaultConfig()
	cfg.AllowPartialFills = true
	snapshot := testSnapshot()

	fill, err := SimulateEntryFill(snapshot, "LONG", 5, false, cfg)
	if err != nil {
		t.Fatalf("entry fill: %v", err)
	}
	if !fill.PartialFill || fill.FilledQty != 3 || fill.UnfilledQty != 2 {
		t.Fatalf("unexpected partial fill: %+v", fill)
	}

	cfg.AllowPartialFills = false
	fill, err = SimulateEntryFill(snapshot, "LONG", 2, false, cfg)
	if err != nil {
		t.Fatalf("entry fill vwap: %v", err)
	}
	if math.Abs(fill.AveragePrice-101.51015) > 0.05 {
		t.Fatalf("unexpected average price: %.5f", fill.AveragePrice)
	}
}

func TestSimulateExitFillAdverseSlippageAndFees(t *testing.T) {
	cfg := DefaultConfig()
	snapshot := testSnapshot()

	tpFill, err := SimulateExitFill(snapshot, "SHORT", 1, false, false, false, cfg)
	if err != nil {
		t.Fatalf("tp fill: %v", err)
	}
	stopFill, err := SimulateExitFill(snapshot, "SHORT", 1, false, true, false, cfg)
	if err != nil {
		t.Fatalf("stop fill: %v", err)
	}
	if stopFill.SlippageBps <= tpFill.SlippageBps {
		t.Fatalf("expected stop slippage > tp slippage: tp=%+v stop=%+v", tpFill, stopFill)
	}
	if tpFill.FeePaid >= stopFill.FeePaid {
		t.Fatalf("expected maker-like TP fee < stop fee: tp=%+v stop=%+v", tpFill, stopFill)
	}
}

func testSnapshot() orderbook.OrderBookSnapshot {
	return orderbook.OrderBookSnapshot{
		Venue:  "aster",
		Symbol: "BTCUSDT",
		Bids: []orderbook.BookLevel{
			{Price: "99", Size: "1"},
			{Price: "98", Size: "2"},
		},
		Asks: []orderbook.BookLevel{
			{Price: "101", Size: "1"},
			{Price: "102", Size: "2"},
		},
	}
}
