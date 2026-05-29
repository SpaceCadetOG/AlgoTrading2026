package backtest

import (
	"testing"

	"AlgoTrading2026/exchanges"
	"AlgoTrading2026/risk"
	"AlgoTrading2026/series"
)

func TestReplayOrdersCandlesByTimestamp(t *testing.T) {
	replay := NewReplay([]exchanges.Candle{
		{StartTime: 3},
		{StartTime: 1},
		{StartTime: 2},
	})

	for i, want := range []int64{1, 2, 3} {
		if replay.Candles[i].StartTime != want {
			t.Fatalf("candle %d start = %d, want %d", i, replay.Candles[i].StartTime, want)
		}
	}
}

func TestSignalPlusOneOpensLong(t *testing.T) {
	engine := NewEngine(testConfig(), testRiskEngine())

	engine.RunWithSignals(
		[]exchanges.Candle{testCandleAt(1, "100", "100")},
		testSignals([]int64{1}, []float64{1}),
	)

	if engine.OpenPosition == nil {
		t.Fatal("expected +1 signal to open long")
	}
}

func TestSignalMinusOneClosesLong(t *testing.T) {
	engine := NewEngine(testConfig(), testRiskEngine())

	report := engine.RunWithSignals(
		[]exchanges.Candle{
			testCandleAt(1, "100", "100"),
			testCandleAt(2, "100", "110"),
		},
		testSignals([]int64{1, 2}, []float64{1, -1}),
	)

	if report.Metrics.TotalTrades != 1 {
		t.Fatalf("trades = %d, want 1", report.Metrics.TotalTrades)
	}
	if report.EndingBalance <= report.StartingBalance {
		t.Fatalf("ending balance = %f, want profit over %f", report.EndingBalance, report.StartingBalance)
	}
}

func TestNoDuplicateOpenWhenAlreadyLong(t *testing.T) {
	engine := NewEngine(testConfig(), testRiskEngine())

	report := engine.RunWithSignals(
		[]exchanges.Candle{
			testCandleAt(1, "100", "100"),
			testCandleAt(2, "100", "105"),
		},
		testSignals([]int64{1, 2}, []float64{1, 1}),
	)

	if engine.OpenPosition == nil {
		t.Fatal("expected position to remain open")
	}
	if report.Metrics.TotalTrades != 0 {
		t.Fatalf("trades = %d, want 0 closed trades", report.Metrics.TotalTrades)
	}
}

func TestNoCloseWhenFlat(t *testing.T) {
	engine := NewEngine(testConfig(), testRiskEngine())

	report := engine.RunWithSignals(
		[]exchanges.Candle{testCandleAt(1, "100", "100")},
		testSignals([]int64{1}, []float64{-1}),
	)

	if engine.OpenPosition != nil {
		t.Fatal("expected flat engine to stay flat")
	}
	if report.Metrics.TotalTrades != 0 {
		t.Fatalf("trades = %d, want 0", report.Metrics.TotalTrades)
	}
}

func TestEngineRecordsRiskRejects(t *testing.T) {
	limits := risk.DefaultLimits()
	limits.LiveTradingEnabled = false
	engine := NewEngine(testConfig(), risk.NewEngine(limits))

	report := engine.RunWithSignals(
		[]exchanges.Candle{testCandleAt(1, "100", "110")},
		testSignals([]int64{1}, []float64{1}),
	)

	if len(report.RejectReasons) != 1 {
		t.Fatalf("rejects = %d, want 1", len(report.RejectReasons))
	}
	if report.Metrics.TotalTrades != 0 {
		t.Fatalf("trades = %d, want 0", report.Metrics.TotalTrades)
	}
}

func testConfig() Config {
	return Config{
		Venue:           "test",
		Symbol:          "BTC",
		Interval:        "15m",
		StartingBalance: 10000,
		FixedNotional:   100,
		FeeModel:        StaticFeeModel{},
		SlippageModel:   StaticSlippageModel{},
	}
}

func testRiskEngine() *risk.Engine {
	limits := risk.DefaultLimits()
	limits.LiveTradingEnabled = true
	limits.MaxOpenPositions = 1
	limits.MaxTotalExposureUSD = 1000
	limits.MaxSymbolExposureUSD = 1000
	return risk.NewEngine(limits)
}

func testCandle(open string, close string) exchanges.Candle {
	return testCandleAt(1, open, close)
}

func testCandleAt(startTime int64, open string, close string) exchanges.Candle {
	return exchanges.Candle{
		Venue:     "test",
		Symbol:    "BTC",
		Interval:  "15m",
		Open:      open,
		High:      close,
		Low:       open,
		Close:     close,
		StartTime: startTime,
		EndTime:   startTime + 1,
		Closed:    true,
	}
}

func testSignals(times []int64, positions []float64) series.SignalResult {
	return testSignalsWithSignal(times, nil, positions)
}

func testSignalsWithSignal(times []int64, signals []float64, positions []float64) series.SignalResult {
	return series.SignalResult{
		Times:     times,
		Signal:    signals,
		Positions: positions,
	}
}
