package chapter4

import (
	"strconv"
	"testing"

	"AlgoTrading2026/backtest"
	"AlgoTrading2026/exchanges"
	"AlgoTrading2026/risk"
	"AlgoTrading2026/series"
)

func TestMomentumOpensAndCloses(t *testing.T) {
	frame := frameFromCloses([]float64{10, 11, 12, 11})
	got := MomentumStrategy(frame, 1)

	assertPosition(t, got.Positions, 1, 1)
	assertPosition(t, got.Positions, 3, -1)
}

func TestDualMADetectsCrosses(t *testing.T) {
	frame := frameFromCloses([]float64{10, 10, 10, 12, 14, 10, 8, 6})
	got := DualMAStrategy(frame, 2, 3)

	if !containsPosition(got.Positions, 1) {
		t.Fatalf("expected dual ma buy signal, got %v", got.Positions)
	}
	if !containsPosition(got.Positions, -1) {
		t.Fatalf("expected dual ma sell signal, got %v", got.Positions)
	}
}

func TestTurtleUsesPriorBreakoutLevels(t *testing.T) {
	frame := frameFromOHLC(
		[]float64{10, 10, 10, 11},
		[]float64{10, 10, 10, 11},
		[]float64{9, 9, 9, 10},
	)
	got := TurtleBreakoutStrategy(frame, 3, 2)

	assertPosition(t, got.Positions, 3, 1)
}

func TestTurtleAvoidsCurrentCandleLookahead(t *testing.T) {
	frame := frameFromOHLC(
		[]float64{10, 10, 10, 9.5},
		[]float64{10, 10, 10, 20},
		[]float64{9, 9, 9, 9},
	)
	got := TurtleBreakoutStrategy(frame, 3, 2)

	if containsPosition(got.Positions, 1) {
		t.Fatalf("unexpected turtle buy from current candle high: %v", got.Positions)
	}
}

func TestMeanReversionOpensBelowLowerBandAndExitsAtMiddle(t *testing.T) {
	frame := frameFromCloses([]float64{100, 100, 100, 90, 100})
	got := MeanReversionStrategy(frame, 3, 1, 2, 0, 100)

	assertPosition(t, got.Positions, 3, 1)
	assertPosition(t, got.Positions, 4, -1)
}

func TestMeanReversionOpensOnLowRSIAndExitsOnRSIRecovery(t *testing.T) {
	frame := frameFromCloses([]float64{100, 99, 98, 97, 98})
	got := MeanReversionStrategy(frame, 3, 100, 2, 30, 50)

	if !containsPosition(got.Positions, 1) {
		t.Fatalf("expected RSI-driven mean reversion buy, got %v", got.Positions)
	}
	if !containsPosition(got.Positions, -1) {
		t.Fatalf("expected RSI-driven mean reversion sell, got %v", got.Positions)
	}
}

func TestChapter4StrategiesProduceSameLengthAsCandles(t *testing.T) {
	frame := frameFromCloses([]float64{100, 101, 102, 99, 98, 103, 104, 100})
	results := []series.SignalResult{
		MomentumStrategy(frame, 2),
		DualMAStrategy(frame, 2, 3),
		TurtleBreakoutStrategy(frame, 3, 2),
		MeanReversionStrategy(frame, 3, 2, 2, 30, 50),
	}

	for i, result := range results {
		if len(result.Positions) != len(frame.Rows) {
			t.Fatalf("strategy %d positions len = %d, want %d", i, len(result.Positions), len(frame.Rows))
		}
	}
}

func TestBacktestRunsChapter4Strategies(t *testing.T) {
	candles := testCandles([]float64{100, 101, 102, 99, 98, 103, 104, 100})
	frame := series.FromCandles(candles)
	limits := risk.DefaultLimits()
	limits.LiveTradingEnabled = true
	limits.MaxOpenPositions = 5
	limits.MaxTotalExposureUSD = 1000
	limits.MaxSymbolExposureUSD = 1000

	results := []series.SignalResult{
		MomentumStrategy(frame, 2),
		DualMAStrategy(frame, 2, 3),
		TurtleBreakoutStrategy(frame, 3, 2),
		MeanReversionStrategy(frame, 3, 2, 2, 30, 50),
	}

	for i, result := range results {
		engine := backtest.NewEngine(backtest.Config{
			Venue:           "test",
			Symbol:          "BTC",
			Interval:        "15m",
			StartingBalance: 10000,
			FixedNotional:   100,
			FeeModel:        backtest.StaticFeeModel{},
			SlippageModel:   backtest.StaticSlippageModel{},
		}, risk.NewEngine(limits))
		report := engine.RunWithSignals(candles, result)
		if report.CandleCount != len(candles) {
			t.Fatalf("strategy %d candle count = %d, want %d", i, report.CandleCount, len(candles))
		}
	}
}

func frameFromCloses(closes []float64) series.Frame {
	return series.FromCandles(testCandles(closes))
}

func frameFromOHLC(closes []float64, highs []float64, lows []float64) series.Frame {
	candles := make([]exchanges.Candle, 0, len(closes))
	for i := range closes {
		candles = append(candles, candle(i, closes[i], highs[i], lows[i]))
	}
	return series.FromCandles(candles)
}

func testCandles(closes []float64) []exchanges.Candle {
	candles := make([]exchanges.Candle, 0, len(closes))
	for i, close := range closes {
		candles = append(candles, candle(i, close, close, close))
	}
	return candles
}

func candle(index int, close float64, high float64, low float64) exchanges.Candle {
	value := formatFloat(close)
	return exchanges.Candle{
		Venue:     "test",
		Symbol:    "BTC",
		Interval:  "15m",
		Open:      value,
		High:      formatFloat(high),
		Low:       formatFloat(low),
		Close:     value,
		Volume:    "1",
		StartTime: int64(index + 1),
		EndTime:   int64(index + 2),
		Closed:    true,
	}
}

func assertPosition(t *testing.T, values []float64, index int, want float64) {
	t.Helper()
	if values[index] != want {
		t.Fatalf("position[%d] = %f, want %f; all=%v", index, values[index], want, values)
	}
}

func containsPosition(values []float64, want float64) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}

func formatFloat(value float64) string {
	return strconv.FormatFloat(value, 'f', -1, 64)
}
