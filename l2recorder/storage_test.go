package l2recorder

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestCSVStorageAppendRows(t *testing.T) {
	path := filepath.Join(t.TempDir(), "data", "l2_snapshots.csv")
	storage := NewCSVStorage(path)

	row := SnapshotToRow(testSnapshot("hyperliquid", "BTC"), nil, time.UnixMilli(123))
	if err := storage.AppendRows([]SnapshotRow{row}); err != nil {
		t.Fatalf("append rows: %v", err)
	}
	if err := storage.AppendRows([]SnapshotRow{ErrorRow("aster", "BTC", os.ErrNotExist, time.UnixMilli(124))}); err != nil {
		t.Fatalf("append error row: %v", err)
	}

	body, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read csv: %v", err)
	}
	text := string(body)
	header := "timestamp,venue,symbol,best_bid,best_ask,spread,spread_pct,mid,bid_depth_1pct,ask_depth_1pct,imbalance_1pct,valid,error"
	if strings.Count(text, header) != 1 {
		t.Fatalf("header count mismatch: %s", text)
	}
	if !strings.Contains(text, "hyperliquid,BTC") {
		t.Fatalf("missing valid row: %s", text)
	}
	if !strings.Contains(text, "aster,BTC") || !strings.Contains(text, "file does not exist") {
		t.Fatalf("missing error row: %s", text)
	}
}
