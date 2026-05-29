package indicators

import "testing"

func TestSMA(t *testing.T) {
	got := SMA([]float64{1, 2, 3, 4}, 2)
	assertFloat(t, got[1], 1.5)
	assertFloat(t, got[3], 3.5)
}

func TestEMA(t *testing.T) {
	got := EMA([]float64{1, 2, 3}, 2)
	assertFloat(t, got[0], 1)
	if got[2] <= got[1] {
		t.Fatalf("ema should rise with rising values: %v", got)
	}
}

func TestAPO(t *testing.T) {
	values := []float64{1, 2, 3, 4}
	got := APO(values, 2, 3)
	fast := EMA(values, 2)
	slow := EMA(values, 3)
	assertFloat(t, got[3], fast[3]-slow[3])
}

func TestMACDHistogram(t *testing.T) {
	got := MACD([]float64{1, 2, 3, 4, 5}, 2, 3, 2)
	for i := range got.MACD {
		assertFloat(t, got.Histogram[i], got.MACD[i]-got.Signal[i])
	}
}

func TestBollingerBands(t *testing.T) {
	got := BollingerBands([]float64{1, 2, 3}, 3, 2)
	if got.Upper[2] <= got.Middle[2] {
		t.Fatalf("upper = %f, middle = %f", got.Upper[2], got.Middle[2])
	}
	if got.Lower[2] >= got.Middle[2] {
		t.Fatalf("lower = %f, middle = %f", got.Lower[2], got.Middle[2])
	}
}

func TestRSIBounded(t *testing.T) {
	got := RSI([]float64{1, 2, 1, 2, 3, 2, 4}, 3)
	for _, value := range got {
		if value < 0 || value > 100 {
			t.Fatalf("rsi out of bounds: %v", got)
		}
	}
}

func TestStdDevNonNegative(t *testing.T) {
	got := StdDev([]float64{1, 2, 3, 4}, 2)
	for _, value := range got {
		if value < 0 {
			t.Fatalf("stddev negative: %v", got)
		}
	}
}

func TestMomentum(t *testing.T) {
	got := Momentum([]float64{10, 12, 15}, 2)
	assertFloat(t, got[2], 5)
}

func TestSupportResistance(t *testing.T) {
	got := RollingSupportResistance(
		[]float64{10, 12, 11, 15},
		[]float64{9, 8, 10, 11},
		3,
	)
	assertFloat(t, got.Support[3], 8)
	assertFloat(t, got.Resistance[3], 15)
}

func assertFloat(t *testing.T, got float64, want float64) {
	t.Helper()
	if got != want {
		t.Fatalf("got %f, want %f", got, want)
	}
}
