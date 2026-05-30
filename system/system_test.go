package system

import "testing"

func TestTradingSimulationBookFlowOrderLifecycleAndPnL(t *testing.T) {
	sim := NewTestTradingSimulation(16)
	summary := sim.RunArbitrageExample()

	if summary.BookBestBid != 100 {
		t.Fatalf("best bid = %f, want 100", summary.BookBestBid)
	}
	if summary.BookBestAsk != 99 {
		t.Fatalf("best ask = %f, want 99", summary.BookBestAsk)
	}
	if summary.Accepted != 2 {
		t.Fatalf("accepted open orders before fill = %d, want 2", summary.Accepted)
	}
	if summary.OpenAfterFill != 0 {
		t.Fatalf("open orders after fill = %d, want 0", summary.OpenAfterFill)
	}
	if summary.OrderStatus != OrderFilled {
		t.Fatalf("last order status = %s, want %s", summary.OrderStatus, OrderFilled)
	}
	if summary.Fills != 2 {
		t.Fatalf("fills = %d, want 2", summary.Fills)
	}
	if summary.RealizedPnL != 1 {
		t.Fatalf("realized pnl = %f, want 1", summary.RealizedPnL)
	}
	if summary.AuditEvents != 4 {
		t.Fatalf("audit events = %d, want 4", summary.AuditEvents)
	}
}

func TestOrderBookSortsBidsAsksAndEmitsOnlyTopChanges(t *testing.T) {
	queues := NewQueues(8)
	lp := NewLiquidityProvider(queues)
	book := NewOrderBook(queues)

	lp.InsertBid("bid-1", 10, 1)
	lp.InsertBid("bid-2", 9, 1)
	lp.InsertBid("bid-3", 10, 2)
	lp.InsertAsk("ask-1", 12, 1)
	lp.InsertAsk("ask-2", 11, 1)

	if !book.ProcessNext() {
		t.Fatal("expected first bid")
	}
	if !book.ProcessNext() {
		t.Fatal("expected lower bid")
	}
	if len(queues.OB2TS) != 1 {
		t.Fatalf("events after non-top bid = %d, want 1", len(queues.OB2TS))
	}
	if !book.ProcessNext() {
		t.Fatal("expected same-price bid")
	}
	if len(queues.OB2TS) != 1 {
		t.Fatalf("events after FIFO same-price bid = %d, want 1", len(queues.OB2TS))
	}
	if !book.ProcessNext() || !book.ProcessNext() {
		t.Fatal("expected asks")
	}

	bids := book.Bids()
	if bids[0].ID != "bid-1" || bids[1].ID != "bid-3" || bids[2].ID != "bid-2" {
		t.Fatalf("bids not sorted/FIFO: %+v", bids)
	}
	asks := book.Asks()
	if asks[0].ID != "ask-2" || asks[1].ID != "ask-1" {
		t.Fatalf("asks not sorted: %+v", asks)
	}
}

func TestCrossedBookCreatesBuyAndSellOrders(t *testing.T) {
	queues := NewQueues(8)
	lp := NewLiquidityProvider(queues)
	book := NewOrderBook(queues)
	strategy := NewTradingStrategy(queues)

	lp.InsertAsk("ask", 99, 2)
	lp.InsertBid("bid", 100, 1)
	book.ProcessNext()
	strategy.ProcessBookEvent()
	book.ProcessNext()
	strategy.ProcessBookEvent()

	if len(queues.TS2OM) != 2 {
		t.Fatalf("strategy orders = %d, want 2", len(queues.TS2OM))
	}
	buy := <-queues.TS2OM
	sell := <-queues.TS2OM
	if buy.Side != Buy || buy.Price != 99 {
		t.Fatalf("buy order = %+v, want buy at 99", buy)
	}
	if sell.Side != Sell || sell.Price != 100 {
		t.Fatalf("sell order = %+v, want sell at 100", sell)
	}
}

func TestOrderManagerAssignsIDsAndAcceptThenFillLifecycle(t *testing.T) {
	queues := NewQueues(8)
	manager := NewOrderManager(queues)
	simulator := NewMarketSimulator(queues)

	queues.TS2OM <- OrderIntent{ClientID: "client-1", Action: ActionNew, Side: Buy, Price: 10, Size: 1}
	if !manager.ProcessStrategyOrder() {
		t.Fatal("expected manager to process strategy order")
	}
	forwarded := <-queues.OM2GW
	if forwarded.ID != "om-1" {
		t.Fatalf("forwarded id = %s, want om-1", forwarded.ID)
	}
	if manager.Orders["om-1"].Status != OrderNew {
		t.Fatalf("manager state = %s, want NEW", manager.Orders["om-1"].Status)
	}

	queues.OM2GW <- forwarded
	if !simulator.ProcessNext() {
		t.Fatal("expected simulator accept")
	}
	manager.ProcessMarketResponse()
	if manager.Orders["om-1"].Status != OrderAccepted {
		t.Fatalf("manager state = %s, want ACCEPTED", manager.Orders["om-1"].Status)
	}

	if simulator.FillAllOrders() != 1 {
		t.Fatal("expected one fill")
	}
	manager.ProcessMarketResponse()
	if manager.Orders["om-1"].Status != OrderFilled {
		t.Fatalf("manager state = %s, want FILLED", manager.Orders["om-1"].Status)
	}
}

func TestMarketSimulatorCancelAndAmend(t *testing.T) {
	queues := NewQueues(8)
	simulator := NewMarketSimulator(queues)

	queues.OM2GW <- OrderIntent{ID: "om-1", Action: ActionNew, Side: Buy, Price: 10, Size: 1}
	simulator.ProcessNext()
	<-queues.GW2OM

	queues.OM2GW <- OrderIntent{ID: "om-1", Action: ActionAmend, Price: 11, Size: 2}
	simulator.ProcessNext()
	amend := <-queues.GW2OM
	if amend.Status != OrderAmended {
		t.Fatalf("amend status = %s, want AMENDED", amend.Status)
	}
	if simulator.orders["om-1"].Price != 11 || simulator.orders["om-1"].Size != 2 {
		t.Fatalf("amended order = %+v", simulator.orders["om-1"])
	}

	queues.OM2GW <- OrderIntent{ID: "om-1", Action: ActionCancel}
	simulator.ProcessNext()
	cancel := <-queues.GW2OM
	if cancel.Status != OrderCanceled {
		t.Fatalf("cancel status = %s, want CANCELED", cancel.Status)
	}
	if len(simulator.orders) != 0 {
		t.Fatalf("open simulator orders = %d, want 0", len(simulator.orders))
	}
}

func TestMarketSimulatorRejectsDuplicateAndInvalidOrders(t *testing.T) {
	queues := NewQueues(8)
	simulator := NewMarketSimulator(queues)

	queues.OM2GW <- OrderIntent{ID: "om-1", Action: ActionNew, Side: Buy, Price: 10, Size: 1}
	simulator.ProcessNext()
	<-queues.GW2OM

	queues.OM2GW <- OrderIntent{ID: "om-1", Action: ActionNew, Side: Buy, Price: 10, Size: 1}
	simulator.ProcessNext()
	duplicate := <-queues.GW2OM
	if duplicate.Status != OrderRejected {
		t.Fatalf("duplicate status = %s, want REJECTED", duplicate.Status)
	}

	queues.OM2GW <- OrderIntent{ID: "missing", Action: ActionCancel}
	simulator.ProcessNext()
	missing := <-queues.GW2OM
	if missing.Status != OrderRejected {
		t.Fatalf("missing cancel status = %s, want REJECTED", missing.Status)
	}
}
