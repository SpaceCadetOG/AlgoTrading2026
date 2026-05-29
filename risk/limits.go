package risk

type Limits struct {
	LiveTradingEnabled bool

	MaxOpenPositions int

	MaxTotalExposureUSD  float64
	MaxSymbolExposureUSD float64

	MaxLeverage float64

	MinAvailableUSD float64

	KillSwitchActive bool
}

func DefaultLimits() Limits {
	return Limits{
		LiveTradingEnabled:   false,
		MaxOpenPositions:     1,
		MaxTotalExposureUSD:  100,
		MaxSymbolExposureUSD: 100,
		MaxLeverage:          3,
		MinAvailableUSD:      10,
		KillSwitchActive:     false,
	}
}
