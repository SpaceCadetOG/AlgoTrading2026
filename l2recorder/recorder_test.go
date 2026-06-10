package l2recorder

import (
	"fmt"
	"path/filepath"
	"testing"
	"time"

	"AlgoTrading2026/orderbook"
)

func TestDefaultRecorderConfig(t *testing.T) {
	cfg := DefaultRecorderConfig()
	if len(cfg.Symbols) != 1 || cfg.Symbols[0] != "BTC" {
		t.Fatalf("symbols=%v", cfg.Symbols)
	}
	if len(cfg.Venues) != 3 {
		t.Fatalf("venues=%v", cfg.Venues)
	}
	if cfg.IntervalSeconds != 5 || cfg.MaxSnapshots != 12 {
		t.Fatalf("unexpected defaults: %+v", cfg)
	}
	if cfg.OutputPath != filepath.Join(".", "data", "l2", "l2_snapshots.csv") {
		t.Fatalf("output=%s", cfg.OutputPath)
	}
}

func TestSnapshotToRow(t *testing.T) {
	row := SnapshotToRow(testSnapshot("hyperliquid", "BTC"), nil, time.UnixMilli(123))
	if !row.Valid {
		t.Fatalf("row invalid: %+v", row)
	}
	if row.BestBid != 100 || row.BestAsk != 101 || row.Spread != 1 {
		t.Fatalf("bad row metrics: %+v", row)
	}
	if row.BidDepth1Pct == 0 || row.AskDepth1Pct == 0 {
		t.Fatalf("missing depth: %+v", row)
	}
}

func TestRecorderContinuesOnVenueFailure(t *testing.T) {
	storage := &memoryStorage{}
	recorder := NewRecorder(RecorderConfig{
		Symbols:         []string{"BTC"},
		Venues:          []string{"hyperliquid", "aster", "lighter"},
		IntervalSeconds: 0,
		MaxSnapshots:    1,
		OutputPath:      "ignored.csv",
	}, mockFetcher{failVenue: "aster"}, storage)
	recorder.Now = func() time.Time { return time.UnixMilli(123) }

	summary, err := recorder.Run()
	if err != nil {
		t.Fatalf("run: %v", err)
	}
	if summary.Rounds != 1 || summary.Rows != 3 || summary.Errors != 1 {
		t.Fatalf("bad summary: %+v", summary)
	}
	if len(storage.rows) != 3 {
		t.Fatalf("rows=%d want 3", len(storage.rows))
	}
	if storage.rows[1].Venue != "aster" || storage.rows[1].Error == "" {
		t.Fatalf("expected aster error row: %+v", storage.rows[1])
	}
}

func TestRecorderStopsAtMaxSnapshots(t *testing.T) {
	storage := &memoryStorage{}
	recorder := NewRecorder(RecorderConfig{
		Symbols:         []string{"BTC"},
		Venues:          []string{"hyperliquid", "lighter"},
		IntervalSeconds: 0,
		MaxSnapshots:    3,
		OutputPath:      "ignored.csv",
	}, mockFetcher{}, storage)
	recorder.Now = func() time.Time { return time.UnixMilli(123) }

	summary, err := recorder.Run()
	if err != nil {
		t.Fatalf("run: %v", err)
	}
	if summary.Rounds != 3 || summary.Rows != 6 {
		t.Fatalf("summary=%+v want 3 rounds 6 rows", summary)
	}
	if len(storage.rows) != 6 {
		t.Fatalf("rows=%d want 6", len(storage.rows))
	}
}

type mockFetcher struct {
	failVenue string
}

func (m mockFetcher) FetchOrderBook(venue string, symbol string) (orderbook.OrderBookSnapshot, error) {
	if venue == m.failVenue {
		return orderbook.OrderBookSnapshot{}, fmt.Errorf("mock failure")
	}
	return testSnapshot(venue, symbol), nil
}

type memoryStorage struct {
	rows []SnapshotRow
}

func (m *memoryStorage) AppendRows(rows []SnapshotRow) error {
	m.rows = append(m.rows, rows...)
	return nil
}

func testSnapshot(venue string, symbol string) orderbook.OrderBookSnapshot {
	return orderbook.OrderBookSnapshot{
		Venue:  venue,
		Symbol: symbol,
		Time:   123,
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
