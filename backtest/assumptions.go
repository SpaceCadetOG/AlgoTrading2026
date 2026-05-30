package backtest

type BacktestAssumptions struct {
	EngineModel               string   `json:"engineModel"`
	FillModel                 string   `json:"fillModel"`
	FeeModel                  string   `json:"feeModel"`
	SlippageModel             string   `json:"slippageModel"`
	FillRatio                 string   `json:"fillRatio"`
	LatencyAssumption         string   `json:"latencyAssumption"`
	PartialFillSupport        string   `json:"partialFillSupport"`
	MarketImpactSupport       string   `json:"marketImpactSupport"`
	LookaheadBiasGuard        string   `json:"lookaheadBiasGuard"`
	SurvivorshipBiasNote      string   `json:"survivorshipBiasNote"`
	DataStorageNote           string   `json:"dataStorageNote"`
	EventDrivenSupportStatus  string   `json:"eventDrivenSupportStatus"`
	PaperForwardTestingStatus string   `json:"paperForwardTestingStatus"`
	Gaps                      []string `json:"gaps"`
}

func DefaultBacktestAssumptions() BacktestAssumptions {
	return BacktestAssumptions{
		EngineModel:               "for-loop candle-driven backtester",
		FillModel:                 "simulated fills using candle close with venue-specific slippage",
		FeeModel:                  "venue-specific fee model is implemented and configurable",
		SlippageModel:             "venue-specific slippage model is implemented as configurable placeholder assumptions",
		FillRatio:                 "assumes 100 percent fill for simulated market fills",
		LatencyAssumption:         "latency is not fully modeled",
		PartialFillSupport:        "partial fills are not fully modeled",
		MarketImpactSupport:       "market impact is not modeled",
		LookaheadBiasGuard:        "replay preserves timestamp order and signal execution is candle-by-candle; strategy-specific lookahead must still be reviewed",
		SurvivorshipBiasNote:      "crypto single-symbol research avoids equity delisting survivorship issues, but multi-asset survivorship-bias-free datasets are not implemented",
		DataStorageNote:           "research CSV and JSON exports exist; no HDF5, relational database, or time-series database layer is implemented",
		EventDrivenSupportStatus:  "Chapter 7 components exist, but event-driven backtester is not yet complete",
		PaperForwardTestingStatus: "paper or forward testing is not enabled",
		Gaps: []string{
			"event-driven backtester not yet complete",
			"paper/forward testing not enabled",
			"latency not fully modeled",
			"partial fills not fully modeled",
			"market impact not modeled",
			"fill ratio is fixed at simulated full fill",
			"database/HDF5/time-series storage not implemented",
			"strategy-specific lookahead review remains required",
		},
	}
}
