package runtime

import (
	"math"
	"testing"

	"AlgoTrading2026/orderbook"
	"AlgoTrading2026/strategy"
)

func TestPlaybookCandidateBuilderBuildsDeterministicRuntimeCandidates(t *testing.T) {
	builder := NewPlaybookCandidateBuilder(strategy.DefaultBookTradeRules())
	builder.NotionalUSD = 100
	ctx := ContextFromOrderBook(orderbook.OrderBookSnapshot{
		Venue:  "aster",
		Symbol: "BTCUSDT",
		Bids: []orderbook.BookLevel{
			{Price: "100.00", Size: "10"},
			{Price: "99.90", Size: "5"},
		},
		Asks: []orderbook.BookLevel{
			{Price: "100.10", Size: "4"},
			{Price: "100.20", Size: "2"},
		},
	})

	candidates := builder.BuildCandidates(ctx)
	if len(candidates) != len(strategy.DefaultBookTradeRules()) {
		t.Fatalf("candidates=%d want %d", len(candidates), len(strategy.DefaultBookTradeRules()))
	}
	first := candidates[0]
	if first.Provenance == "synthetic_test_candidate" {
		t.Fatalf("expected runtime provenance, got %+v", first)
	}
	if first.Symbol != "BTCUSDT" || first.CanonicalSymbol != "BTC" || first.EntryPrice <= 0 {
		t.Fatalf("unexpected candidate identity: %+v", first)
	}
	if first.Side != "LONG" || !(first.StopPrice < first.EntryPrice && first.TP1 > first.EntryPrice) {
		t.Fatalf("expected long bracket geometry, got %+v", first)
	}
	if math.Abs(first.Quantity-(100/first.EntryPrice)) > 1e-9 {
		t.Fatalf("quantity=%f not within tolerance", first.Quantity)
	}
}
