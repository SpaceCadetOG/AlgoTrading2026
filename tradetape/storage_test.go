package tradetape

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestAppendRowsAndReadRows(t *testing.T) {
	path := filepath.Join(t.TempDir(), "trades.csv")
	rows := []TradeTapeRow{
		RowFromPrint(testPrint("hyperliquid", "BTC", 1, 100, 0.5, "B")),
		RowFromPrint(testPrint("aster", "BTCUSDT", 2, 101, 0.25, "A")),
	}
	if err := AppendRows(path, rows[:1]); err != nil {
		t.Fatalf("append first: %v", err)
	}
	if err := AppendRows(path, rows[1:]); err != nil {
		t.Fatalf("append second: %v", err)
	}
	body, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read file: %v", err)
	}
	if strings.Count(string(body), "venue,venue_symbol,canonical_symbol") != 1 {
		t.Fatalf("expected one header, got:\n%s", string(body))
	}
	decoded, err := ReadRows(path)
	if err != nil {
		t.Fatalf("read rows: %v", err)
	}
	if len(decoded) != 2 {
		t.Fatalf("len=%d want 2", len(decoded))
	}
	if decoded[1].Print.Venue != "aster" || decoded[1].Print.Size != 0.25 {
		t.Fatalf("unexpected decoded row: %+v", decoded[1])
	}
	if decoded[1].Print.VenueSymbol != "BTCUSDT" || decoded[1].Print.CanonicalSymbol != "BTC" {
		t.Fatalf("unexpected symbol identity: %+v", decoded[1].Print)
	}
}

func TestValidationFailureRow(t *testing.T) {
	row := RowFromPrint(TradeTapePrint{Venue: "x", Symbol: "BTC", Timestamp: 1, Price: 0, Size: 1})
	if row.Valid || row.Error == "" {
		t.Fatalf("expected invalid row: %+v", row)
	}
}

func testPrint(venue string, symbol string, timestamp int64, price float64, size float64, side string) TradeTapePrint {
	return TradeTapePrint{
		Venue:         venue,
		Symbol:        symbol,
		MarketID:      "1",
		Timestamp:     timestamp,
		Price:         price,
		Size:          size,
		Side:          side,
		AggressorSide: side,
		TradeID:       venue + "-1",
	}
}
