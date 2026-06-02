package orderbook

import "testing"

func TestBestBidBestAskSpreadMid(t *testing.T) {
	book := testSnapshot()
	if got := BestBid(book).Price; got != "100" {
		t.Fatalf("best bid=%s want 100", got)
	}
	if got := BestAsk(book).Price; got != "101" {
		t.Fatalf("best ask=%s want 101", got)
	}
	if got := Spread(book); got != 1 {
		t.Fatalf("spread=%.4f want 1", got)
	}
	if got := Mid(book); got != 100.5 {
		t.Fatalf("mid=%.4f want 100.5", got)
	}
	assertClose(t, SpreadPct(book), 0.9950248756218906)
}

func TestDepthWithinPct(t *testing.T) {
	book := testSnapshot()
	assertClose(t, BidDepthWithinPct(book, 1), 100*2+99.75*1)
	assertClose(t, AskDepthWithinPct(book, 1), 101*3+101.25*1)
	assertClose(t, DepthWithinPct(book, 1), 100*2+99.75*1+101*3+101.25*1)
}

func TestImbalance(t *testing.T) {
	book := testSnapshot()
	bidDepth := 100*2 + 99.75*1
	askDepth := 101*3 + 101.25*1
	want := (bidDepth - askDepth) / (bidDepth + askDepth)
	assertClose(t, Imbalance(book, 1), want)
}

func TestLiquidityNearPrice(t *testing.T) {
	book := testSnapshot()
	bid, ask := LiquidityNearPrice(book, 100, 0.5)
	assertClose(t, bid, 100*2+99.75*1)
	assertClose(t, ask, 0)

	bid, ask = LiquidityNearPrice(book, 100.5, 1)
	assertClose(t, bid, 100*2+99.75*1)
	assertClose(t, ask, 101*3+101.25*1)
}

func TestValidateSnapshot(t *testing.T) {
	if err := ValidateSnapshot(testSnapshot()); err != nil {
		t.Fatalf("valid snapshot: %v", err)
	}

	book := testSnapshot()
	book.Bids[0].Price = "102"
	if err := ValidateSnapshot(book); err == nil {
		t.Fatal("expected crossed book error")
	}

	book = testSnapshot()
	book.Asks[0].Size = "0"
	if err := ValidateSnapshot(book); err == nil {
		t.Fatal("expected invalid size error")
	}

	book = testSnapshot()
	book.Venue = ""
	if err := ValidateSnapshot(book); err == nil {
		t.Fatal("expected missing venue error")
	}
}

func testSnapshot() OrderBookSnapshot {
	return OrderBookSnapshot{
		Venue:  "synthetic",
		Symbol: "BTC",
		Time:   1,
		Bids: []BookLevel{
			{Price: "99.75", Size: "1"},
			{Price: "100", Size: "2"},
			{Price: "98", Size: "5"},
		},
		Asks: []BookLevel{
			{Price: "102", Size: "4"},
			{Price: "101.25", Size: "1"},
			{Price: "101", Size: "3"},
		},
		IsSnapshot: true,
	}
}

func assertClose(t *testing.T, got float64, want float64) {
	t.Helper()
	if got-want > 1e-9 || want-got > 1e-9 {
		t.Fatalf("got %.12f want %.12f", got, want)
	}
}
