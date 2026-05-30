package system

import "testing"

func TestSimulatedGatewayStartStop(t *testing.T) {
	gateway := NewSimulatedGateway(NewQueues(8))
	if gateway.Status() != GatewayCreated {
		t.Fatalf("expected created status, got %s", gateway.Status())
	}
	if err := gateway.Start(); err != nil {
		t.Fatalf("start gateway: %v", err)
	}
	if gateway.Status() != GatewayRunning {
		t.Fatalf("expected running status, got %s", gateway.Status())
	}
	if err := gateway.Stop(); err != nil {
		t.Fatalf("stop gateway: %v", err)
	}
	if gateway.Status() != GatewayStopped {
		t.Fatalf("expected stopped status, got %s", gateway.Status())
	}
}

func TestSimulatedGatewayReceivesPriceUpdate(t *testing.T) {
	queues := NewQueues(8)
	gateway := NewSimulatedGateway(queues)
	if err := gateway.Start(); err != nil {
		t.Fatalf("start gateway: %v", err)
	}

	err := gateway.InjectPriceUpdate(GatewayPriceUpdate{
		Venue:   "sim",
		Symbol:  "BTC",
		Bid:     100,
		BidSize: 1,
		Ask:     101,
		AskSize: 2,
	})
	if err != nil {
		t.Fatalf("inject price update: %v", err)
	}

	bid, ok := gateway.ReceivePriceUpdate()
	if !ok {
		t.Fatal("expected bid liquidity event")
	}
	if bid.Side != LiquidityBid || bid.Price != 100 || bid.Size != 1 {
		t.Fatalf("unexpected bid event: %+v", bid)
	}
	ask, ok := gateway.ReceivePriceUpdate()
	if !ok {
		t.Fatal("expected ask liquidity event")
	}
	if ask.Side != LiquidityAsk || ask.Price != 101 || ask.Size != 2 {
		t.Fatalf("unexpected ask event: %+v", ask)
	}

	select {
	case event := <-queues.LP2Gateway:
		if event.Side != LiquidityBid {
			t.Fatalf("expected gateway to publish bid first, got %+v", event)
		}
	default:
		t.Fatal("expected price update to fit Chapter 7 LP2Gateway path")
	}
}

func TestSimulatedGatewaySendsOrderAndReceivesAcceptedFill(t *testing.T) {
	gateway := NewSimulatedGateway(NewQueues(8))
	if err := gateway.Start(); err != nil {
		t.Fatalf("start gateway: %v", err)
	}

	order := OrderIntent{
		ID:       "sim-1",
		ClientID: "client-1",
		Action:   ActionNew,
		Side:     Buy,
		Price:    100,
		Size:     1,
	}
	if err := gateway.SendOrder(order); err != nil {
		t.Fatalf("send order: %v", err)
	}
	if !gateway.ProcessNextOrder() {
		t.Fatal("expected simulator to process order")
	}
	accepted, ok := gateway.ReceiveOrderResponse()
	if !ok {
		t.Fatal("expected accept response")
	}
	if accepted.Status != OrderAccepted {
		t.Fatalf("expected accepted response, got %+v", accepted)
	}

	if filled := gateway.FillAllOrders(); filled != 1 {
		t.Fatalf("expected one fill, got %d", filled)
	}
	fill, ok := gateway.ReceiveOrderResponse()
	if !ok {
		t.Fatal("expected fill response")
	}
	if fill.Status != OrderFilled || fill.FillPrice != order.Price || fill.FillSize != order.Size {
		t.Fatalf("unexpected fill response: %+v", fill)
	}
}

func TestSimulatedGatewayFitsChapter7OrderManagerPath(t *testing.T) {
	queues := NewQueues(8)
	manager := NewOrderManager(queues)
	gateway := NewSimulatedGateway(queues)
	if err := gateway.Start(); err != nil {
		t.Fatalf("start gateway: %v", err)
	}

	queues.TS2OM <- OrderIntent{
		ClientID: "strategy-1",
		Action:   ActionNew,
		Side:     Sell,
		Price:    105,
		Size:     2,
	}
	if !manager.ProcessStrategyOrder() {
		t.Fatal("expected order manager to process strategy order")
	}
	if !gateway.ProcessNextOrder() {
		t.Fatal("expected gateway simulator to process forwarded order")
	}
	if !manager.ProcessMarketResponse() {
		t.Fatal("expected order manager to consume accept response")
	}
	accepted := <-queues.OM2TS
	if accepted.Status != OrderAccepted {
		t.Fatalf("expected accepted strategy response, got %+v", accepted)
	}

	if filled := gateway.FillAllOrders(); filled != 1 {
		t.Fatalf("expected one filled order, got %d", filled)
	}
	if !manager.ProcessMarketResponse() {
		t.Fatal("expected order manager to consume fill response")
	}
	filled := <-queues.OM2TS
	if filled.Status != OrderFilled {
		t.Fatalf("expected filled strategy response, got %+v", filled)
	}
}
