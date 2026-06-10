package runtime

import "testing"

func TestHybridReconcilerMergesWebsocketAndPollingState(t *testing.T) {
	hybrid := NewHybridReconciler()
	hybrid.ApplyStreamEvent(StreamEvent{
		Venue: "lighter",
		Orders: []OrderSnapshot{{
			Venue:        "lighter",
			Symbol:       "BTC",
			OrderID:      "44",
			Status:       "OPEN",
			OrderType:    "take-profit",
			RequestedQty: 1,
			FilledQty:    0.5,
		}},
		Fills: []FillSnapshot{{
			Venue:    "lighter",
			Symbol:   "BTC",
			OrderID:  "44",
			FillID:   "t1",
			Quantity: 0.5,
			Price:    65000,
		}},
	})
	hybrid.ApplyPollingResult(ReconcileResult{
		Venue:  "lighter",
		Source: ReconcileSourcePolling,
		Positions: []PositionSnapshot{{
			Venue:    "lighter",
			Symbol:   "BTC",
			Side:     "LONG",
			Quantity: 0.5,
			Entry:    64000,
		}},
	})
	state := hybrid.Snapshot()
	if state.ByVenue["lighter"].Source != ReconcileSourceHybrid {
		t.Fatalf("expected hybrid source, got %+v", state.ByVenue["lighter"])
	}
	if len(state.Orders) != 1 || len(state.Fills) != 1 || len(state.Positions) != 1 {
		t.Fatalf("unexpected merged state: %+v", state)
	}
}

func TestHybridReconcilerClassifiesMismatches(t *testing.T) {
	hybrid := NewHybridReconciler()
	hybrid.ApplyStreamEvent(StreamEvent{
		Venue: "aster",
		Fills: []FillSnapshot{{
			Venue:    "aster",
			Symbol:   "ETHUSDT",
			OrderID:  "missing",
			FillID:   "fill",
			Quantity: 1,
		}},
		Positions: []PositionSnapshot{{
			Venue:    "aster",
			Symbol:   "DOGEUSDT",
			Side:     "LONG",
			Quantity: 10,
		}},
	})
	state := hybrid.Snapshot()
	if !hasMismatch(state.Mismatches, "order_missing_on_venue") || !hasMismatch(state.Mismatches, "orphan_position_on_venue") {
		t.Fatalf("expected mismatch classifications, got %+v", state.Mismatches)
	}
}

func TestStreamStatusMarksStale(t *testing.T) {
	hybrid := NewHybridReconciler()
	hybrid.SetStreamStatus(StreamStatus{Venue: "lighter", Supported: true, Connected: false, Stale: true, RequiresFreshStream: true})
	state := hybrid.Snapshot()
	if len(state.Streams) != 1 || !state.Streams[0].Stale {
		t.Fatalf("expected stale stream status: %+v", state.Streams)
	}
}

func hasMismatch(rows []ReconciliationMismatch, typ string) bool {
	for _, row := range rows {
		if row.Type == typ {
			return true
		}
	}
	return false
}
