package backtest

import (
	"strings"
)

type SlippageModel interface {
	Apply(venue string, side string, price float64, notional float64, liquidity string) float64
}

type VenueSlippageConfig struct {
	MakerBps float64
	TakerBps float64
}

type StaticSlippageModel struct {
	Venues map[string]VenueSlippageConfig
}

func DefaultSlippageModel() StaticSlippageModel {
	return StaticSlippageModel{
		Venues: map[string]VenueSlippageConfig{
			"aster":       {MakerBps: 0, TakerBps: 2},
			"hyperliquid": {MakerBps: 0, TakerBps: 1.5},
			"lighter":     {MakerBps: 0, TakerBps: 1},
		},
	}
}

func (m StaticSlippageModel) WithVenue(venue string, config VenueSlippageConfig) StaticSlippageModel {
	if m.Venues == nil {
		m.Venues = make(map[string]VenueSlippageConfig)
	}
	m.Venues[normalizeVenue(venue)] = config
	return m
}

func (m StaticSlippageModel) Apply(venue string, side string, price float64, _ float64, liquidity string) float64 {
	config := m.Venues[normalizeVenue(venue)]
	bps := config.TakerBps
	if strings.EqualFold(liquidity, LiquidityMaker) {
		bps = config.MakerBps
	}
	if bps == 0 {
		return price
	}

	adjustment := bps / 10000
	switch strings.ToUpper(strings.TrimSpace(side)) {
	case "BUY", "LONG":
		return price * (1 + adjustment)
	case "SELL", "SHORT":
		return price * (1 - adjustment)
	default:
		return price
	}
}
