package volatility

import (
	"math"
	"testing"
)

func TestTrueRange(t *testing.T) {
	got := TrueRange(
		[]float64{10, 12},
		[]float64{8, 9},
		[]float64{9, 11},
	)

	assertFloat(t, got[0], 2)
	assertFloat(t, got[1], 3)
}

func TestATR(t *testing.T) {
	got := ATR(
		[]float64{10, 12, 14},
		[]float64{8, 9, 10},
		[]float64{9, 11, 13},
		2,
	)

	if len(got) != 3 {
		t.Fatalf("len = %d, want 3", len(got))
	}
	assertFloat(t, got[1], 2.5)
}

func TestLogReturns(t *testing.T) {
	got := LogReturns([]float64{100, 110})
	want := math.Log(1.1)
	if math.Abs(got[1]-want) > 1e-12 {
		t.Fatalf("log return = %f, want %f", got[1], want)
	}
}

func TestRealizedVolatilityNonNegative(t *testing.T) {
	got := RealizedVolatility([]float64{100, 101, 99, 102}, 2)
	for _, value := range got {
		if value < 0 || math.IsNaN(value) || math.IsInf(value, 0) {
			t.Fatalf("invalid realized vol: %v", got)
		}
	}
}

func TestNormalizeByVolatilityZeroVol(t *testing.T) {
	got := NormalizeByVolatility([]float64{10, 10}, []float64{2, 0})
	assertFloat(t, got[0], 5)
	assertFloat(t, got[1], 0)
}

func TestClassifyRegime(t *testing.T) {
	got := ClassifyRegime([]float64{1, 1, 1, 0.5, 2}, 3, 0.75, 1.25)
	if got[3] != RegimeLow {
		t.Fatalf("regime[3] = %s, want low; all=%v", got[3], got)
	}
	if got[4] != RegimeHigh {
		t.Fatalf("regime[4] = %s, want high; all=%v", got[4], got)
	}
}

func assertFloat(t *testing.T, got float64, want float64) {
	t.Helper()
	if math.Abs(got-want) > 1e-12 {
		t.Fatalf("got %f, want %f", got, want)
	}
}
