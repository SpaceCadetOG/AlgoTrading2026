package runtime

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

const (
	ReconcileSourcePolling   = "polling"
	ReconcileSourceWebsocket = "websocket"
	ReconcileSourceHybrid    = "hybrid"
)

type StreamStatus struct {
	Venue               string `json:"venue"`
	Supported           bool   `json:"supported"`
	Connected           bool   `json:"connected"`
	Source              string `json:"source"`
	LastEventTime       int64  `json:"lastEventTime,omitempty"`
	LastEventTimeUTC    string `json:"lastEventTimeUtc,omitempty"`
	PollingFallback     bool   `json:"pollingFallback"`
	Stale               bool   `json:"stale"`
	RequiresFreshStream bool   `json:"requiresFreshStream"`
	Message             string `json:"message,omitempty"`
}

type StreamEvent struct {
	Venue     string             `json:"venue"`
	Source    string             `json:"source"`
	Timestamp int64              `json:"timestamp,omitempty"`
	Orders    []OrderSnapshot    `json:"orders,omitempty"`
	Fills     []FillSnapshot     `json:"fills,omitempty"`
	Positions []PositionSnapshot `json:"positions,omitempty"`
	Raw       any                `json:"raw,omitempty"`
}

type ReconciliationState struct {
	Timestamp  int64                      `json:"timestamp"`
	TimeUTC    string                     `json:"timeUtc"`
	Streams    []StreamStatus             `json:"streams"`
	Orders     []OrderSnapshot            `json:"orders"`
	Fills      []FillSnapshot             `json:"fills"`
	Positions  []PositionSnapshot         `json:"positions"`
	Mismatches []ReconciliationMismatch   `json:"mismatches"`
	ByVenue    map[string]VenueReconState `json:"byVenue"`
}

type VenueReconState struct {
	Source    string             `json:"source"`
	Orders    []OrderSnapshot    `json:"orders"`
	Fills     []FillSnapshot     `json:"fills"`
	Positions []PositionSnapshot `json:"positions"`
}

type ReconciliationMismatch struct {
	Venue   string `json:"venue"`
	Symbol  string `json:"symbol,omitempty"`
	OrderID string `json:"orderId,omitempty"`
	Type    string `json:"type"`
	Message string `json:"message,omitempty"`
}

type HybridReconciler struct {
	states  map[string]*VenueReconState
	streams map[string]StreamStatus
}

func NewHybridReconciler() *HybridReconciler {
	return &HybridReconciler{
		states:  map[string]*VenueReconState{},
		streams: map[string]StreamStatus{},
	}
}

func (h *HybridReconciler) SetStreamStatus(status StreamStatus) {
	venue := normalizeVenue(status.Venue)
	if venue == "" {
		return
	}
	status.Venue = venue
	if status.Source == "" {
		status.Source = ReconcileSourceWebsocket
	}
	if status.LastEventTime > 0 && status.LastEventTimeUTC == "" {
		status.LastEventTimeUTC = time.UnixMilli(status.LastEventTime).UTC().Format(time.RFC3339)
	}
	h.streams[venue] = status
}

func (h *HybridReconciler) ApplyStreamEvent(event StreamEvent) {
	venue := normalizeVenue(event.Venue)
	if venue == "" {
		return
	}
	state := h.stateFor(venue)
	state.Source = ReconcileSourceWebsocket
	state.Orders = mergeOrderSnapshots(state.Orders, event.Orders)
	state.Fills = mergeFillSnapshots(state.Fills, event.Fills)
	state.Positions = mergePositionSnapshots(state.Positions, event.Positions)
	ts := event.Timestamp
	if ts <= 0 {
		ts = time.Now().UTC().UnixMilli()
	}
	h.SetStreamStatus(StreamStatus{
		Venue:           venue,
		Supported:       true,
		Connected:       true,
		Source:          ReconcileSourceWebsocket,
		LastEventTime:   ts,
		PollingFallback: true,
	})
}

func (h *HybridReconciler) ApplyPollingResult(result ReconcileResult) {
	venue := normalizeVenue(result.Venue)
	if venue == "" {
		return
	}
	state := h.stateFor(venue)
	if state.Source == ReconcileSourceWebsocket {
		state.Source = ReconcileSourceHybrid
	} else {
		state.Source = ReconcileSourcePolling
	}
	state.Orders = mergeOrderSnapshots(state.Orders, result.Orders)
	state.Fills = mergeFillSnapshots(state.Fills, result.Fills)
	state.Positions = mergePositionSnapshots(state.Positions, result.Positions)
	status := h.streams[venue]
	if status.Venue == "" {
		status = StreamStatus{Venue: venue, Source: ReconcileSourcePolling}
	}
	status.PollingFallback = true
	h.streams[venue] = status
}

func (h *HybridReconciler) Snapshot() ReconciliationState {
	now := time.Now().UTC()
	out := ReconciliationState{
		Timestamp: now.UnixMilli(),
		TimeUTC:   now.Format(time.RFC3339),
		ByVenue:   map[string]VenueReconState{},
	}
	venues := make([]string, 0, len(h.states))
	for venue := range h.states {
		venues = append(venues, venue)
	}
	for venue := range h.streams {
		if h.states[venue] == nil {
			venues = append(venues, venue)
		}
	}
	sort.Strings(venues)
	seenVenue := map[string]bool{}
	for _, venue := range venues {
		if seenVenue[venue] {
			continue
		}
		seenVenue[venue] = true
		if state := h.states[venue]; state != nil {
			row := *state
			out.ByVenue[venue] = row
			out.Orders = append(out.Orders, row.Orders...)
			out.Fills = append(out.Fills, row.Fills...)
			out.Positions = append(out.Positions, row.Positions...)
			out.Mismatches = append(out.Mismatches, classifyStateMismatches(venue, row)...)
		}
		if status, ok := h.streams[venue]; ok {
			out.Streams = append(out.Streams, status)
		}
	}
	return out
}

func (h *HybridReconciler) stateFor(venue string) *VenueReconState {
	if h.states[venue] == nil {
		h.states[venue] = &VenueReconState{}
	}
	return h.states[venue]
}

func WriteReconciliationState(root string, state ReconciliationState) error {
	if err := os.MkdirAll(root, 0o755); err != nil {
		return err
	}
	if err := writeRuntimeJSON(filepath.Join(root, "reconciliation_state.json"), state); err != nil {
		return err
	}
	return writeRuntimeJSON(filepath.Join(root, "reconciliation_mismatches.json"), state.Mismatches)
}

func writeRuntimeJSON(path string, value any) error {
	body, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, append(body, '\n'), 0o644)
}

func classifyStateMismatches(venue string, state VenueReconState) []ReconciliationMismatch {
	var out []ReconciliationMismatch
	openOrders := map[string]OrderSnapshot{}
	for _, order := range state.Orders {
		status := strings.ToUpper(strings.TrimSpace(order.Status))
		if status == "OPEN" || status == "NEW" || status == "ACKNOWLEDGED" || status == "PARTIAL_FILL" {
			openOrders[order.OrderID] = order
		}
		if order.OrderID == "" && order.ClientID == "" {
			out = append(out, ReconciliationMismatch{Venue: venue, Symbol: order.Symbol, Type: "unknown_order_state", Message: "order missing stable id"})
		}
	}
	fillQtyByOrder := map[string]float64{}
	for _, fill := range state.Fills {
		if fill.OrderID == "" {
			out = append(out, ReconciliationMismatch{Venue: venue, Symbol: fill.Symbol, Type: "runtime_missing_fill", Message: "fill missing order attribution"})
			continue
		}
		fillQtyByOrder[fill.OrderID] += fill.Quantity
		if _, ok := openOrders[fill.OrderID]; !ok {
			out = append(out, ReconciliationMismatch{Venue: venue, Symbol: fill.Symbol, OrderID: fill.OrderID, Type: "order_missing_on_venue", Message: "fill references order not present in latest order set"})
		}
	}
	for _, order := range openOrders {
		if order.RequestedQty > 0 && order.FilledQty > 0 && fillQtyByOrder[order.OrderID] > 0 && absFloat(order.FilledQty-fillQtyByOrder[order.OrderID]) > 1e-9 {
			out = append(out, ReconciliationMismatch{Venue: venue, Symbol: order.Symbol, OrderID: order.OrderID, Type: "quantity_mismatch", Message: "order filled quantity differs from fills"})
		}
		if isProtectionOrder(order) && !strings.Contains(strings.ToLower(order.OrderType), "reduce") && order.Side == "" {
			out = append(out, ReconciliationMismatch{Venue: venue, Symbol: order.Symbol, OrderID: order.OrderID, Type: "protection_missing", Message: "protection order lacks side/reduce-only evidence"})
		}
	}
	for _, position := range state.Positions {
		if position.Quantity != 0 && !hasSymbolOrderOrFill(state, position.Symbol) {
			out = append(out, ReconciliationMismatch{Venue: venue, Symbol: position.Symbol, Type: "orphan_position_on_venue", Message: "venue position has no matching order/fill in runtime state"})
		}
	}
	return out
}

func mergeFillSnapshots(base []FillSnapshot, extra []FillSnapshot) []FillSnapshot {
	seen := map[string]bool{}
	out := make([]FillSnapshot, 0, len(base)+len(extra))
	for _, row := range append(base, extra...) {
		key := strings.ToLower(row.Venue) + ":" + strings.ToUpper(row.Symbol) + ":" + row.OrderID + ":" + row.FillID
		if seen[key] {
			continue
		}
		seen[key] = true
		out = append(out, row)
	}
	return out
}

func mergePositionSnapshots(base []PositionSnapshot, extra []PositionSnapshot) []PositionSnapshot {
	seen := map[string]bool{}
	out := make([]PositionSnapshot, 0, len(base)+len(extra))
	for _, row := range append(base, extra...) {
		key := strings.ToLower(row.Venue) + ":" + strings.ToUpper(row.Symbol) + ":" + strings.ToUpper(row.Side)
		if seen[key] {
			continue
		}
		seen[key] = true
		out = append(out, row)
	}
	return out
}

func hasSymbolOrderOrFill(state VenueReconState, symbol string) bool {
	for _, order := range state.Orders {
		if strings.EqualFold(order.Symbol, symbol) {
			return true
		}
	}
	for _, fill := range state.Fills {
		if strings.EqualFold(fill.Symbol, symbol) {
			return true
		}
	}
	return false
}

func isProtectionOrder(order OrderSnapshot) bool {
	t := strings.ToLower(order.OrderType)
	return strings.Contains(t, "stop") || strings.Contains(t, "take-profit") || strings.Contains(t, "tp") || strings.Contains(t, "sl")
}

func normalizeVenue(venue string) string {
	return strings.ToLower(strings.TrimSpace(venue))
}

func absFloat(value float64) float64 {
	if value < 0 {
		return -value
	}
	return value
}
