package lighter

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"AlgoTrading2026/tradetape"
)

func TestGetRecentTradesAtParsesTrades(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/api/v1/recentTrades" {
			t.Fatalf("unexpected request %s %s", r.Method, r.URL.Path)
		}
		if r.URL.Query().Get("market_id") != "1" || r.URL.Query().Get("limit") != "2" {
			t.Fatalf("unexpected query: %s", r.URL.RawQuery)
		}
		_, _ = w.Write([]byte(`{
			"recent_trades": [
				{"market_id":1,"trade_id":"t1","timestamp":123,"price":"100.5","size":"0.25","side":"buy"},
				{"market_id":1,"trade_id":"t2","timestamp":124,"price":"101.5","size":"0.50","side":"sell"}
			]
		}`))
	}))
	defer server.Close()

	trades, err := getRecentTradesAt(server.URL, 1, "BTC", 2)
	if err != nil {
		t.Fatalf("recent trades: %v", err)
	}
	if len(trades) != 2 {
		t.Fatalf("len=%d want 2", len(trades))
	}
	if trades[0].Venue != "lighter" || trades[0].Symbol != "BTC" || trades[0].MarketID != "1" {
		t.Fatalf("unexpected normalized metadata: %+v", trades[0])
	}
	if trades[0].Price != 100.5 || trades[0].Size != 0.25 || trades[0].Timestamp != 123 {
		t.Fatalf("unexpected normalized values: %+v", trades[0])
	}
	if trades[0].Side != "BUY" || trades[0].AggressorSide != tradetape.UnknownAggressorSide {
		t.Fatalf("unexpected side handling: %+v", trades[0])
	}
}

func TestGetTradesAtParsesDataArray(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/trades" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		if r.URL.Query().Get("sort_by") != "timestamp" || r.URL.Query().Get("type") != "trade" {
			t.Fatalf("unexpected query: %s", r.URL.RawQuery)
		}
		_, _ = w.Write([]byte(`{
			"data": [
				{"market_index":"1","id":"abc","time":"456","px":"200","qty":"1.25","direction":"ask"}
			]
		}`))
	}))
	defer server.Close()

	trades, err := getTradesAt(server.URL, 1, "BTC", 1)
	if err != nil {
		t.Fatalf("trades: %v", err)
	}
	if len(trades) != 1 {
		t.Fatalf("len=%d want 1", len(trades))
	}
	if trades[0].TradeID != "abc" || trades[0].Side != "ASK" || trades[0].AggressorSide != tradetape.UnknownAggressorSide {
		t.Fatalf("unexpected trade: %+v", trades[0])
	}
}

func TestNormalizeLighterTradeMissingSideUsesUnknownAggressor(t *testing.T) {
	print, err := NormalizeLighterTrade(LighterRawTrade{
		"symbol":    "BTC",
		"market_id": "1",
		"timestamp": float64(123),
		"price":     "100",
		"size":      "0.1",
	})
	if err != nil {
		t.Fatalf("normalize: %v", err)
	}
	if print.Side != "" || print.AggressorSide != tradetape.UnknownAggressorSide {
		t.Fatalf("unexpected side fields: %+v", print)
	}
}

func TestNormalizeLighterTradeInvalidPriceReturnsError(t *testing.T) {
	_, err := NormalizeLighterTrade(LighterRawTrade{
		"symbol":    "BTC",
		"market_id": "1",
		"timestamp": float64(123),
		"price":     "0",
		"size":      "0.1",
	})
	if err == nil {
		t.Fatal("expected invalid price error")
	}
}

func TestLighterMarketIDSymbolMapping(t *testing.T) {
	marketID, symbol, err := lighterMarketID("BTCUSDT")
	if err != nil {
		t.Fatalf("market id: %v", err)
	}
	if marketID != 1 || symbol != "BTC" {
		t.Fatalf("marketID=%d symbol=%s want 1 BTC", marketID, symbol)
	}
}

func TestDecodeLighterTradesEmptyResponse(t *testing.T) {
	trades, err := decodeLighterTrades([]byte(`{"trades":[]}`), 1, "BTC")
	if err != nil {
		t.Fatalf("empty response: %v", err)
	}
	if len(trades) != 0 {
		t.Fatalf("len=%d want 0", len(trades))
	}
}
