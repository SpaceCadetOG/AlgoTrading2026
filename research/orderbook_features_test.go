package research

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"AlgoTrading2026/exchanges"
	"AlgoTrading2026/orderbook"
)

func TestBuildOrderBookFeatureRow(t *testing.T) {
	row := BuildOrderBookFeatureRow(testOrderBookSnapshot())
	if row.Venue != "hyperliquid" || row.Symbol != "BTC" {
		t.Fatalf("unexpected metadata: %+v", row)
	}
	if row.BestBid != 100 || row.BestAsk != 101 || row.Spread != 1 || !row.Valid {
		t.Fatalf("unexpected metrics: %+v", row)
	}
}

func TestBuildVWAPL2InteractionRow(t *testing.T) {
	candles := []exchanges.Candle{
		{High: "100", Low: "100", Close: "100", Volume: "1", StartTime: 1},
		{High: "100", Low: "100", Close: "100", Volume: "1", StartTime: 2},
	}
	row := BuildVWAPL2InteractionRow(testOrderBookSnapshot(), candles)
	if row.Mid != 100.5 {
		t.Fatalf("mid %.4f want 100.5", row.Mid)
	}
	if row.SessionVWAP != 100 {
		t.Fatalf("vwap %.4f want 100", row.SessionVWAP)
	}
	if row.DistanceFromVWAP != 0.5 {
		t.Fatalf("distance %.4f want 0.5", row.DistanceFromVWAP)
	}
	if row.LiquidityNearVWAP == 0 {
		t.Fatal("expected liquidity near vwap")
	}
}

func TestOrderBookFeatureWriters(t *testing.T) {
	dir := t.TempDir()
	featuresPath := filepath.Join(dir, "orderbook_features.csv")
	interactionPath := filepath.Join(dir, "vwap_l2_interaction.csv")

	feature := BuildOrderBookFeatureRow(testOrderBookSnapshot())
	if err := WriteOrderBookFeaturesCSV(featuresPath, []OrderBookFeatureRow{feature}); err != nil {
		t.Fatalf("write features: %v", err)
	}
	assertContains(t, featuresPath, "timestamp,venue,symbol,best_bid,best_ask,spread,spread_pct,mid,bid_depth_1pct,ask_depth_1pct,imbalance_1pct,valid")
	assertContains(t, featuresPath, "hyperliquid")

	interaction := VWAPL2InteractionRow{Timestamp: 1, Venue: "hyperliquid", Symbol: "BTC", Mid: 100.5}
	if err := WriteVWAPL2InteractionCSV(interactionPath, []VWAPL2InteractionRow{interaction}); err != nil {
		t.Fatalf("write interaction: %v", err)
	}
	assertContains(t, interactionPath, "timestamp,venue,symbol,mid,session_vwap,distance_from_vwap,spread_pct,imbalance_1pct,liquidity_near_vwap")
	assertContains(t, interactionPath, "hyperliquid")
}

func testOrderBookSnapshot() orderbook.OrderBookSnapshot {
	return orderbook.OrderBookSnapshot{
		Venue:  "hyperliquid",
		Symbol: "BTC",
		Time:   1,
		Bids: []orderbook.BookLevel{
			{Price: "100", Size: "2"},
			{Price: "99", Size: "3"},
		},
		Asks: []orderbook.BookLevel{
			{Price: "101", Size: "4"},
			{Price: "102", Size: "5"},
		},
		IsSnapshot: true,
	}
}

func assertContains(t *testing.T, path string, needle string) {
	t.Helper()
	body, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	if !strings.Contains(string(body), needle) {
		t.Fatalf("%s missing %q: %s", path, needle, string(body))
	}
}
