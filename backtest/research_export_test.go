package backtest

import (
	"encoding/json"
	"errors"
	"os"
	"strings"
	"testing"

	"AlgoTrading2026/exchanges"
	"AlgoTrading2026/series"
)

func TestWriteSignalCSVWritesHeaderAndRows(t *testing.T) {
	path := t.TempDir() + "/signals.csv"
	signals := series.SignalResult{
		Times:     []int64{1},
		Close:     []float64{100},
		Diff:      []float64{1},
		Signal:    []float64{1},
		Positions: []float64{1},
	}

	if err := WriteSignalCSV(path, signals); err != nil {
		t.Fatal(err)
	}

	body, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}

	text := string(body)
	if !strings.Contains(text, "time,close,diff,signal,position") {
		t.Fatalf("missing header: %s", text)
	}
	if !strings.Contains(text, "1,100.00000000,1.00000000,1.00000000,1.00000000") {
		t.Fatalf("missing row: %s", text)
	}
}

func TestWriteSummaryJSONWritesValidJSON(t *testing.T) {
	path := t.TempDir() + "/summary.json"
	report := Report{
		Venue:           "aster",
		Symbol:          "BTCUSDT",
		Interval:        "15m",
		CandleCount:     100,
		StartingBalance: 10000,
		EndingBalance:   10001,
		Metrics: Metrics{
			TotalTrades:    1,
			Wins:           1,
			WinRate:        100,
			GrossPnL:       1,
			NetPnL:         1,
			PeakEquity:     10001,
			MaxDrawdownPct: 0,
		},
		CSVPath:       "research/backtest_trades.csv",
		EquityCSVPath: "research/equity_curve.csv",
		SignalCSVPath: "research/signals.csv",
	}

	if err := WriteSummaryJSON(path, report); err != nil {
		t.Fatal(err)
	}

	body, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}

	var parsed SummaryJSON
	if err := json.Unmarshal(body, &parsed); err != nil {
		t.Fatal(err)
	}
	if parsed.Venue != "aster" || parsed.SignalCSV != "research/signals.csv" {
		t.Fatalf("unexpected summary: %+v", parsed)
	}
}

func TestVenueComparisonWriterWritesRows(t *testing.T) {
	path := t.TempDir() + "/comparison.csv"
	results := []VenueComparisonResult{{
		Venue:  "aster",
		Symbol: "BTCUSDT",
		Report: Report{
			Venue:         "aster",
			Symbol:        "BTCUSDT",
			Interval:      "15m",
			CandleCount:   100,
			EndingBalance: 10000,
			Metrics:       Metrics{TotalTrades: 1},
		},
	}}

	if err := WriteVenueComparisonCSV(path, results); err != nil {
		t.Fatal(err)
	}

	body, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}

	text := string(body)
	if !strings.Contains(text, "venue,symbol,interval,candles,trades") {
		t.Fatalf("missing header: %s", text)
	}
	if !strings.Contains(text, "aster,BTCUSDT,15m,100,1") {
		t.Fatalf("missing row: %s", text)
	}
}

func TestRunVenueComparisonSkipsFailedVenue(t *testing.T) {
	results := RunVenueComparison([]VenueComparisonInput{
		{
			Venue:    "bad",
			Symbol:   "BTC",
			Interval: "15m",
			Limit:    100,
			Reader:   failingReader{},
		},
		{
			Venue:    "good",
			Symbol:   "BTC",
			Interval: "15m",
			Limit:    100,
			Reader: fakeReader{candles: []exchanges.Candle{
				testCandleAt(1, "100", "100"),
				testCandleAt(2, "100", "101"),
			}},
			Config: testConfig(),
		},
	})

	if len(results) != 2 {
		t.Fatalf("results = %d, want 2", len(results))
	}
	if results[0].Error == "" {
		t.Fatal("expected first venue error")
	}
	if results[1].Error != "" {
		t.Fatalf("expected second venue success, got %s", results[1].Error)
	}
	if results[1].Report.CandleCount != 2 {
		t.Fatalf("candle count = %d, want 2", results[1].Report.CandleCount)
	}
}

type fakeReader struct {
	candles []exchanges.Candle
}

func (r fakeReader) GetCandles(string, string, int) ([]exchanges.Candle, error) {
	return r.candles, nil
}

type failingReader struct{}

func (failingReader) GetCandles(string, string, int) ([]exchanges.Candle, error) {
	return nil, errors.New("venue unavailable")
}
