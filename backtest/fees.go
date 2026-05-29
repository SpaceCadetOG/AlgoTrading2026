package backtest

import "strings"

const (
	LiquidityMaker = "maker"
	LiquidityTaker = "taker"
)

type FeeModel interface {
	CalculateFee(venue string, notional float64, liquidity string) float64
}

type VenueFeeConfig struct {
	MakerRate float64
	TakerRate float64
}

type StaticFeeModel struct {
	Venues map[string]VenueFeeConfig
}

func DefaultFeeModel() StaticFeeModel {
	return StaticFeeModel{
		Venues: map[string]VenueFeeConfig{
			"aster":       {MakerRate: 0, TakerRate: 0.0004},
			"hyperliquid": {MakerRate: 0, TakerRate: 0},
			"lighter":     {MakerRate: 0, TakerRate: 0},
		},
	}
}

func (m StaticFeeModel) WithVenue(venue string, config VenueFeeConfig) StaticFeeModel {
	if m.Venues == nil {
		m.Venues = make(map[string]VenueFeeConfig)
	}
	m.Venues[normalizeVenue(venue)] = config
	return m
}

func (m StaticFeeModel) CalculateFee(venue string, notional float64, liquidity string) float64 {
	config := m.Venues[normalizeVenue(venue)]
	if strings.EqualFold(liquidity, LiquidityMaker) {
		return notional * config.MakerRate
	}

	return notional * config.TakerRate
}

func normalizeVenue(venue string) string {
	return strings.ToLower(strings.TrimSpace(venue))
}
