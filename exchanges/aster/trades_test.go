package aster

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNormalizeAsterTrade(t *testing.T) {
	print, err := NormalizeAsterTrade("BTCUSDT", AsterRawTrade{
		ID:           123,
		Price:        "100.5",
		Qty:          "0.25",
		Time:         456,
		IsBuyerMaker: true,
	})
	if err != nil {
		t.Fatalf("normalize: %v", err)
	}
	if print.Venue != "aster" || print.Symbol != "BTCUSDT" || print.Price != 100.5 || print.Size != 0.25 {
		t.Fatalf("unexpected print: %+v", print)
	}
	if print.Side != "A" || print.AggressorSide != "A" || print.TradeID != "123" {
		t.Fatalf("unexpected side/id: %+v", print)
	}
}

func TestGetRecentTradesAt(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/fapi/v1/trades" {
			t.Fatalf("unexpected request %s %s", r.Method, r.URL.Path)
		}
		if r.URL.Query().Get("symbol") != "BTCUSDT" || r.URL.Query().Get("limit") != "2" {
			t.Fatalf("unexpected query: %s", r.URL.RawQuery)
		}
		_, _ = w.Write([]byte(`[{"id":1,"price":"101","qty":"0.5","time":124,"isBuyerMaker":false}]`))
	}))
	defer server.Close()

	trades, err := getRecentTradesAt(server.URL, "BTCUSDT", 2)
	if err != nil {
		t.Fatalf("recent trades: %v", err)
	}
	if len(trades) != 1 || trades[0].AggressorSide != "B" {
		t.Fatalf("unexpected trades: %+v", trades)
	}
}
