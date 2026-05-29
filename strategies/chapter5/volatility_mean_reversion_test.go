package chapter5

import (
	"strconv"
	"testing"

	"AlgoTrading2026/backtest"
	"AlgoTrading2026/exchanges"
	"AlgoTrading2026/risk"
	"AlgoTrading2026/series"
	"AlgoTrading2026/volatility"
)

func TestAdaptiveRSIThresholdsChangeByRegime(t *testing.T) {
	cfg := DefaultVolatilityMeanReversionConfig()

	lowEntry, lowExit := AdaptiveRSIThresholds(volatility.RegimeLow, cfg)
	normalEntry, normalExit := AdaptiveRSIThresholds(volatility.RegimeNormal, cfg)
	highEntry, highExit := AdaptiveRSIThresholds(volatility.RegimeHigh, cfg)

	if lowEntry != 35 || lowExit != 55 {
		t.Fatalf("low thresholds = %f/%f, want 35/55", lowEntry, lowExit)
	}
	if normalEntry != 30 || normalExit != 50 {
		t.Fatalf("normal thresholds = %f/%f, want 30/50", normalEntry, normalExit)
	}
	if highEntry != 25 || highExit != 45 {
		t.Fatalf("high thresholds = %f/%f, want 25/45", highEntry, highExit)
	}
	if !(highEntry < normalEntry && lowEntry > normalEntry) {
		t.Fatalf("expected high wider and low tighter entries")
	}
}

func TestVolatilityMeanReversionSignalLength(t *testing.T) {
	frame := frameFromCloses([]float64{100, 101, 99, 98, 100, 102, 101})
	got := VolatilityMeanReversion(frame, VolatilityMeanReversionConfig{BollingerPeriod: 2, RSIPeriod: 2})

	if len(got.Positions) != len(frame.Rows) {
		t.Fatalf("positions len = %d, want %d", len(got.Positions), len(frame.Rows))
	}
}

func TestVolatilityMeanReversionBacktestRuns(t *testing.T) {
	candles := testCandles([]float64{100, 101, 99, 98, 100, 102, 101})
	frame := series.FromCandles(candles)
	signals := VolatilityMeanReversion(frame, VolatilityMeanReversionConfig{BollingerPeriod: 2, RSIPeriod: 2})

	engine := backtest.NewEngine(testConfig(), testRiskEngine())
	report := engine.RunWithSignals(candles, signals)
	if report.CandleCount != len(candles) {
		t.Fatalf("candle count = %d, want %d", report.CandleCount, len(candles))
	}
}

func frameFromCloses(closes []float64) series.Frame {
	return series.FromCandles(testCandles(closes))
}

func testCandles(closes []float64) []exchanges.Candle {
	candles := make([]exchanges.Candle, 0, len(closes))
	for i, close := range closes {
		value := strconv.FormatFloat(close, 'f', -1, 64)
		candles = append(candles, exchanges.Candle{
			Venue:     "test",
			Symbol:    "BTC",
			Interval:  "15m",
			Open:      value,
			High:      value,
			Low:       value,
			Close:     value,
			Volume:    "1",
			StartTime: int64(i + 1),
			EndTime:   int64(i + 2),
			Closed:    true,
		})
	}
	return candles
}

func testConfig() backtest.Config {
	return backtest.Config{
		Venue:           "test",
		Symbol:          "BTC",
		Interval:        "15m",
		StartingBalance: 10000,
		FixedNotional:   100,
		FeeModel:        backtest.StaticFeeModel{},
		SlippageModel:   backtest.StaticSlippageModel{},
	}
}

func testRiskEngine() *risk.Engine {
	limits := risk.DefaultLimits()
	limits.LiveTradingEnabled = true
	limits.MaxOpenPositions = 5
	limits.MaxTotalExposureUSD = 1000
	limits.MaxSymbolExposureUSD = 1000
	return risk.NewEngine(limits)
}
