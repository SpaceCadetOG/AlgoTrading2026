package backtest

import (
	"strconv"
	"testing"

	"AlgoTrading2026/exchanges"
	"AlgoTrading2026/risk"
	"AlgoTrading2026/series"
)

func TestRunRiskEnforcedBlocksOversizedEntries(t *testing.T) {
	candles := riskTestCandles([]float64{100, 101, 102})
	signals := series.SignalResult{
		Times:     []int64{0, 900000, 1800000},
		Signal:    []float64{0, 1, 1},
		Positions: []float64{0, 1, 0},
	}

	result := RunRiskEnforcedWithSignals(candles, signals, testConfig(), risk.RiskControlConfig{
		MaxTradesPerDay:           10,
		MaxTradeSize:              10,
		MaxNotional:               50,
		MaxHoldBars:               10,
		StopLossPct:               0.05,
		MaxVolumeParticipationPct: 100,
	}, "test")

	if result.Before.Metrics.TotalTrades != 0 && result.After.Metrics.TotalTrades > result.Before.Metrics.TotalTrades {
		t.Fatalf("risk controls should not increase trades")
	}
	if result.ViolationsByRule[risk.RuleMaxNotional] == 0 {
		t.Fatalf("expected max notional violation, got %+v", result.ViolationsByRule)
	}
}

func floatString(value float64) string {
	return strconv.FormatFloat(value, 'f', -1, 64)
}

func TestRunRiskEnforcedForcesStopLossExit(t *testing.T) {
	candles := riskTestCandles([]float64{100, 94, 89, 88})
	signals := series.SignalResult{
		Times:     []int64{0, 900000, 1800000, 2700000},
		Signal:    []float64{0, 1, 1, 1},
		Positions: []float64{0, 1, 0, 0},
	}

	result := RunRiskEnforcedWithSignals(candles, signals, testConfig(), risk.RiskControlConfig{
		MaxTradesPerDay:           10,
		MaxTradeSize:              10,
		MaxNotional:               1000,
		MaxHoldBars:               10,
		StopLossPct:               0.03,
		MaxVolumeParticipationPct: 100,
	}, "test")

	if result.ViolationsByRule[risk.RuleStopLoss] == 0 {
		t.Fatalf("expected stop loss violation, got %+v", result.ViolationsByRule)
	}
	if result.After.Metrics.TotalTrades == 0 {
		t.Fatalf("expected forced exit to close a trade")
	}
}

func riskTestCandles(closes []float64) []exchanges.Candle {
	out := make([]exchanges.Candle, 0, len(closes))
	for i, close := range closes {
		start := int64(i) * 900000
		out = append(out, exchanges.Candle{
			Venue:     "test",
			Symbol:    "BTC",
			Interval:  "15m",
			Open:      floatString(close),
			High:      floatString(close + 1),
			Low:       floatString(close - 1),
			Close:     floatString(close),
			Volume:    "1000",
			StartTime: start,
			EndTime:   start + 900000,
			Closed:    true,
		})
	}
	return out
}
