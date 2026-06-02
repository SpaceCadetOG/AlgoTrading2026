package research

type InstrumentUniverseSummary struct {
	Venues        int            `json:"venues"`
	Symbols       int            `json:"symbols"`
	Crypto        int            `json:"crypto"`
	RWA           int            `json:"rwa"`
	FX            int            `json:"fx"`
	Commodity     int            `json:"commodity"`
	Index         int            `json:"index"`
	Unknown       int            `json:"unknown"`
	TopInstrument string         `json:"topInstrument"`
	AssetCounts   map[string]int `json:"assetCounts"`
	Notes         []string       `json:"notes"`
}

func BuildInstrumentUniverseSummary(rows []InstrumentUniverseRow) InstrumentUniverseSummary {
	summary := InstrumentUniverseSummary{
		AssetCounts: map[string]int{},
		Notes: []string{
			"Research-only instrument universe scanner.",
			"Not a trade scanner, alert scanner, or execution scanner.",
			"Uses defensive placeholders where venue universe endpoints are not normalized yet.",
		},
	}
	venues := map[string]bool{}
	for i, row := range rows {
		venues[row.Venue] = true
		summary.AssetCounts[row.AssetType]++
		if i == 0 {
			summary.TopInstrument = row.Venue + ":" + row.Symbol
		}
	}
	summary.Venues = len(venues)
	summary.Symbols = len(rows)
	summary.Crypto = summary.AssetCounts["crypto"]
	summary.RWA = summary.AssetCounts["rwa"]
	summary.FX = summary.AssetCounts["fx"]
	summary.Commodity = summary.AssetCounts["commodity"]
	summary.Index = summary.AssetCounts["index"]
	summary.Unknown = summary.AssetCounts["unknown"]
	return summary
}
