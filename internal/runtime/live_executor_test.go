package runtime

import (
	"errors"
	"testing"

	"AlgoTrading2026/exchanges"
	"AlgoTrading2026/execution"
)

func TestLiveGateRefusesWhenNotArmed(t *testing.T) {
	gate := LiveGate{VenueHealthy: true, AccountReady: true}
	if got := gate.Refusal(Candidate{Venue: "aster"}); got != "live_trading_not_enabled" {
		t.Fatalf("refusal=%q", got)
	}
	gate.LiveEnabled = true
	if got := gate.Refusal(Candidate{Venue: "aster"}); got != "execution_not_enabled" {
		t.Fatalf("refusal=%q", got)
	}
	gate.ExecutionEnabled = true
	gate.AccountReady = false
	if got := gate.Refusal(Candidate{Venue: "aster"}); got != "account_not_ready" {
		t.Fatalf("refusal=%q", got)
	}
}

type mockPlacer struct {
	result *execution.OrderResult
	order  execution.OrderRequest
}

func (m *mockPlacer) PlaceOrder(order execution.OrderRequest) (*execution.OrderResult, error) {
	m.order = order
	return m.result, nil
}

func TestLiveExecutorBuildsSharedExecutionResult(t *testing.T) {
	placer := &mockPlacer{result: &execution.OrderResult{
		Success: true,
		Venue:   "aster",
		Symbol:  "BTCUSDT",
		OrderID: "42",
		Status:  "NEW",
	}}
	executor := LiveExecutor{
		Placer: placer,
		Gate:   LiveGate{LiveEnabled: true, ExecutionEnabled: true, VenueHealthy: true, AccountReady: true},
	}
	result, err := executor.Execute(ExecutionDecision{
		Candidate: Candidate{Venue: "aster", Symbol: "BTCUSDT", Side: "LONG", Quantity: 0.1, EntryPrice: 100},
		Risk:      RiskDecision{Allowed: true},
		Mode:      "live",
	})
	if err != nil {
		t.Fatalf("execute: %v", err)
	}
	if !result.Accepted || result.Order == nil || result.Order.OrderID != "42" {
		t.Fatalf("unexpected execution result: %+v", result)
	}
	if placer.order.Venue != "aster" || placer.order.Side != execution.Buy || placer.order.Size != "0.1" {
		t.Fatalf("unexpected placed order: %+v", placer.order)
	}
	if result.Request.Protection.Mode != "native_stop_tp_runtime_trailing" || !result.Request.Protection.ReduceOnly {
		t.Fatalf("expected protection plan on live request: %+v", result.Request.Protection)
	}
}

func TestProtectionModeSelectionByVenue(t *testing.T) {
	asterPlan := BuildProtectionPlan(Candidate{Venue: "aster", StopPrice: 99, TP1: 101})
	if asterPlan.Mode != "native_stop_tp_runtime_trailing" || !asterPlan.StopArmed || !asterPlan.TPLadderArmed {
		t.Fatalf("unexpected aster protection plan: %+v", asterPlan)
	}
	hyperPlan := BuildProtectionPlan(Candidate{Venue: "hyperliquid", StopPrice: 99, TP1: 101})
	if hyperPlan.Mode != "runtime_managed_protection" || !hyperPlan.TrailingArmed {
		t.Fatalf("unexpected hyperliquid protection plan: %+v", hyperPlan)
	}
	unknownPlan := BuildProtectionPlan(Candidate{Venue: "unknown", StopPrice: 99})
	if unknownPlan.Mode != "unsupported" {
		t.Fatalf("unexpected unknown protection plan: %+v", unknownPlan)
	}
}

func TestRoutedOrderPlacerPreservesVenueIdentity(t *testing.T) {
	asterPlacer := &mockPlacer{result: &execution.OrderResult{Success: true, Status: "NEW"}}
	router := RoutedOrderPlacer{Placers: map[string]execution.OrderPlacer{"aster": asterPlacer}}
	result, err := router.PlaceOrder(execution.OrderRequest{Venue: "aster", Symbol: "BTCUSDT"})
	if err != nil {
		t.Fatalf("route order: %v", err)
	}
	if !result.Success || result.Venue != "aster" || asterPlacer.order.Symbol != "BTCUSDT" {
		t.Fatalf("unexpected routed result=%+v order=%+v", result, asterPlacer.order)
	}
	missing, err := router.PlaceOrder(execution.OrderRequest{Venue: "lighter", Symbol: "BTC"})
	if err != nil {
		t.Fatalf("missing venue should refuse safely, got err=%v", err)
	}
	if missing.Success || missing.Message != "venue_order_placer_unavailable" {
		t.Fatalf("unexpected missing venue result: %+v", missing)
	}
}

type mockAccountProvider struct {
	snapshot *exchanges.AccountSnapshot
	err      error
}

func (m mockAccountProvider) GetAccountSnapshot() (*exchanges.AccountSnapshot, error) {
	return m.snapshot, m.err
}

type mockPositionProvider struct {
	positions []exchanges.Position
	err       error
}

func (m mockPositionProvider) GetPositions() ([]exchanges.Position, error) {
	return m.positions, m.err
}

func TestVenueHealthCheckerBlocksFailedAccountOrPositionSync(t *testing.T) {
	health := VenueHealthChecker{
		Venue:                 "hyperliquid",
		Account:               mockAccountProvider{snapshot: &exchanges.AccountSnapshot{Available: "100"}},
		Positions:             mockPositionProvider{err: errors.New("private endpoint down")},
		RequirePositions:      true,
		MarketDataReady:       true,
		PrivateEndpointsReady: true,
	}.Check()
	if health.Healthy || health.PositionSyncReady {
		t.Fatalf("expected position sync failure to block live health: %+v", health)
	}
	if health.RefusalReason() == "" {
		t.Fatalf("expected refusal reason: %+v", health)
	}
}

type mockOrderReader struct {
	orders []OrderSnapshot
	err    error
}

func (m mockOrderReader) OpenOrders(symbol string) ([]OrderSnapshot, error) {
	return m.orders, m.err
}

type mockFillReader struct {
	fills []FillSnapshot
	err   error
}

func (m mockFillReader) Fills(symbol string, orderID string) ([]FillSnapshot, error) {
	return m.fills, m.err
}

func TestVenueReconcilerReportsSyncedPositionAndMismatch(t *testing.T) {
	reconciler := VenueReconciler{
		Venue:  "hyperliquid",
		Orders: mockOrderReader{orders: []OrderSnapshot{{Venue: "hyperliquid", Symbol: "BTC", OrderID: "7", Status: "OPEN"}}},
		Positions: mockPositionProvider{positions: []exchanges.Position{
			{Venue: "hyperliquid", Symbol: "BTC", Side: "LONG", Size: "0.2", Entry: "100"},
		}},
	}
	result, err := reconciler.Reconcile(ReconcileRequest{
		Decision: ExecutionDecision{Candidate: Candidate{Venue: "hyperliquid", Symbol: "BTC"}},
		Result:   ExecutionResult{Accepted: true, Venue: "hyperliquid", Symbol: "BTC", OrderID: "7", Status: "NEW"},
	})
	if err != nil {
		t.Fatalf("reconcile: %v", err)
	}
	if result.Status != "position_synced" || len(result.Positions) != 1 || len(result.Mismatches) != 0 {
		t.Fatalf("expected synced position, got %+v", result)
	}

	mismatch, err := (VenueReconciler{Venue: "lighter"}).Reconcile(ReconcileRequest{
		Decision: ExecutionDecision{Candidate: Candidate{Venue: "lighter", Symbol: "BTC"}},
		Result:   ExecutionResult{Accepted: true, Venue: "lighter", Symbol: "BTC", OrderID: "8", Status: "NEW"},
	})
	if err != nil {
		t.Fatalf("reconcile mismatch: %v", err)
	}
	if len(mismatch.Mismatches) == 0 || mismatch.Events[len(mismatch.Events)-1] != "POSITION_MISMATCH" {
		t.Fatalf("expected unsupported position reconciliation mismatch, got %+v", mismatch)
	}
}

func TestVenueReconcilerReportsPartialAndFullFills(t *testing.T) {
	request := ReconcileRequest{
		Decision: ExecutionDecision{Candidate: Candidate{Venue: "aster", Symbol: "BTCUSDT"}},
		Result: ExecutionResult{
			Accepted: true,
			Venue:    "aster",
			Symbol:   "BTCUSDT",
			OrderID:  "9",
			Status:   "NEW",
			Request:  ExecutionRequest{RequestedQty: 1},
		},
	}
	partial, err := (VenueReconciler{
		Venue: "aster",
		Fills: mockFillReader{fills: []FillSnapshot{{Venue: "aster", Symbol: "BTCUSDT", OrderID: "9", Quantity: 0.4}}},
		Positions: mockPositionProvider{positions: []exchanges.Position{
			{Venue: "aster", Symbol: "BTCUSDT", Side: "LONG", Size: "0.4", Entry: "100"},
		}},
	}).Reconcile(request)
	if err != nil {
		t.Fatalf("partial reconcile: %v", err)
	}
	if partial.Status != "position_synced" || !hasEvent(partial.Events, "ORDER_PARTIAL_FILL") {
		t.Fatalf("expected partial fill and synced position, got %+v", partial)
	}

	full, err := (VenueReconciler{
		Venue: "aster",
		Fills: mockFillReader{fills: []FillSnapshot{{Venue: "aster", Symbol: "BTCUSDT", OrderID: "9", Quantity: 1}}},
		Positions: mockPositionProvider{positions: []exchanges.Position{
			{Venue: "aster", Symbol: "BTCUSDT", Side: "LONG", Size: "1", Entry: "100"},
		}},
	}).Reconcile(request)
	if err != nil {
		t.Fatalf("full reconcile: %v", err)
	}
	if full.Status != "position_synced" || !hasEvent(full.Events, "ORDER_FULL_FILL") {
		t.Fatalf("expected full fill and synced position, got %+v", full)
	}
}

func TestVenueHealthRequiresFillHistoryAndProtection(t *testing.T) {
	health := VenueHealthChecker{
		Venue:                 "aster",
		Account:               mockAccountProvider{snapshot: &exchanges.AccountSnapshot{Available: "100"}},
		Positions:             mockPositionProvider{},
		RequirePositions:      true,
		RequireFills:          true,
		MarketDataReady:       true,
		PrivateEndpointsReady: true,
		Protection:            ProtectionForVenue("aster"),
	}.Check()
	if health.Healthy || health.FillHistoryReady {
		t.Fatalf("expected missing fill reader to block health: %+v", health)
	}

	healthy := VenueHealthChecker{
		Venue:                 "lighter",
		Account:               mockAccountProvider{snapshot: &exchanges.AccountSnapshot{Available: "100"}},
		Positions:             mockPositionProvider{},
		RequirePositions:      true,
		RequireFills:          true,
		Fills:                 mockFillReader{},
		MarketDataReady:       true,
		PrivateEndpointsReady: true,
		Protection:            ProtectionForVenue("lighter"),
	}.Check()
	if !healthy.Healthy || !healthy.ReconciliationReady || !healthy.ProtectionReady {
		t.Fatalf("expected health when account, fills, positions, protection are ready: %+v", healthy)
	}
}

func TestVenueHealthCanRequireFreshStream(t *testing.T) {
	health := VenueHealthChecker{
		Venue:                 "lighter",
		Account:               mockAccountProvider{snapshot: &exchanges.AccountSnapshot{Available: "100"}},
		Positions:             mockPositionProvider{},
		RequirePositions:      true,
		RequireFills:          true,
		Fills:                 mockFillReader{},
		MarketDataReady:       true,
		PrivateEndpointsReady: true,
		Protection:            ProtectionForVenue("lighter"),
		RequireFreshStream:    true,
		Stream:                StreamStatus{Venue: "lighter", Supported: true, Connected: false, Stale: true, RequiresFreshStream: true},
	}.Check()
	if health.Healthy || health.StreamReady {
		t.Fatalf("expected stale required stream to block health: %+v", health)
	}
	if health.RefusalReason() == "" {
		t.Fatalf("expected refusal reason for stale stream")
	}
}

func hasEvent(events []string, want string) bool {
	for _, event := range events {
		if event == want {
			return true
		}
	}
	return false
}
