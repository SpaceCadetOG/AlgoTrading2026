package realism

type RealismAssumptions struct {
	SlippageModeled                   bool `json:"slippageModeled"`
	FeesModeled                       bool `json:"feesModeled"`
	LatencyModeled                    bool `json:"latencyModeled"`
	LatencyVarianceModeled            bool `json:"latencyVarianceModeled"`
	PlaceInLineModeled                bool `json:"placeInLineModeled"`
	MarketImpactModeled               bool `json:"marketImpactModeled"`
	MarketDataAccuracyChecked         bool `json:"marketDataAccuracyChecked"`
	HistoricalLiveFormatParityChecked bool `json:"historicalLiveFormatParityChecked"`
	OperationalInterventionModeled    bool `json:"operationalInterventionModeled"`
	LiveAnalyticsAvailable            bool `json:"liveAnalyticsAvailable"`
	ProfitDecayTracked                bool `json:"profitDecayTracked"`
}

func DefaultRealismAssumptions() RealismAssumptions {
	return RealismAssumptions{
		SlippageModeled: true,
		FeesModeled:     true,
	}
}

func (a RealismAssumptions) Strengths() []string {
	var strengths []string
	if a.SlippageModeled {
		strengths = append(strengths, "slippage modeled")
	}
	if a.FeesModeled {
		strengths = append(strengths, "fees modeled")
	}
	if a.LatencyModeled {
		strengths = append(strengths, "latency modeled")
	}
	if a.LatencyVarianceModeled {
		strengths = append(strengths, "latency variance modeled")
	}
	if a.PlaceInLineModeled {
		strengths = append(strengths, "place-in-line modeled")
	}
	if a.MarketImpactModeled {
		strengths = append(strengths, "market impact modeled")
	}
	if a.MarketDataAccuracyChecked {
		strengths = append(strengths, "market data accuracy checked")
	}
	if a.HistoricalLiveFormatParityChecked {
		strengths = append(strengths, "historical/live format parity checked")
	}
	if a.OperationalInterventionModeled {
		strengths = append(strengths, "operational intervention modeled")
	}
	if a.LiveAnalyticsAvailable {
		strengths = append(strengths, "live analytics available")
	}
	if a.ProfitDecayTracked {
		strengths = append(strengths, "profit decay tracked")
	}
	return strengths
}

func (a RealismAssumptions) Missing() []string {
	var missing []string
	if !a.SlippageModeled {
		missing = append(missing, "slippage not modeled")
	}
	if !a.FeesModeled {
		missing = append(missing, "fees not modeled")
	}
	if !a.LatencyModeled {
		missing = append(missing, "latency not modeled")
	}
	if !a.LatencyVarianceModeled {
		missing = append(missing, "latency variance not modeled")
	}
	if !a.PlaceInLineModeled {
		missing = append(missing, "place-in-line not modeled")
	}
	if !a.MarketImpactModeled {
		missing = append(missing, "market impact not modeled")
	}
	if !a.MarketDataAccuracyChecked {
		missing = append(missing, "market data accuracy not checked")
	}
	if !a.HistoricalLiveFormatParityChecked {
		missing = append(missing, "historical/live format parity not checked")
	}
	if !a.OperationalInterventionModeled {
		missing = append(missing, "operational intervention not modeled")
	}
	if !a.LiveAnalyticsAvailable {
		missing = append(missing, "live analytics not available")
	}
	if !a.ProfitDecayTracked {
		missing = append(missing, "profit decay not tracked")
	}
	return missing
}

func (a RealismAssumptions) MissingCount() int {
	return len(a.Missing())
}
