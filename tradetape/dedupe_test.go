package tradetape

import "testing"

func TestDeduplicatePrintsByTradeID(t *testing.T) {
	prints := []TradeTapePrint{
		dedupePrint("1", 100, 1, "BUY"),
		dedupePrint("1", 100, 1, "BUY"),
		dedupePrint("2", 101, 1, "SELL"),
	}
	deduped, summary := DeduplicatePrints(prints)
	if len(deduped) != 2 || summary.InputRows != 3 || summary.OutputRows != 2 || summary.DuplicateRows != 1 {
		t.Fatalf("unexpected dedupe: deduped=%+v summary=%+v", deduped, summary)
	}
}

func TestDeduplicatePrintsFallbackKeyAndInvalid(t *testing.T) {
	prints := []TradeTapePrint{
		dedupePrint("", 100, 1, "BUY"),
		dedupePrint("", 100, 1, "BUY"),
		dedupePrint("", 0, 1, "BUY"),
	}
	deduped, summary := DeduplicatePrints(prints)
	if len(deduped) != 1 || summary.DuplicateRows != 1 || summary.InvalidRows != 1 {
		t.Fatalf("unexpected fallback dedupe: deduped=%+v summary=%+v", deduped, summary)
	}
}

func dedupePrint(tradeID string, price float64, size float64, side string) TradeTapePrint {
	return TradeTapePrint{
		Venue: "aster", Symbol: "BTCUSDT", VenueSymbol: "BTCUSDT", CanonicalSymbol: "BTC",
		MarketID: "BTCUSDT", Timestamp: 1000, Price: price, Size: size,
		Side: side, AggressorSide: side, TradeID: tradeID,
	}
}
