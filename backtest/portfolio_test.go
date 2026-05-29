package backtest

import (
	"os"
	"strings"
	"testing"

	"AlgoTrading2026/exchanges"
)

func TestPortfolioCashDecreasesOnBuy(t *testing.T) {
	engine := NewEngine(testConfig(), testRiskEngine())

	engine.RunWithSignals(
		[]exchanges.Candle{testCandleAt(1, "100", "100")},
		testSignalsWithSignal([]int64{1}, []float64{1}, []float64{1}),
	)

	point := engine.EquityCurve.Points[0]
	if point.Cash != 9900 {
		t.Fatalf("cash = %f, want 9900", point.Cash)
	}
	if point.Holdings != 100 {
		t.Fatalf("holdings = %f, want 100", point.Holdings)
	}
	if point.TotalEquity != point.Cash+point.Holdings {
		t.Fatalf("total equity = %f, want cash + holdings %f", point.TotalEquity, point.Cash+point.Holdings)
	}
}

func TestPortfolioHoldingsUpdateWithClose(t *testing.T) {
	engine := NewEngine(testConfig(), testRiskEngine())

	engine.RunWithSignals(
		[]exchanges.Candle{
			testCandleAt(1, "100", "100"),
			testCandleAt(2, "110", "110"),
		},
		testSignalsWithSignal([]int64{1, 2}, []float64{1, 1}, []float64{1, 0}),
	)

	point := engine.EquityCurve.Points[1]
	if point.Cash != 9900 {
		t.Fatalf("cash = %f, want 9900", point.Cash)
	}
	if point.Holdings != 110 {
		t.Fatalf("holdings = %f, want 110", point.Holdings)
	}
	if point.TotalEquity != 10010 {
		t.Fatalf("total equity = %f, want 10010", point.TotalEquity)
	}
}

func TestPortfolioCashIncreasesOnSell(t *testing.T) {
	engine := NewEngine(testConfig(), testRiskEngine())

	engine.RunWithSignals(
		[]exchanges.Candle{
			testCandleAt(1, "100", "100"),
			testCandleAt(2, "120", "120"),
		},
		testSignalsWithSignal([]int64{1, 2}, []float64{1, 0}, []float64{1, -1}),
	)

	point := engine.EquityCurve.Points[1]
	if point.Cash != 10020 {
		t.Fatalf("cash = %f, want 10020", point.Cash)
	}
	if point.Holdings != 0 {
		t.Fatalf("holdings = %f, want 0", point.Holdings)
	}
	if point.TotalEquity != 10020 {
		t.Fatalf("total equity = %f, want 10020", point.TotalEquity)
	}
}

func TestEquityCurveRowsMatchCandleCount(t *testing.T) {
	engine := NewEngine(testConfig(), testRiskEngine())

	report := engine.RunWithSignals(
		[]exchanges.Candle{
			testCandleAt(1, "100", "100"),
			testCandleAt(2, "101", "101"),
			testCandleAt(3, "102", "102"),
		},
		testSignalsWithSignal([]int64{1, 2, 3}, []float64{1, 1, 1}, []float64{1, 0, 0}),
	)

	if engine.EquityCurve.Len() != report.CandleCount {
		t.Fatalf("equity rows = %d, want candle count %d", engine.EquityCurve.Len(), report.CandleCount)
	}
}

func TestWriteEquityCSV(t *testing.T) {
	path := t.TempDir() + "/equity.csv"
	curve := EquityCurve{Points: []PortfolioPoint{{
		Time:        1,
		Close:       100,
		Signal:      1,
		Position:    1,
		Holdings:    100,
		Cash:        9900,
		TotalEquity: 10000,
	}}}

	if err := WriteEquityCSV(path, curve); err != nil {
		t.Fatal(err)
	}

	body, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}

	text := string(body)
	if !strings.Contains(text, "time,close,signal,position,holdings,cash,total_equity") {
		t.Fatalf("missing header: %s", text)
	}
	if !strings.Contains(text, "1,100.00000000,1.00000000,1.00000000") {
		t.Fatalf("missing row: %s", text)
	}
}

func TestDrawdownDerivesFromEquityCurve(t *testing.T) {
	curve := EquityCurve{Points: []PortfolioPoint{
		{TotalEquity: 100},
		{TotalEquity: 120},
		{TotalEquity: 90},
	}}

	drawdown := curve.Drawdown(100)
	if drawdown.PeakEquity != 120 {
		t.Fatalf("peak = %f, want 120", drawdown.PeakEquity)
	}
	if drawdown.MaxDrawdown != 30 {
		t.Fatalf("max drawdown = %f, want 30", drawdown.MaxDrawdown)
	}
	if drawdown.MaxDrawdownPct != 25 {
		t.Fatalf("max drawdown pct = %f, want 25", drawdown.MaxDrawdownPct)
	}
}
