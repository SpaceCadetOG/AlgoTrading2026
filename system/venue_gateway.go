package system

import (
	"fmt"
	"strings"

	"AlgoTrading2026/exchanges"
	"AlgoTrading2026/execution"
)

type VenueGatewayConfig struct {
	Venue           string
	Enabled         bool
	DryRun          bool
	AllowLiveOrders bool
}

type VenueGateway struct {
	config    VenueGatewayConfig
	status    GatewayStatus
	responses chan OrderResponse
}

func NewVenueGateway(config VenueGatewayConfig) *VenueGateway {
	if config.Venue == "" {
		config.Venue = "unknown"
	}
	return &VenueGateway{
		config:    config,
		status:    GatewayCreated,
		responses: make(chan OrderResponse, 16),
	}
}

func (g *VenueGateway) Start() error {
	if !g.config.Enabled {
		g.status = GatewayDisabled
		return nil
	}
	g.status = GatewayRunning
	return nil
}

func (g *VenueGateway) Stop() error {
	g.status = GatewayStopped
	return nil
}

func (g *VenueGateway) Status() GatewayStatus {
	return g.status
}

func (g *VenueGateway) ReceivePriceUpdate() (LiquidityEvent, bool) {
	return LiquidityEvent{}, false
}

func (g *VenueGateway) SendOrder(order OrderIntent) error {
	if !g.config.Enabled {
		g.responses <- blockedVenueOrder(order, "venue gateway disabled")
		return nil
	}
	if !g.config.AllowLiveOrders {
		g.responses <- blockedVenueOrder(order, "live orders disabled")
		return nil
	}
	if g.config.DryRun {
		g.responses <- blockedVenueOrder(order, "dry run venue gateway")
		return nil
	}

	g.responses <- blockedVenueOrder(order, "real venue adapter execution not wired")
	return nil
}

func (g *VenueGateway) ReceiveOrderResponse() (OrderResponse, bool) {
	select {
	case response := <-g.responses:
		return response, true
	default:
		return OrderResponse{}, false
	}
}

func AsterCandleToGatewayPriceUpdate(candle exchanges.Candle) GatewayPriceUpdate {
	return candleToGatewayPriceUpdate("aster", candle)
}

func HyperliquidCandleToGatewayPriceUpdate(candle exchanges.Candle) GatewayPriceUpdate {
	return candleToGatewayPriceUpdate("hyperliquid", candle)
}

func LighterCandleToGatewayPriceUpdate(candle exchanges.Candle) GatewayPriceUpdate {
	return candleToGatewayPriceUpdate("lighter", candle)
}

func AsterOrderResultToGatewayOrderResponse(result execution.OrderResult) GatewayOrderResponse {
	return orderResultToGatewayOrderResponse("aster", result)
}

func HyperliquidOrderResultToGatewayOrderResponse(result execution.OrderResult) GatewayOrderResponse {
	return orderResultToGatewayOrderResponse("hyperliquid", result)
}

func LighterOrderResultToGatewayOrderResponse(result execution.OrderResult) GatewayOrderResponse {
	return orderResultToGatewayOrderResponse("lighter", result)
}

func candleToGatewayPriceUpdate(defaultVenue string, candle exchanges.Candle) GatewayPriceUpdate {
	venue := strings.TrimSpace(candle.Venue)
	if venue == "" {
		venue = defaultVenue
	}
	mid := candle.CloseFloat()
	return GatewayPriceUpdate{
		Venue:   venue,
		Symbol:  candle.Symbol,
		BidID:   fmt.Sprintf("%s:%s:%d:bid", venue, candle.Symbol, candle.StartTime),
		Bid:     mid,
		BidSize: candle.VolumeFloat(),
		AskID:   fmt.Sprintf("%s:%s:%d:ask", venue, candle.Symbol, candle.StartTime),
		Ask:     mid,
		AskSize: candle.VolumeFloat(),
	}
}

func orderResultToGatewayOrderResponse(defaultVenue string, result execution.OrderResult) GatewayOrderResponse {
	venue := strings.TrimSpace(result.Venue)
	if venue == "" {
		venue = defaultVenue
	}

	status := mapExecutionStatus(result)
	response := OrderResponse{
		OrderID:  result.OrderID,
		Status:   status,
		Reason:   result.Message,
		ClientID: result.OrderID,
	}
	if !result.Success && response.Reason == "" {
		response.Reason = "venue order failed"
	}

	return GatewayOrderResponse{
		Venue:    venue,
		Symbol:   result.Symbol,
		Response: response,
	}
}

func mapExecutionStatus(result execution.OrderResult) OrderStatus {
	status := strings.ToUpper(strings.TrimSpace(result.Status))
	switch status {
	case string(OrderAccepted), "NEW", "OPEN", "RESTING", "SUCCESS", "SUBMITTED":
		return OrderAccepted
	case string(OrderFilled):
		return OrderFilled
	case string(OrderCanceled), "CANCELLED":
		return OrderCanceled
	case string(OrderAmended):
		return OrderAmended
	case string(OrderRejected), "ERROR", "FAILED":
		return OrderRejected
	case "DRY_RUN", "SIGNED_DRY_RUN":
		return OrderRejected
	default:
		if result.Success {
			return OrderAccepted
		}
		return OrderRejected
	}
}

func blockedVenueOrder(order OrderIntent, reason string) OrderResponse {
	return OrderResponse{
		OrderID:  order.ID,
		ClientID: order.ClientID,
		Status:   OrderRejected,
		Side:     order.Side,
		Reason:   reason,
	}
}
