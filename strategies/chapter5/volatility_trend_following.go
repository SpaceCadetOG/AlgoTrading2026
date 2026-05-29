package chapter5

import (
	"AlgoTrading2026/indicators"
	"AlgoTrading2026/series"
	"AlgoTrading2026/volatility"
)

type VolatilityTrendConfig struct {
	MomentumPeriod int

	BaseThreshold float64

	VolatilityMultiplier float64
}

func DefaultVolatilityTrendConfig() VolatilityTrendConfig {
	return VolatilityTrendConfig{
		MomentumPeriod:       20,
		BaseThreshold:        0,
		VolatilityMultiplier: 1,
	}
}

func VolatilityTrendFollowing(frame series.Frame, cfg VolatilityTrendConfig) series.SignalResult {
	cfg = normalizeTrendConfig(cfg)
	close := frame.CloseColumn()
	momentum := indicators.Momentum(close, cfg.MomentumPeriod)
	regimes := regimeForFrame(frame)
	positions := make([]float64, len(close))

	long := false
	for i, value := range momentum {
		if i < cfg.MomentumPeriod {
			continue
		}
		threshold := AdaptiveMomentumThreshold(regimes[i], cfg)
		switch {
		case value > threshold && !long:
			positions[i] = 1
			long = true
		case value < -threshold && long:
			positions[i] = -1
			long = false
		}
	}

	return resultFromPositions(frame, positions)
}

func AdaptiveMomentumThreshold(regime volatility.Regime, cfg VolatilityTrendConfig) float64 {
	cfg = normalizeTrendConfig(cfg)
	switch regime {
	case volatility.RegimeLow:
		return cfg.BaseThreshold * 0.5 * cfg.VolatilityMultiplier
	case volatility.RegimeHigh:
		return cfg.BaseThreshold * 2 * cfg.VolatilityMultiplier
	default:
		return cfg.BaseThreshold * cfg.VolatilityMultiplier
	}
}

func RegimesForTrendFollowing(frame series.Frame) []volatility.Regime {
	return regimeForFrame(frame)
}

func normalizeTrendConfig(cfg VolatilityTrendConfig) VolatilityTrendConfig {
	def := DefaultVolatilityTrendConfig()
	if cfg.MomentumPeriod == 0 {
		cfg.MomentumPeriod = def.MomentumPeriod
	}
	if cfg.VolatilityMultiplier == 0 {
		cfg.VolatilityMultiplier = def.VolatilityMultiplier
	}
	return cfg
}
