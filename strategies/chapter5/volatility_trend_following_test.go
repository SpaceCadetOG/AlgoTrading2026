package chapter5

import (
	"testing"

	"AlgoTrading2026/backtest"
	"AlgoTrading2026/series"
	"AlgoTrading2026/volatility"
)

func TestAdaptiveMomentumThresholdsChangeByRegime(t *testing.T) {
	cfg := VolatilityTrendConfig{MomentumPeriod: 2, BaseThreshold: 10, VolatilityMultiplier: 1}

	low := AdaptiveMomentumThreshold(volatility.RegimeLow, cfg)
	normal := AdaptiveMomentumThreshold(volatility.RegimeNormal, cfg)
	high := AdaptiveMomentumThreshold(volatility.RegimeHigh, cfg)

	if low != 5 {
		t.Fatalf("low threshold = %f, want 5", low)
	}
	if normal != 10 {
		t.Fatalf("normal threshold = %f, want 10", normal)
	}
	if high != 20 {
		t.Fatalf("high threshold = %f, want 20", high)
	}
	if !(low < normal && high > normal) {
		t.Fatalf("expected low tighter and high wider thresholds")
	}
}

func TestVolatilityTrendFollowingSignalLength(t *testing.T) {
	frame := frameFromCloses([]float64{100, 101, 102, 101, 103})
	got := VolatilityTrendFollowing(frame, VolatilityTrendConfig{MomentumPeriod: 2, BaseThreshold: 0})

	if len(got.Positions) != len(frame.Rows) {
		t.Fatalf("positions len = %d, want %d", len(got.Positions), len(frame.Rows))
	}
}

func TestVolatilityTrendFollowingBacktestRuns(t *testing.T) {
	candles := testCandles([]float64{100, 101, 102, 101, 103})
	frame := series.FromCandles(candles)
	signals := VolatilityTrendFollowing(frame, VolatilityTrendConfig{MomentumPeriod: 2, BaseThreshold: 0})

	engine := backtest.NewEngine(testConfig(), testRiskEngine())
	report := engine.RunWithSignals(candles, signals)
	if report.CandleCount != len(candles) {
		t.Fatalf("candle count = %d, want %d", report.CandleCount, len(candles))
	}
}
