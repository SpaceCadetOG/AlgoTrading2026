package chapter2

import (
	"testing"

	"AlgoTrading2026/backtest"
	"AlgoTrading2026/exchanges"
	"AlgoTrading2026/risk"
	"AlgoTrading2026/series"
)

func TestChapter2StrategiesProduceValidSignals(t *testing.T) {
	frame := testFrame()
	strategies := []series.SignalResult{
		SMAStrategy(frame, 3),
		EMAStrategy(frame, 3),
		APOStrategy(frame, 2, 4),
		MACDStrategy(frame, 2, 4, 2),
		BollingerStrategy(frame, 3, 2),
		RSIStrategy(frame, 3, 30, 70),
		MomentumStrategy(frame, 2),
		SupportResistanceStrategy(frame, 3),
	}

	for i, result := range strategies {
		if len(result.Signal) != len(frame.Rows) {
			t.Fatalf("strategy %d signal len = %d, want %d", i, len(result.Signal), len(frame.Rows))
		}
		if len(result.Positions) != len(frame.Rows) {
			t.Fatalf("strategy %d position len = %d, want %d", i, len(result.Positions), len(frame.Rows))
		}
	}
}

func TestBacktestRunsEachChapter2Strategy(t *testing.T) {
	candles := testCandles()
	frame := series.FromCandles(candles)
	limits := risk.DefaultLimits()
	limits.LiveTradingEnabled = true
	limits.MaxOpenPositions = 5
	limits.MaxTotalExposureUSD = 1000
	limits.MaxSymbolExposureUSD = 1000

	strategies := []series.SignalResult{
		SMAStrategy(frame, 3),
		EMAStrategy(frame, 3),
		APOStrategy(frame, 2, 4),
		MACDStrategy(frame, 2, 4, 2),
		BollingerStrategy(frame, 3, 2),
		RSIStrategy(frame, 3, 30, 70),
		MomentumStrategy(frame, 2),
		SupportResistanceStrategy(frame, 3),
	}

	for i, result := range strategies {
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
			t.Fatalf("strategy %d candles = %d, want %d", i, report.CandleCount, len(candles))
		}
	}
}

func testFrame() series.Frame {
	return series.FromCandles(testCandles())
}

func testCandles() []exchanges.Candle {
	closes := []string{"100", "101", "99", "102", "104", "103", "105", "107", "106", "108"}
	candles := make([]exchanges.Candle, 0, len(closes))
	for i, close := range closes {
		candles = append(candles, exchanges.Candle{
			Venue:     "test",
			Symbol:    "BTC",
			Interval:  "15m",
			Open:      "100",
			High:      "110",
			Low:       "90",
			Close:     close,
			Volume:    "1",
			StartTime: int64(i + 1),
			EndTime:   int64(i + 2),
			Closed:    true,
		})
	}
	return candles
}
