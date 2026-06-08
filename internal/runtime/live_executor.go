package runtime

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"AlgoTrading2026/execution"
)

type LiveGate struct {
	EnableLiveTrading bool
	VenueHealthy      bool
	AccountReady      bool
	AllowedVenues     map[string]bool
}

func LiveGateFromEnv() LiveGate {
	return LiveGate{
		EnableLiveTrading: envBool("LIVE_ENABLE_LIVE_TRADING") || envBool("ENABLE_LIVE_ORDERS"),
		VenueHealthy:      envDefaultBool("LIVE_VENUE_HEALTHY", true),
		AccountReady:      envDefaultBool("LIVE_ACCOUNT_READY", false),
		AllowedVenues:     allowedSet(os.Getenv("LIVE_ALLOWED_VENUES")),
	}
}

func (g LiveGate) Refusal(candidate Candidate) string {
	if !g.EnableLiveTrading {
		return "live_trading_not_enabled"
	}
	if !g.VenueHealthy {
		return "venue_not_healthy"
	}
	if !g.AccountReady {
		return "account_not_ready"
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
	return ExecutionResult{
		Accepted: result.Success,
		Venue:    result.Venue,
		Symbol:   result.Symbol,
		OrderID:  result.OrderID,
		Status:   result.Status,
		Message:  result.Message,
		Raw:      result.Raw,
	}, nil
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
