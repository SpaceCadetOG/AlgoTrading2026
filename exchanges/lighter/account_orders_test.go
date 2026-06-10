package lighter

import "testing"

func TestParseActiveOrdersPreservesLifecycleFields(t *testing.T) {
	body := []byte(`{
		"orders": [{
			"market_id": 1,
			"order_index": 44,
			"client_order_index": 55,
			"status": "open",
			"side": "buy",
			"type": "stop-loss",
			"initial_base_amount": "10",
			"remaining_base_amount": "4",
			"filled_base_amount": "6",
			"price": "65000",
			"reduce_only": true
		}]
	}`)
	orders := parseActiveOrders(body)
	if len(orders) != 1 {
		t.Fatalf("orders=%d", len(orders))
	}
	order := orders[0]
	if order.MarketID != 1 || order.OrderIndex != 44 || order.ClientOrderIndex != 55 {
		t.Fatalf("unexpected ids: %+v", order)
	}
	if order.Type != "stop-loss" || !order.ReduceOnly || order.FilledBaseAmount != "6" {
		t.Fatalf("expected protection/fill fields preserved: %+v", order)
	}
}

func TestParsePositionsPreservesVenueReconcileFields(t *testing.T) {
	body := []byte(`{
		"positions": {
			"1": {
				"market_index": 1,
				"symbol": "BTC",
				"position": "0.25",
				"side": "long",
				"entry_price": "64000",
				"pnl": "12.5"
			}
		}
	}`)
	positions := parsePositions(body)
	if len(positions) != 1 {
		t.Fatalf("positions=%d", len(positions))
	}
	position := positions[0]
	if position.MarketID != 1 || position.Symbol != "BTC" || position.Side != "long" || position.Entry != "64000" || position.PnL != "12.5" {
		t.Fatalf("unexpected position: %+v", position)
	}
}

func TestParseTradesPreservesOrderAttributionFields(t *testing.T) {
	body := []byte(`{
		"trades": [{
			"trade_id_str": "abc",
			"market_id": 1,
			"ask_id_str": "44",
			"bid_id_str": "43",
			"ask_client_id_str": "144",
			"bid_client_id_str": "143",
			"type": "trade",
			"side": "buy",
			"size": "0.1",
			"price": "65010",
			"timestamp": 1770500000000
		}]
	}`)
	trades := parseTrades(body)
	if len(trades) != 1 {
		t.Fatalf("trades=%d", len(trades))
	}
	trade := trades[0]
	if trade.TradeID != "abc" || trade.AskID != "44" || trade.BidID != "43" || trade.AskClientID != "144" || trade.BidClientID != "143" || trade.Size != "0.1" || trade.Price != "65010" {
		t.Fatalf("unexpected trade: %+v", trade)
	}
}

func TestParseAccountWSStateFromDocumentedChannels(t *testing.T) {
	body := []byte(`{
		"channel": "account_all_orders:331",
		"type": "update/account_all_orders",
		"orders": {
			"1": [{
				"order_index": 44,
				"client_order_index": 55,
				"market_index": 1,
				"initial_base_amount": "1",
				"remaining_base_amount": "0.5",
				"filled_base_amount": "0.5",
				"side": "buy",
				"type": "take-profit",
				"reduce_only": true,
				"status": "open"
			}]
		},
		"trades": {
			"1": [{
				"trade_id_str": "t1",
				"market_id": 1,
				"ask_id_str": "44",
				"bid_id_str": "43",
				"size": "0.5",
				"price": "65000",
				"timestamp": 1770500000000
			}]
		},
		"positions": {
			"1": [{
				"market_id": 1,
				"symbol": "BTC",
				"position": "0.5",
				"avg_entry_price": "64000",
				"unrealized_pnl": "5"
			}]
		}
	}`)
	state, err := ParseAccountWSState(body)
	if err != nil {
		t.Fatalf("parse ws state: %v", err)
	}
	if len(state.Orders) != 1 || len(state.Trades) != 1 || len(state.Positions) != 1 {
		t.Fatalf("unexpected ws state: %+v", state)
	}
	if !state.Orders[0].ReduceOnly || state.Orders[0].Type != "take-profit" {
		t.Fatalf("expected documented protection order fields: %+v", state.Orders[0])
	}
}
