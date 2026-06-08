package aster

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNormalizeAsterAggTradeSideMapping(t *testing.T) {
	sell, err := NormalizeAsterAggTrade("BTCUSDT", AsterAggTrade{
		AggregateTradeID: 1, Price: "100.5", Quantity: "0.25", Timestamp: 123, BuyerMaker: true,
	})
	if err != nil {
		t.Fatalf("normalize sell: %v", err)
	}
	if sell.Side != "SELL" || sell.AggressorSide != "SELL" || sell.CanonicalSymbol != "BTC" {
		t.Fatalf("unexpected sell print: %+v", sell)
	}
	buy, err := NormalizeAsterAggTrade("BTCUSDT", AsterAggTrade{
		AggregateTradeID: 2, Price: "101.5", Quantity: "0.5", Timestamp: 124, BuyerMaker: false,
	})
	if err != nil {
		t.Fatalf("normalize buy: %v", err)
	}
	if buy.Side != "BUY" || buy.AggressorSide != "BUY" || buy.Price != 101.5 || buy.Size != 0.5 {
		t.Fatalf("unexpected buy print: %+v", buy)
	}
}

func TestNormalizeAsterAggTradeInvalidPrice(t *testing.T) {
	if _, err := NormalizeAsterAggTrade("BTCUSDT", AsterAggTrade{
		AggregateTradeID: 1, Price: "bad", Quantity: "1", Timestamp: 1,
	}); err == nil {
		t.Fatal("expected invalid price error")
	}
}

func TestBackfillAsterAggTradeWindows(t *testing.T) {
	windows := BackfillAsterAggTradeWindows(0, int64(24*60*60*1000))
	if len(windows) != 0 {
		t.Fatalf("zero start should produce no windows: %+v", windows)
	}
	windows = BackfillAsterAggTradeWindows(1, 1+int64(24*60*60*1000))
	if len(windows) != 25 {
		t.Fatalf("windows=%d want 25 for inclusive 24h+1ms range", len(windows))
	}
	for _, window := range windows {
		if window[1]-window[0] >= int64(60*60*1000) {
			t.Fatalf("window is not sub-1h: %+v", window)
		}
	}
}

func TestBackfillAsterAggTradesAt(t *testing.T) {
	var calls int
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		if r.Method != http.MethodGet || r.URL.Path != asterAggTradesEndpoint {
			t.Fatalf("unexpected request %s %s", r.Method, r.URL.Path)
		}
		if r.URL.Query().Get("symbol") != "BTCUSDT" || r.URL.Query().Get("limit") != "1000" {
			t.Fatalf("unexpected query: %s", r.URL.RawQuery)
		}
		_, _ = w.Write([]byte(`[{"a":2,"p":"101","q":"0.2","f":1,"l":2,"T":2000,"m":false},{"a":1,"p":"100","q":"0.1","f":1,"l":1,"T":1000,"m":true}]`))
	}))
	defer server.Close()
	prints, err := backfillAsterAggTradesAt(server.URL, "BTCUSDT", 1000, 2000)
	if err != nil {
		t.Fatalf("backfill: %v", err)
	}
	if calls != 1 || len(prints) != 2 {
		t.Fatalf("unexpected calls/prints: calls=%d prints=%+v", calls, prints)
	}
	if prints[0].TradeID != "1" || prints[1].TradeID != "2" {
		t.Fatalf("prints not sorted: %+v", prints)
	}
}
