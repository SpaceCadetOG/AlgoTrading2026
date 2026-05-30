package system

import (
	"testing"

	"AlgoTrading2026/exchanges"
	"AlgoTrading2026/execution"
)

func TestVenueGatewayStartsDisabledByDefault(t *testing.T) {
	gateway := NewVenueGateway(VenueGatewayConfig{Venue: "aster"})
	if err := gateway.Start(); err != nil {
		t.Fatalf("start disabled gateway: %v", err)
	}
	if gateway.Status() != GatewayDisabled {
		t.Fatalf("expected disabled status, got %s", gateway.Status())
	}
}

func TestVenueGatewaySendOrderBlockedWhenLiveOrdersFalse(t *testing.T) {
	gateway := NewVenueGateway(VenueGatewayConfig{
		Venue:   "hyperliquid",
		Enabled: true,
		DryRun:  true,
	})
	if err := gateway.Start(); err != nil {
		t.Fatalf("start gateway: %v", err)
	}

	err := gateway.SendOrder(OrderIntent{
		ID:       "blocked-1",
		ClientID: "client-1",
		Action:   ActionNew,
		Side:     Buy,
		Price:    100,
		Size:     1,
	})
	if err != nil {
		t.Fatalf("send blocked order: %v", err)
	}

	response, ok := gateway.ReceiveOrderResponse()
	if !ok {
		t.Fatal("expected blocked response")
	}
	if response.Status != OrderRejected || response.Reason != "live orders disabled" {
		t.Fatalf("unexpected blocked response: %+v", response)
	}
}

func TestVenueCandleTranslators(t *testing.T) {
	candle := exchanges.Candle{
		Symbol:    "BTCUSDT",
		Close:     "100.25",
		Volume:    "4.5",
		StartTime: 123,
	}

	cases := []struct {
		name       string
		translate  func(exchanges.Candle) GatewayPriceUpdate
		wantVenue  string
		wantSymbol string
	}{
		{"aster", AsterCandleToGatewayPriceUpdate, "aster", "BTCUSDT"},
		{"hyperliquid", HyperliquidCandleToGatewayPriceUpdate, "hyperliquid", "BTCUSDT"},
		{"lighter", LighterCandleToGatewayPriceUpdate, "lighter", "BTCUSDT"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := tc.translate(candle)
			if got.Venue != tc.wantVenue || got.Symbol != tc.wantSymbol {
				t.Fatalf("unexpected translated candle: %+v", got)
			}
			if got.Bid != 100.25 || got.Ask != 100.25 || got.BidSize != 4.5 || got.AskSize != 4.5 {
				t.Fatalf("unexpected bid/ask translation: %+v", got)
			}
			if len(got.ToLiquidityEvents()) != 2 {
				t.Fatalf("expected two liquidity events, got %+v", got.ToLiquidityEvents())
			}
		})
	}
}

func TestVenueOrderResultTranslators(t *testing.T) {
	result := execution.OrderResult{
		Success: true,
		Symbol:  "BTC",
		OrderID: "42",
		Status:  "filled",
		Message: "done",
	}

	cases := []struct {
		name      string
		translate func(execution.OrderResult) GatewayOrderResponse
		wantVenue string
	}{
		{"aster", AsterOrderResultToGatewayOrderResponse, "aster"},
		{"hyperliquid", HyperliquidOrderResultToGatewayOrderResponse, "hyperliquid"},
		{"lighter", LighterOrderResultToGatewayOrderResponse, "lighter"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := tc.translate(result)
			if got.Venue != tc.wantVenue || got.Symbol != "BTC" {
				t.Fatalf("unexpected translated order venue/symbol: %+v", got)
			}
			if got.Response.OrderID != "42" || got.Response.Status != OrderFilled {
				t.Fatalf("unexpected translated order response: %+v", got)
			}
		})
	}
}
