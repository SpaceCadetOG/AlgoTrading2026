package lighter

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gorilla/websocket"

	"AlgoTrading2026/orderbook"
)

func TestGetOrderBookAt(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/api/v1/orderBook" {
			t.Fatalf("unexpected request %s %s", r.Method, r.URL.Path)
		}
		if r.URL.Query().Get("market_id") != "1" {
			t.Fatalf("unexpected query: %s", r.URL.RawQuery)
		}
		_, _ = w.Write([]byte(`{
			"timestamp": 123,
			"bids": [
				{"price":"100","size":"2"},
				{"price":"99","size":"3"}
			],
			"asks": [
				{"price":"101","size":"4"},
				{"price":"102","size":"5"}
			]
		}`))
	}))
	defer server.Close()

	snapshot, err := getOrderBookAt(server.URL, 1, "BTC")
	if err != nil {
		t.Fatalf("orderbook: %v", err)
	}
	if snapshot.Venue != "lighter" || snapshot.Symbol != "BTC" || !snapshot.IsSnapshot {
		t.Fatalf("unexpected snapshot metadata: %+v", snapshot)
	}
	if got := orderbook.BestBid(snapshot).Price; got != "100" {
		t.Fatalf("best bid=%s want 100", got)
	}
	if got := orderbook.BestAsk(snapshot).Price; got != "101" {
		t.Fatalf("best ask=%s want 101", got)
	}
}

func TestDecodeLighterOrderBookNestedArrays(t *testing.T) {
	snapshot, err := decodeLighterOrderBook([]byte(`{
		"order_book": {
			"time": "123",
			"bids": [["100","2"]],
			"asks": [["101","3"]]
		}
	}`), "BTC")
	if err != nil {
		t.Fatalf("decode: %v", err)
	}
	if err := orderbook.ValidateSnapshot(snapshot); err != nil {
		t.Fatalf("validate: %v", err)
	}
}

func TestGetOrderBookAtRejectsInvalidSnapshot(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{
			"bids": [{"price":"102","size":"2"}],
			"asks": [{"price":"101","size":"4"}]
		}`))
	}))
	defer server.Close()

	if _, err := getOrderBookAt(server.URL, 1, "BTC"); err == nil {
		t.Fatal("expected validation error")
	}
}

func TestGetOrderBookAtFallsBackToOrderBookOrders(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/v1/orderBook":
			http.NotFound(w, r)
		case "/api/v1/orderBookDetails":
			http.NotFound(w, r)
		case "/api/v1/orderBookOrders":
			if r.URL.Query().Get("market_id") != "1" || r.URL.Query().Get("limit") != "250" {
				t.Fatalf("unexpected fallback query: %s", r.URL.RawQuery)
			}
			_, _ = w.Write([]byte(`{
				"timestamp":123,
				"orders":[
					{"price":"100","remaining_base_amount":"1","is_ask":false},
					{"price":"100","remaining_base_amount":"2","is_ask":false},
					{"price":"101","remaining_base_amount":"3","is_ask":true}
				]
			}`))
		default:
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
	}))
	defer server.Close()

	snapshot, err := getOrderBookAt(server.URL, 1, "BTC")
	if err != nil {
		t.Fatalf("fallback orderbook: %v", err)
	}
	if got := orderbook.BestBid(snapshot); got.Price != "100" || got.Size != "3" {
		t.Fatalf("best bid=%+v want aggregated 100 x 3", got)
	}
	if got := orderbook.BestAsk(snapshot).Price; got != "101" {
		t.Fatalf("best ask=%s want 101", got)
	}
}

func TestGetOrderBookTriesDocumentedBTCMarketIDFallback(t *testing.T) {
	requestedIDs := make([]string, 0, 3)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/v1/orderBook" {
			http.NotFound(w, r)
			return
		}
		if r.URL.Path == "/api/v1/orderBookDetails" {
			http.NotFound(w, r)
			return
		}
		if r.URL.Path != "/api/v1/orderBookOrders" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		marketID := r.URL.Query().Get("market_id")
		requestedIDs = append(requestedIDs, marketID)
		if marketID == "1" {
			_, _ = w.Write([]byte(`{"orders":[{"price":"100","remaining_base_amount":"1","is_ask":false}]}`))
			return
		}
		_, _ = w.Write([]byte(`{
			"orders":[
				{"price":"100","remaining_base_amount":"1","is_ask":"0"},
				{"price":"101","remaining_base_amount":"2","is_ask":"1"}
			]
		}`))
	}))
	defer server.Close()

	var snapshot orderbook.OrderBookSnapshot
	var err error
	for _, candidateID := range lighterOrderBookMarketIDs("BTC", 1) {
		snapshot, err = getOrderBookAt(server.URL, candidateID, "BTC")
		if err == nil {
			break
		}
	}
	if err != nil {
		t.Fatalf("fallback ids: %v", err)
	}
	if snapshot.Symbol != "BTC" || orderbook.BestAsk(snapshot).Price != "101" {
		t.Fatalf("unexpected snapshot: %+v", snapshot)
	}
	if len(requestedIDs) < 2 || requestedIDs[0] != "1" || requestedIDs[1] != "2" {
		t.Fatalf("requested ids=%v want 1 then 2", requestedIDs)
	}
}

func TestGetOrderBookAtFallsBackToOrderBookDetails(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/v1/orderBook":
			http.NotFound(w, r)
		case "/api/v1/orderBookDetails":
			if r.URL.Query().Get("market_id") != "1" || r.URL.Query().Get("depth") != "100" {
				t.Fatalf("unexpected details query: %s", r.URL.RawQuery)
			}
			_, _ = w.Write([]byte(`{
				"order_book": {
					"time": 123,
					"bids": [{"price":"100","size":"2"}],
					"asks": [{"price":"101","size":"3"}]
				}
			}`))
		default:
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
	}))
	defer server.Close()

	snapshot, err := getOrderBookAt(server.URL, 1, "BTC")
	if err != nil {
		t.Fatalf("details fallback: %v", err)
	}
	if orderbook.BestBid(snapshot).Price != "100" || orderbook.BestAsk(snapshot).Price != "101" {
		t.Fatalf("unexpected details snapshot: %+v", snapshot)
	}
}

func TestGetOrderBookViaWSAt(t *testing.T) {
	upgrader := websocket.Upgrader{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			t.Fatalf("upgrade: %v", err)
		}
		defer conn.Close()

		_, msg, err := conn.ReadMessage()
		if err != nil {
			t.Fatalf("read subscribe: %v", err)
		}
		if string(msg) != `{"type":"subscribe","channel":"order_book/1"}` {
			t.Fatalf("unexpected subscribe: %s", string(msg))
		}

		err = conn.WriteMessage(websocket.TextMessage, []byte(`{
			"channel":"order_book:1",
			"timestamp":123,
			"type":"update/order_book",
			"order_book":{
				"code":0,
				"asks":[{"price":"101","size":"2"}],
				"bids":[{"price":"100","size":"3"}],
				"offset":10,
				"nonce":20,
				"last_updated_at":123,
				"begin_nonce":19
			}
		}`))
		if err != nil {
			t.Fatalf("write snapshot: %v", err)
		}
	}))
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")
	snapshot, err := getOrderBookViaWSAt(wsURL, 1, "BTC")
	if err != nil {
		t.Fatalf("ws orderbook: %v", err)
	}
	if got := orderbook.BestBid(snapshot).Price; got != "100" {
		t.Fatalf("best bid=%s want 100", got)
	}
	if got := orderbook.BestAsk(snapshot).Price; got != "101" {
		t.Fatalf("best ask=%s want 101", got)
	}
}
