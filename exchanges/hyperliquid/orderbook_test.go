package hyperliquid

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"AlgoTrading2026/orderbook"
)

func TestGetL2OrderBookSnapshotAt(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/info" {
			t.Fatalf("unexpected request %s %s", r.Method, r.URL.Path)
		}
		var req map[string]string
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		if req["type"] != "l2Book" || req["coin"] != "BTC" {
			t.Fatalf("unexpected request body: %+v", req)
		}
		_, _ = w.Write([]byte(`{
			"coin":"BTC",
			"time":123,
			"levels":[
				[
					{"px":"100","sz":"2","n":1},
					{"px":"99","sz":"3","n":2}
				],
				[
					{"px":"101","sz":"4","n":1},
					{"px":"102","sz":"5","n":2}
				]
			]
		}`))
	}))
	defer server.Close()

	book, err := getL2OrderBookSnapshotAt(server.URL, "BTC")
	if err != nil {
		t.Fatalf("snapshot: %v", err)
	}
	if book.Venue != "hyperliquid" || book.Symbol != "BTC" || !book.IsSnapshot {
		t.Fatalf("unexpected book metadata: %+v", book)
	}
	if got := orderbook.BestBid(*book).Price; got != "100" {
		t.Fatalf("best bid=%s want 100", got)
	}
	if got := orderbook.BestAsk(*book).Price; got != "101" {
		t.Fatalf("best ask=%s want 101", got)
	}
}

func TestDecodeL2BookLegacyLevelsShape(t *testing.T) {
	raw := json.RawMessage(`[
		[{"px":"100","sz":"1","n":1}],
		[{"px":"101","sz":"2","n":1}]
	]`)
	book, err := decodeL2Book("BTC", raw)
	if err != nil {
		t.Fatalf("decode legacy: %v", err)
	}
	if err := orderbook.ValidateSnapshot(*book); err != nil {
		t.Fatalf("validate legacy: %v", err)
	}
}

func TestDecodeL2BookRejectsBadBook(t *testing.T) {
	raw := json.RawMessage(`{
		"coin":"BTC",
		"time":123,
		"levels":[
			[{"px":"102","sz":"1","n":1}],
			[{"px":"101","sz":"1","n":1}]
		]
	}`)
	book, err := decodeL2Book("BTC", raw)
	if err != nil {
		t.Fatalf("decode: %v", err)
	}
	if err := orderbook.ValidateSnapshot(*book); err == nil {
		t.Fatal("expected crossed book validation error")
	}
}
