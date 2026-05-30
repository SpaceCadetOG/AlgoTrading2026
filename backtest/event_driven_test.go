package backtest

import (
	"testing"

	"AlgoTrading2026/exchanges"
)

func TestEventDrivenBacktesterProcessesCandlesAndAdvancesClock(t *testing.T) {
	candles := []exchanges.Candle{
		{Symbol: "BTC", Close: "100", StartTime: 3000},
		{Symbol: "BTC", Close: "101", StartTime: 1000},
		{Symbol: "BTC", Close: "102", StartTime: 2000},
	}

	backtester := NewEventDrivenBacktester(EventDrivenBacktestConfig{
		Symbol:              "BTC",
		StartingCash:        10000,
		UseSimulatedGateway: true,
		AllowLiveOrders:     true,
	})

	if backtester.Config.AllowLiveOrders {
		t.Fatal("event-driven backtester must force allowLiveOrders=false")
	}

	result, err := backtester.Run(candles)
	if err != nil {
		t.Fatalf("run event-driven backtest: %v", err)
	}

	if result.CandlesProcessed != 3 {
		t.Fatalf("expected three candles processed, got %d", result.CandlesProcessed)
	}
	if result.ClockStart.UnixMilli() != 1000 || result.ClockEnd.UnixMilli() != 3000 {
		t.Fatalf("unexpected clock range: %s -> %s", result.ClockStart, result.ClockEnd)
	}
}

func TestEventDrivenBacktesterCreatesAndFillsOrders(t *testing.T) {
	candles := []exchanges.Candle{
		{Symbol: "BTC", Close: "100", StartTime: 1000},
	}

	backtester := NewEventDrivenBacktester(EventDrivenBacktestConfig{
		Symbol:              "BTC",
		StartingCash:        10000,
		UseSimulatedGateway: true,
	})

	result, err := backtester.Run(candles)
	if err != nil {
		t.Fatalf("run event-driven backtest: %v", err)
	}

	if result.OrdersCreated != 2 {
		t.Fatalf("expected buy/sell order pair, got %d orders", result.OrdersCreated)
	}
	if result.OrdersFilled != 2 {
		t.Fatalf("expected two fills, got %d", result.OrdersFilled)
	}
	if result.FinalPosition != 0 {
		t.Fatalf("expected flat final position, got %.4f", result.FinalPosition)
	}
	if result.FinalPnL <= 0 {
		t.Fatalf("expected deterministic spread capture, got pnl %.4f", result.FinalPnL)
	}
	if result.AuditEvents < 4 {
		t.Fatalf("expected accept/fill audit events, got %d", result.AuditEvents)
	}
}

func TestEventDrivenBacktesterEmptyInput(t *testing.T) {
	backtester := NewEventDrivenBacktester(EventDrivenBacktestConfig{StartingCash: 1234})
	result, err := backtester.Run(nil)
	if err != nil {
		t.Fatalf("empty run returned error: %v", err)
	}
	if result.CandlesProcessed != 0 || result.FinalCash != 1234 {
		t.Fatalf("unexpected empty result: %+v", result)
	}
}
