package scanner

func DefaultUniverse() []Instrument {
	return []Instrument{
		{Venue: "aster", Symbol: "BTCUSDT", AssetType: AssetCrypto, CandlesAvailable: true, L2Available: true, TradesAvailable: true, ProfileReady: true, Notes: "Aster BTC research baseline."},
		{Venue: "aster", Symbol: "ETHUSDT", AssetType: AssetCrypto, Notes: "Defensive placeholder; verify venue market-data availability before research."},
		{Venue: "aster", Symbol: "SOLUSDT", AssetType: AssetCrypto, Notes: "Defensive placeholder; verify venue market-data availability before research."},
		{Venue: "hyperliquid", Symbol: "BTC", AssetType: AssetCrypto, CandlesAvailable: true, L2Available: true, TradesAvailable: true, ProfileReady: true, Notes: "Hyperliquid BTC candles/trades/L2 available in repo."},
		{Venue: "hyperliquid", Symbol: "ETH", AssetType: AssetCrypto, Notes: "Defensive placeholder; verify venue market-data availability before research."},
		{Venue: "hyperliquid", Symbol: "SOL", AssetType: AssetCrypto, Notes: "Defensive placeholder; verify venue market-data availability before research."},
		{Venue: "lighter", Symbol: "BTC", AssetType: AssetCrypto, CandlesAvailable: true, L2Available: true, TradesAvailable: false, ProfileReady: true, Notes: "Lighter BTC candles/L2 available; trade feed support not established."},
		{Venue: "lighter", Symbol: "ETH", AssetType: AssetCrypto, Notes: "Defensive placeholder; supported market ID known, verify data before research."},
	}
}

func MergeUniverseFacts(base []Instrument, facts []Instrument) []Instrument {
	byKey := map[string]Instrument{}
	for _, instrument := range base {
		if instrument.AssetType == "" {
			instrument.AssetType = ClassifyAsset(instrument.Symbol)
		}
		byKey[instrument.Venue+":"+instrument.Symbol] = instrument
	}
	for _, fact := range facts {
		key := fact.Venue + ":" + fact.Symbol
		current := byKey[key]
		if current.Venue == "" {
			current.Venue = fact.Venue
			current.Symbol = fact.Symbol
			current.AssetType = ClassifyAsset(fact.Symbol)
		}
		if fact.AssetType != "" {
			current.AssetType = fact.AssetType
		}
		if fact.Price != 0 {
			current.Price = fact.Price
		}
		if fact.Volume24h != 0 {
			current.Volume24h = fact.Volume24h
		}
		if fact.OpenInterest != 0 {
			current.OpenInterest = fact.OpenInterest
		}
		if fact.FundingRate != 0 {
			current.FundingRate = fact.FundingRate
		}
		if fact.SpreadPct != 0 {
			current.SpreadPct = fact.SpreadPct
		}
		if fact.VolatilityProxy != 0 {
			current.VolatilityProxy = fact.VolatilityProxy
		}
		current.L2Available = current.L2Available || fact.L2Available
		current.CandlesAvailable = current.CandlesAvailable || fact.CandlesAvailable
		current.TradesAvailable = current.TradesAvailable || fact.TradesAvailable
		current.ProfileReady = current.ProfileReady || fact.ProfileReady
		if fact.Notes != "" {
			if current.Notes != "" {
				current.Notes += " "
			}
			current.Notes += fact.Notes
		}
		byKey[key] = current
	}
	out := make([]Instrument, 0, len(byKey))
	for _, instrument := range byKey {
		out = append(out, instrument)
	}
	return out
}
