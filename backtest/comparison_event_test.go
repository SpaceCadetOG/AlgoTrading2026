package backtest

import (
	"testing"

	"AlgoTrading2026/exchanges"
)

func TestCompareForLoopVsEventDriven(t *testing.T) {
	candles := []exchanges.Candle{
		{Venue: "aster", Symbol: "BTCUSDT", Interval: "15m", Close: "100", StartTime: 1000, EndTime: 2000},
		{Venue: "aster", Symbol: "BTCUSDT", Interval: "15m", Close: "101", StartTime: 2000, EndTime: 3000},
		{Venue: "aster", Symbol: "BTCUSDT", Interval: "15m", Close: "99", StartTime: 3000, EndTime: 4000},
	}

	result, err := CompareForLoopVsEventDriven(
		candles,
		Config{
			Venue:           "aster",
			Symbol:          "BTCUSDT",
			Interval:        "15m",
			StartingBalance: 10000,
			FixedNotional:   100,
		},
		EventDrivenBacktestConfig{
			Symbol:              "BTCUSDT",
			StartingCash:        10000,
			MaxCandles:          3,
			UseSimulatedGateway: true,
			AllowLiveOrders:     true,
		},
	)
	if err != nil {
		t.Fatalf("compare backtesters: %v", err)
	}

	if result.Symbol != "BTCUSDT" || result.Candles != 3 {
		t.Fatalf("unexpected comparison identity: %+v", result)
	}
	if result.EventDrivenOrders == 0 || result.EventDrivenFills == 0 {
		t.Fatalf("expected event-driven orders/fills: %+v", result)
	}
	if result.AssumptionsDifference == "" {
		t.Fatal("expected assumption difference")
	}
	if result.Recommendation == "" {
		t.Fatal("expected recommendation")
	}
	_ = result.PnLDifference
}
