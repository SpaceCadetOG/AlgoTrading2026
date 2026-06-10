package runtime

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"

	"AlgoTrading2026/exchanges"
	"AlgoTrading2026/execution"
)

type LiveGate struct {
	LiveEnabled      bool
	ExecutionEnabled bool
	VenueHealthy     bool
	AccountReady     bool
	KillSwitchActive bool
	AllowedVenues    map[string]bool
}

func LiveGateFromEnv() LiveGate {
	return LiveGate{
		LiveEnabled:      envBool("LIVE_ENABLE_LIVE_TRADING"),
		ExecutionEnabled: envBool("ENABLE_LIVE_ORDERS"),
		VenueHealthy:     envDefaultBool("LIVE_VENUE_HEALTHY", true),
		AccountReady:     envDefaultBool("LIVE_ACCOUNT_READY", false),
		KillSwitchActive: envBool("LIVE_KILL_SWITCH"),
		AllowedVenues:    allowedSet(os.Getenv("LIVE_ALLOWED_VENUES")),
	}
}

func (g LiveGate) Refusal(candidate Candidate) string {
	if !g.LiveEnabled {
		return "live_trading_not_enabled"
	}
	if !g.ExecutionEnabled {
		return "execution_not_enabled"
	}
	if !g.VenueHealthy {
		return "venue_not_healthy"
	}
	if !g.AccountReady {
		return "account_not_ready"
	}
	if g.KillSwitchActive {
		return "kill_switch_active"
	}
	if len(g.AllowedVenues) > 0 && !g.AllowedVenues[strings.ToLower(candidate.Venue)] {
		return "venue_not_allowed"
	}
	return ""
}

type LiveExecutor struct {
	Placer execution.OrderPlacer
	Gate   LiveGate
}

func (e LiveExecutor) Execute(decision ExecutionDecision) (ExecutionResult, error) {
	if !decision.Risk.Allowed {
		return ExecutionResult{Accepted: false, Message: strings.Join(decision.Risk.Reasons, ",")}, nil
	}
	if e.Placer == nil {
		return ExecutionResult{}, fmt.Errorf("live executor requires order placer")
	}
	if reason := e.Gate.Refusal(decision.Candidate); reason != "" {
		return ExecutionResult{Accepted: false, Venue: decision.Candidate.Venue, Symbol: decision.Candidate.Symbol, Status: "REFUSED", Message: reason}, nil
	}
	request := ExecutionRequest{
		Decision:      decision,
		OrderType:     string(execution.Market),
		RequestedQty:  decision.Candidate.Quantity,
		ExpectedEntry: decision.Candidate.EntryPrice,
		Protection:    BuildProtectionPlan(decision.Candidate),
	}
	order := execution.OrderRequest{
		Venue:  decision.Candidate.Venue,
		Symbol: decision.Candidate.Symbol,
		Side:   orderSide(decision.Candidate.Side),
		Type:   execution.Market,
		Size:   strconv.FormatFloat(decision.Candidate.Quantity, 'f', -1, 64),
		Price:  strconv.FormatFloat(decision.Candidate.EntryPrice, 'f', -1, 64),
	}
	result, err := e.Placer.PlaceOrder(order)
	if err != nil {
		return ExecutionResult{}, err
	}
	orderSnapshot := &OrderSnapshot{
		Venue:        firstNonEmpty(result.Venue, decision.Candidate.Venue),
		Symbol:       firstNonEmpty(result.Symbol, decision.Candidate.Symbol),
		Side:         decision.Candidate.Side,
		OrderID:      result.OrderID,
		Status:       normalizeOrderStatus(result.Status),
		OrderType:    string(order.Type),
		RequestedQty: decision.Candidate.Quantity,
		Price:        decision.Candidate.EntryPrice,
		Raw:          result.Raw,
	}
	return ExecutionResult{
		Accepted: result.Success,
		Venue:    result.Venue,
		Symbol:   result.Symbol,
		OrderID:  result.OrderID,
		Status:   orderSnapshot.Status,
		Message:  result.Message,
		Request:  request,
		Order:    orderSnapshot,
		Raw:      result.Raw,
	}, nil
}

type VenueReconciler struct {
	Venue     string
	Orders    OpenOrderReader
	Fills     FillReader
	Positions exchanges.PositionProvider
}

type OpenOrderReader interface {
	OpenOrders(symbol string) ([]OrderSnapshot, error)
}

type FillReader interface {
	Fills(symbol string, orderID string) ([]FillSnapshot, error)
}

func (r VenueReconciler) Reconcile(request ReconcileRequest) (ReconcileResult, error) {
	result := ReconcileResult{
		Venue:   firstNonEmpty(r.Venue, request.Decision.Candidate.Venue, request.Result.Venue),
		Symbol:  firstNonEmpty(request.Decision.Candidate.Symbol, request.Result.Symbol),
		OrderID: request.Result.OrderID,
		Status:  "unknown",
		Source:  "polling",
	}
	if request.Result.Order != nil {
		result.Orders = append(result.Orders, *request.Result.Order)
		result.Status = request.Result.Order.Status
	}
	if !request.Result.Accepted {
		result.Status = firstNonEmpty(request.Result.Status, "rejected")
		result.Events = append(result.Events, "ORDER_REJECTED")
		result.Message = request.Result.Message
		return result, nil
	}
	if request.Result.OrderID != "" || strings.EqualFold(request.Result.Status, "DRY_RUN") {
		result.Events = append(result.Events, "ORDER_ACKNOWLEDGED")
	}
	if r.Orders != nil {
		orders, err := r.Orders.OpenOrders(result.Symbol)
		if err != nil {
			result.Mismatches = append(result.Mismatches, "open_orders_unavailable:"+err.Error())
		} else {
			result.Orders = mergeOrderSnapshots(result.Orders, orders)
		}
	}
	if r.Fills != nil {
		fills, err := r.Fills.Fills(result.Symbol, result.OrderID)
		if err != nil {
			result.Mismatches = append(result.Mismatches, "fills_unavailable:"+err.Error())
		} else {
			result.Fills = fills
			filledQty := 0.0
			for _, fill := range fills {
				filledQty += fill.Quantity
			}
			if request.Result.Request.RequestedQty > 0 && filledQty > 0 {
				if filledQty+1e-12 >= request.Result.Request.RequestedQty {
					result.Events = append(result.Events, "ORDER_FULL_FILL")
					result.Status = "filled"
				} else {
					result.Events = append(result.Events, "ORDER_PARTIAL_FILL")
					result.Status = "partial_fill"
				}
			}
		}
	}
	if r.Positions != nil {
		positions, err := r.Positions.GetPositions()
		if err != nil {
			result.Mismatches = append(result.Mismatches, "positions_unavailable:"+err.Error())
		} else {
			result.Positions = normalizeExchangePositions(positions)
			if matchingPosition(result.Positions, result.Venue, result.Symbol) {
				result.Events = append(result.Events, "POSITION_SYNCED")
				result.Status = "position_synced"
			} else if request.Result.Accepted && request.Result.OrderID != "" {
				result.Mismatches = append(result.Mismatches, "accepted_order_without_position_snapshot")
			}
		}
	} else {
		result.Mismatches = append(result.Mismatches, "positions_reconciliation_unsupported")
	}
	if len(result.Mismatches) > 0 {
		result.Events = append(result.Events, "POSITION_MISMATCH")
	}
	if result.Status == "" || result.Status == "unknown" {
		result.Status = normalizeOrderStatus(request.Result.Status)
	}
	return result, nil
}

func normalizeExchangePositions(positions []exchanges.Position) []PositionSnapshot {
	out := make([]PositionSnapshot, 0, len(positions))
	for _, position := range positions {
		out = append(out, PositionSnapshot{
			Venue:         strings.ToLower(strings.TrimSpace(position.Venue)),
			Symbol:        strings.ToUpper(strings.TrimSpace(position.Symbol)),
			Side:          strings.ToUpper(strings.TrimSpace(position.Side)),
			Quantity:      parseFloat(position.Size),
			Entry:         parseFloat(position.Entry),
			UnrealizedPnL: parseFloat(position.PnL),
			Leverage:      parseFloat(position.Lev),
			Raw:           position,
		})
	}
	return out
}

func matchingPosition(positions []PositionSnapshot, venue string, symbol string) bool {
	for _, position := range positions {
		if strings.EqualFold(position.Venue, venue) && strings.EqualFold(position.Symbol, symbol) && position.Quantity != 0 {
			return true
		}
	}
	return false
}

func mergeOrderSnapshots(base []OrderSnapshot, extra []OrderSnapshot) []OrderSnapshot {
	seen := map[string]bool{}
	out := make([]OrderSnapshot, 0, len(base)+len(extra))
	for _, row := range append(base, extra...) {
		key := strings.ToLower(row.Venue) + ":" + strings.ToUpper(row.Symbol) + ":" + row.OrderID + ":" + row.ClientID
		if seen[key] {
			continue
		}
		seen[key] = true
		out = append(out, row)
	}
	return out
}

func normalizeOrderStatus(status string) string {
	status = strings.ToUpper(strings.TrimSpace(status))
	switch status {
	case "", "OK":
		return "ACKNOWLEDGED"
	case "FILLED":
		return "FILLED"
	case "PARTIALLY_FILLED", "PARTIAL":
		return "PARTIAL_FILL"
	case "CANCELED", "CANCELLED":
		return "CANCELLED"
	case "REJECTED", "ERROR":
		return "REJECTED"
	default:
		return status
	}
}

func parseFloat(value string) float64 {
	parsed, _ := strconv.ParseFloat(strings.TrimSpace(value), 64)
	return parsed
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}

func MarshalReconcileResult(result ReconcileResult) ([]byte, error) {
	return json.MarshalIndent(result, "", "  ")
}

func orderSide(side string) execution.OrderSide {
	switch strings.ToUpper(strings.TrimSpace(side)) {
	case "SHORT", "SELL":
		return execution.Sell
	default:
		return execution.Buy
	}
}

func envBool(name string) bool {
	return strings.EqualFold(strings.TrimSpace(os.Getenv(name)), "true") || strings.TrimSpace(os.Getenv(name)) == "1"
}

func envDefaultBool(name string, fallback bool) bool {
	value := strings.TrimSpace(os.Getenv(name))
	if value == "" {
		return fallback
	}
	return envBool(name) || strings.EqualFold(value, "y") || strings.EqualFold(value, "yes")
}

func allowedSet(value string) map[string]bool {
	out := map[string]bool{}
	for _, part := range strings.Split(value, ",") {
		trimmed := strings.ToLower(strings.TrimSpace(part))
		if trimmed != "" {
			out[trimmed] = true
		}
	}
	return out
}
