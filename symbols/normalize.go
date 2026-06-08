package symbols

import "strings"

var commonQuoteSuffixes = []string{"USDT", "USDC", "USD"}

var knownCryptoMajors = map[string]bool{
	"BTC": true,
	"ETH": true,
	"SOL": true,
}

func NormalizeSymbol(venue string, venueSymbol string) SymbolIdentity {
	trimmed := strings.ToUpper(strings.TrimSpace(venueSymbol))
	identity := SymbolIdentity{
		Venue:           strings.ToLower(strings.TrimSpace(venue)),
		VenueSymbol:     trimmed,
		CanonicalSymbol: trimmed,
		BaseAsset:       trimmed,
		AssetType:       AssetTypeUnknown,
	}
	if trimmed == "" {
		return identity
	}
	for _, quote := range commonQuoteSuffixes {
		if strings.HasSuffix(trimmed, quote) && len(trimmed) > len(quote) {
			identity.BaseAsset = strings.TrimSuffix(trimmed, quote)
			identity.CanonicalSymbol = identity.BaseAsset
			identity.QuoteAsset = quote
			break
		}
	}
	if knownCryptoMajors[identity.CanonicalSymbol] {
		identity.AssetType = AssetTypeCrypto
	}
	return identity
}
