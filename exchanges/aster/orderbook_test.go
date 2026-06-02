package aster

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"AlgoTrading2026/orderbook"
)

func TestGetOrderBookAt(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/fapi/v1/depth" {
			t.Fatalf("unexpected request %s %s", r.Method, r.URL.Path)
		}
		if r.URL.Query().Get("symbol") != "BTCUSDT" || r.URL.Query().Get("limit") != "100" {
			t.Fatalf("unexpected query: %s", r.URL.RawQuery)
		}
		_, _ = w.Write([]byte(`{
			"lastUpdateId": 10,
			"E": 123,
			"T": 122,
			"bids": [["100", "2"], ["99", "3"]],
			"asks": [["101", "4"], ["102", "5"]]
		}`))
	}))
	defer server.Close()

	snapshot, err := getOrderBookAt(server.URL, "BTCUSDT")
	if err != nil {
		t.Fatalf("orderbook: %v", err)
	}
	if snapshot.Venue != "aster" || snapshot.Symbol != "BTCUSDT" || !snapshot.IsSnapshot {
		t.Fatalf("unexpected snapshot metadata: %+v", snapshot)
	}
	if got := orderbook.BestBid(snapshot).Price; got != "100" {
		t.Fatalf("best bid=%s want 100", got)
	}
	if got := orderbook.BestAsk(snapshot).Price; got != "101" {
		t.Fatalf("best ask=%s want 101", got)
	}
}

func TestGetOrderBookAtRejectsInvalidSnapshot(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{
			"lastUpdateId": 10,
			"bids": [["102", "2"]],
			"asks": [["101", "4"]]
		}`))
	}))
	defer server.Close()

	if _, err := getOrderBookAt(server.URL, "BTCUSDT"); err == nil {
		t.Fatal("expected validation error")
	}
}
