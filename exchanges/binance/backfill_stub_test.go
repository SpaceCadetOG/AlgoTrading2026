package binance

import (
	"strings"
	"testing"
)

func TestBackfillStubProviderReturnsNotImplemented(t *testing.T) {
	provider := BackfillStubProvider{}
	if provider.Venue() != "binance" {
		t.Fatalf("unexpected venue: %s", provider.Venue())
	}
	_, err := provider.BackfillTrades("BTCUSDT", 1, 2)
	if err == nil || !strings.Contains(err.Error(), "not implemented") || !strings.Contains(err.Error(), "data.binance.vision") {
		t.Fatalf("unexpected error: %v", err)
	}
}
