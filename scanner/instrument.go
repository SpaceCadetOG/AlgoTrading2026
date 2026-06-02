package scanner

import "strings"

const (
	AssetCrypto    = "crypto"
	AssetRWA       = "rwa"
	AssetFX        = "fx"
	AssetCommodity = "commodity"
	AssetIndex     = "index"
	AssetUnknown   = "unknown"
)

type Instrument struct {
	Venue            string
	Symbol           string
	AssetType        string
	Price            float64
	Volume24h        float64
	OpenInterest     float64
	FundingRate      float64
	SpreadPct        float64
	VolatilityProxy  float64
	L2Available      bool
	CandlesAvailable bool
	TradesAvailable  bool
	ProfileReady     bool
	ResearchScore    float64
	Rank             int
	Notes            string
}

func ClassifyAsset(symbol string) string {
	upper := strings.ToUpper(strings.TrimSpace(symbol))
	switch upper {
	case "BTC", "BTCUSDT", "ETH", "ETHUSDT", "SOL", "SOLUSDT", "BNB", "XRP", "DOGE", "ADA", "AVAX", "LINK", "SUI", "HYPE":
		return AssetCrypto
	}
	if strings.Contains(upper, "GOLD") || strings.Contains(upper, "XAU") || strings.Contains(upper, "OIL") {
		return AssetCommodity
	}
	if strings.Contains(upper, "SPX") || strings.Contains(upper, "NASDAQ") || strings.Contains(upper, "NDX") || strings.Contains(upper, "DOW") {
		return AssetIndex
	}
	if strings.Contains(upper, "UST") || strings.Contains(upper, "TBILL") || strings.Contains(upper, "STOCK") {
		return AssetRWA
	}
	if strings.Contains(upper, "EUR") || strings.Contains(upper, "JPY") || strings.Contains(upper, "GBP") {
		return AssetFX
	}
	return AssetUnknown
}
