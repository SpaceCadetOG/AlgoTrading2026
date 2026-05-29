package backtest

import (
	"os"
	"strings"
	"testing"
	"time"

	"AlgoTrading2026/strategy"
)

func TestWriteTradesCSVProducesHeaderAndRows(t *testing.T) {
	path := t.TempDir() + "/trades.csv"
	trade := strategy.Trade{
		Venue:       "aster",
		Symbol:      "BTCUSDT",
		Side:        "LONG",
		Size:        0.01,
		EntryPrice:  100,
		ExitPrice:   110,
		OpenedAt:    time.Unix(1, 0).UTC(),
		ClosedAt:    time.Unix(2, 0).UTC(),
		GrossPnL:    0.1,
		Fees:        0.01,
		RealizedPnL: 0.09,
		ExitReason:  "test",
	}

	if err := WriteTradesCSV(path, []strategy.Trade{trade}); err != nil {
		t.Fatal(err)
	}

	body, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}

	text := string(body)
	if !strings.Contains(text, "trade_id,venue,symbol,side,entry_time") {
		t.Fatalf("missing header: %s", text)
	}
	if !strings.Contains(text, "aster,BTCUSDT,LONG") {
		t.Fatalf("missing row: %s", text)
	}
}
