package hyperliquid

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestNormalizeHyperliquidTrade(t *testing.T) {
	print, err := NormalizeHyperliquidTrade(HyperliquidRawTrade{
		Coin: "BTC",
		Side: "B",
		Px:   "100.5",
		Sz:   "0.25",
		Time: 123,
		Hash: "h1",
	})
	if err != nil {
		t.Fatalf("normalize: %v", err)
	}
	if print.Venue != "hyperliquid" || print.Symbol != "BTC" || print.Price != 100.5 || print.Size != 0.25 {
		t.Fatalf("unexpected print: %+v", print)
	}
	if print.Side != "B" || print.AggressorSide != "B" {
		t.Fatalf("unexpected side: %+v", print)
	}
}

func TestGetRecentTradesAt(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/info" {
			t.Fatalf("unexpected request %s %s", r.Method, r.URL.Path)
		}
		body, _ := io.ReadAll(r.Body)
		if !strings.Contains(string(body), `"recentTrades"`) {
			t.Fatalf("unexpected body: %s", string(body))
		}
		_, _ = w.Write([]byte(`[{"coin":"BTC","side":"A","px":"101","sz":"0.5","time":124,"hash":"h2"}]`))
	}))
	defer server.Close()

	trades, err := getRecentTradesAt(server.URL, "BTC")
	if err != nil {
		t.Fatalf("recent trades: %v", err)
	}
	if len(trades) != 1 || trades[0].AggressorSide != "A" {
		t.Fatalf("unexpected trades: %+v", trades)
	}
}
