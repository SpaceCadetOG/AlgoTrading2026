package symbols

type SymbolIdentity struct {
	Venue           string
	VenueSymbol     string
	CanonicalSymbol string
	QuoteAsset      string
	BaseAsset       string
	AssetType       string
}

const (
	AssetTypeCrypto  = "crypto"
	AssetTypeRWA     = "rwa"
	AssetTypeUnknown = "unknown"
)
