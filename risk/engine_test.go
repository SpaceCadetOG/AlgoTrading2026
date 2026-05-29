package risk

import (
	"strings"
	"testing"

	"AlgoTrading2026/execution"
	"AlgoTrading2026/strategy"
)

func TestReduceOnlyAllowedWhenLiveDisabled(t *testing.T) {
	engine := NewEngine(DefaultLimits())

	decision := engine.CheckOrder(testAccount(), nil, testOrder(func(order *execution.OrderRequest) {
		order.ReduceOnly = true
	}))

	if !decision.Approved {
		t.Fatalf("expected reduce-only approval, got %+v", decision)
	}
}

func TestReduceOnlyRejectedWhenKillSwitchActive(t *testing.T) {
	limits := DefaultLimits()
	limits.KillSwitchActive = true
	engine := NewEngine(limits)

	decision := engine.CheckOrder(testAccount(), nil, testOrder(func(order *execution.OrderRequest) {
		order.ReduceOnly = true
	}))

	assertRejectedContains(t, decision, "kill switch")
}

func TestLiveDisabledRejectsNewOrders(t *testing.T) {
	engine := NewEngine(DefaultLimits())

	decision := engine.CheckOrder(testAccount(), nil, testOrder())

	assertRejectedContains(t, decision, "live trading disabled")
}

func TestKillSwitchRejectsAllNonReduceOrders(t *testing.T) {
	limits := DefaultLimits()
	limits.LiveTradingEnabled = true
	limits.KillSwitchActive = true
	engine := NewEngine(limits)

	decision := engine.CheckOrder(testAccount(), nil, testOrder())

	assertRejectedContains(t, decision, "kill switch")
}

func TestBadPriceOrSizeRejects(t *testing.T) {
	limits := DefaultLimits()
	limits.LiveTradingEnabled = true
	engine := NewEngine(limits)

	assertRejectedContains(t, engine.CheckOrder(testAccount(), nil, testOrder(func(order *execution.OrderRequest) {
		order.Price = "nope"
	})), "invalid order price")

	assertRejectedContains(t, engine.CheckOrder(testAccount(), nil, testOrder(func(order *execution.OrderRequest) {
		order.Size = "0"
	})), "invalid order size")
}

func TestMaxPositionsRejects(t *testing.T) {
	limits := generousLimits()
	limits.MaxOpenPositions = 1
	engine := NewEngine(limits)

	decision := engine.CheckOrder(testAccount(), []strategy.PositionState{testPosition("BTC", 50)}, testOrder())

	assertRejectedContains(t, decision, "max open positions")
}

func TestMaxTotalExposureRejects(t *testing.T) {
	limits := generousLimits()
	limits.MaxTotalExposureUSD = 55
	engine := NewEngine(limits)

	decision := engine.CheckOrder(testAccount(), []strategy.PositionState{testPosition("ETH", 50)}, testOrder())

	assertRejectedContains(t, decision, "max total exposure")
}

func TestMaxSymbolExposureRejects(t *testing.T) {
	limits := generousLimits()
	limits.MaxSymbolExposureUSD = 55
	engine := NewEngine(limits)

	decision := engine.CheckOrder(testAccount(), []strategy.PositionState{testPosition("BTC", 50)}, testOrder())

	assertRejectedContains(t, decision, "max symbol exposure")
}

func TestEnoughAvailableEquityApproves(t *testing.T) {
	engine := NewEngine(generousLimits())

	decision := engine.CheckOrder(testAccount(), nil, testOrder())

	if !decision.Approved {
		t.Fatalf("expected approval, got %+v", decision)
	}
}

func TestMinAvailableRejects(t *testing.T) {
	limits := generousLimits()
	limits.MinAvailableUSD = 995
	engine := NewEngine(limits)

	decision := engine.CheckOrder(testAccount(), nil, testOrder())

	assertRejectedContains(t, decision, "min available USD")
}

func TestMaxLeverageRejects(t *testing.T) {
	limits := generousLimits()
	limits.MaxLeverage = 0.005
	engine := NewEngine(limits)

	decision := engine.CheckOrder(testAccount(), nil, testOrder())

	assertRejectedContains(t, decision, "max leverage")
}

func testOrder(modifiers ...func(*execution.OrderRequest)) execution.OrderRequest {
	order := execution.OrderRequest{
		Venue:  "aster",
		Symbol: "BTC",
		Side:   execution.Buy,
		Type:   execution.Limit,
		Size:   "0.01",
		Price:  "1000",
	}

	for _, modify := range modifiers {
		modify(&order)
	}

	return order
}

func testAccount() strategy.AccountState {
	return strategy.AccountState{
		TotalEquity:     1000,
		AvailableEquity: 1000,
	}
}

func testPosition(symbol string, exposure float64) strategy.PositionState {
	return strategy.PositionState{
		Venue:       "test",
		Symbol:      symbol,
		ExposureUSD: exposure,
	}
}

func generousLimits() Limits {
	limits := DefaultLimits()
	limits.LiveTradingEnabled = true
	limits.MaxOpenPositions = 10
	limits.MaxTotalExposureUSD = 10000
	limits.MaxSymbolExposureUSD = 10000
	limits.MaxLeverage = 10
	limits.MinAvailableUSD = 10
	return limits
}

func assertRejectedContains(t *testing.T, decision Decision, want string) {
	t.Helper()
	if decision.Approved {
		t.Fatalf("expected rejection containing %q, got approval %+v", want, decision)
	}
	if !strings.Contains(decision.Reason, want) {
		t.Fatalf("expected reason containing %q, got %q", want, decision.Reason)
	}
}
