package vwap

import (
	"math"
	"testing"

	"AlgoTrading2026/exchanges"
)

func TestSessionVWAP(t *testing.T) {
	candles := []exchanges.Candle{
		testCandle(0, "10", "10", "10", "10", "2"),
		testCandle(1, "20", "20", "20", "20", "1"),
	}
	got := SessionVWAP(candles)
	assertClose(t, got[0], 10)
	assertClose(t, got[1], 40.0/3.0)
}

func TestAnchoredVWAP(t *testing.T) {
	candles := []exchanges.Candle{
		testCandle(0, "10", "10", "10", "10", "2"),
		testCandle(1, "20", "20", "20", "20", "1"),
		testCandle(2, "30", "30", "30", "30", "1"),
	}
	got := AnchoredVWAP(candles, 1)
	assertClose(t, got[0], 0)
	assertClose(t, got[1], 20)
	assertClose(t, got[2], 25)
}

func TestDistanceAndSlope(t *testing.T) {
	dist := Distance([]float64{11, 9}, []float64{10, 10})
	assertClose(t, dist[0], 1)
	assertClose(t, dist[1], -1)

	pct := DistancePct([]float64{11}, []float64{10})
	assertClose(t, pct[0], 0.1)

	slope := Slope([]float64{1, 2, 4}, 1)
	assertClose(t, slope[0], 0)
	assertClose(t, slope[1], 1)
	assertClose(t, slope[2], 2)
}

func TestRegimes(t *testing.T) {
	if Classify(11, 10, 1) != RegimeAboveRising {
		t.Fatal("expected above rising")
	}
	if Classify(9, 10, -1) != RegimeBelowFalling {
		t.Fatal("expected below falling")
	}
	if Classify(10, 10, 1) != RegimeNeutral {
		t.Fatal("expected neutral")
	}
}

func testCandle(t int64, open string, high string, low string, close string, volume string) exchanges.Candle {
	return exchanges.Candle{
		Open:      open,
		High:      high,
		Low:       low,
		Close:     close,
		Volume:    volume,
		StartTime: t,
		EndTime:   t + 1,
	}
}

func assertClose(t *testing.T, got float64, want float64) {
	t.Helper()
	if math.Abs(got-want) > 1e-9 {
		t.Fatalf("got %.12f want %.12f", got, want)
	}
}
