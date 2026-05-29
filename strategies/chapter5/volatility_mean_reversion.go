package chapter5

import (
	"AlgoTrading2026/indicators"
	"AlgoTrading2026/series"
	"AlgoTrading2026/volatility"
)

type VolatilityMeanReversionConfig struct {
	BollingerPeriod int
	BollingerStdDev float64

	RSIPeriod int

	BaseEntryRSI float64
	BaseExitRSI  float64

	VolatilityMultiplier float64
}

func DefaultVolatilityMeanReversionConfig() VolatilityMeanReversionConfig {
	return VolatilityMeanReversionConfig{
		BollingerPeriod:      20,
		BollingerStdDev:      2,
		RSIPeriod:            14,
		BaseEntryRSI:         30,
		BaseExitRSI:          50,
		VolatilityMultiplier: 1,
	}
}

func VolatilityMeanReversion(frame series.Frame, cfg VolatilityMeanReversionConfig) series.SignalResult {
	cfg = normalizeMeanReversionConfig(cfg)
	close := frame.CloseColumn()
	bands := indicators.BollingerBands(close, cfg.BollingerPeriod, cfg.BollingerStdDev)
	rsi := indicators.RSI(close, cfg.RSIPeriod)
	regimes := regimeForFrame(frame)
	positions := make([]float64, len(close))

	long := false
	warmup := max(cfg.BollingerPeriod, cfg.RSIPeriod)
	for i := range close {
		if i < warmup {
			continue
		}

		entry, exit := AdaptiveRSIThresholds(regimes[i], cfg)
		switch {
		case (close[i] < bands.Lower[i] || rsi[i] < entry) && !long:
			positions[i] = 1
			long = true
		case (close[i] >= bands.Middle[i] || rsi[i] > exit) && long:
			positions[i] = -1
			long = false
		}
	}

	return resultFromPositions(frame, positions)
}

func AdaptiveRSIThresholds(regime volatility.Regime, cfg VolatilityMeanReversionConfig) (float64, float64) {
	cfg = normalizeMeanReversionConfig(cfg)
	switch regime {
	case volatility.RegimeLow:
		return cfg.BaseEntryRSI + 5*cfg.VolatilityMultiplier, cfg.BaseExitRSI + 5*cfg.VolatilityMultiplier
	case volatility.RegimeHigh:
		return cfg.BaseEntryRSI - 5*cfg.VolatilityMultiplier, cfg.BaseExitRSI - 5*cfg.VolatilityMultiplier
	default:
		return cfg.BaseEntryRSI, cfg.BaseExitRSI
	}
}

func RegimesForMeanReversion(frame series.Frame) []volatility.Regime {
	return regimeForFrame(frame)
}

func normalizeMeanReversionConfig(cfg VolatilityMeanReversionConfig) VolatilityMeanReversionConfig {
	def := DefaultVolatilityMeanReversionConfig()
	if cfg.BollingerPeriod == 0 {
		cfg.BollingerPeriod = def.BollingerPeriod
	}
	if cfg.BollingerStdDev == 0 {
		cfg.BollingerStdDev = def.BollingerStdDev
	}
	if cfg.RSIPeriod == 0 {
		cfg.RSIPeriod = def.RSIPeriod
	}
	if cfg.BaseEntryRSI == 0 {
		cfg.BaseEntryRSI = def.BaseEntryRSI
	}
	if cfg.BaseExitRSI == 0 {
		cfg.BaseExitRSI = def.BaseExitRSI
	}
	if cfg.VolatilityMultiplier == 0 {
		cfg.VolatilityMultiplier = def.VolatilityMultiplier
	}
	return cfg
}
