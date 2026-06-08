package symbols

import "testing"

func TestNormalizeSymbolKnownCrypto(t *testing.T) {
	tests := []struct {
		venue     string
		symbol    string
		canonical string
		base      string
		quote     string
		assetType string
	}{
		{"aster", "BTCUSDT", "BTC", "BTC", "USDT", AssetTypeCrypto},
		{"hyperliquid", "BTC", "BTC", "BTC", "", AssetTypeCrypto},
		{"aster", "ETHUSDT", "ETH", "ETH", "USDT", AssetTypeCrypto},
		{"lighter", "SOLUSDT", "SOL", "SOL", "USDT", AssetTypeCrypto},
	}
	for _, tt := range tests {
		got := NormalizeSymbol(tt.venue, tt.symbol)
		if got.CanonicalSymbol != tt.canonical || got.BaseAsset != tt.base || got.QuoteAsset != tt.quote || got.AssetType != tt.assetType {
			t.Fatalf("NormalizeSymbol(%q, %q)=%+v", tt.venue, tt.symbol, got)
		}
	}
}

func TestNormalizeSymbolUnknownFallback(t *testing.T) {
	got := NormalizeSymbol("venue", "WEIRDPAIR")
	if got.CanonicalSymbol != "WEIRDPAIR" || got.BaseAsset != "WEIRDPAIR" || got.AssetType != AssetTypeUnknown {
		t.Fatalf("unexpected unknown fallback: %+v", got)
	}
}
