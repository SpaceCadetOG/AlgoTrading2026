package features

import "testing"

func TestSetupDistancePct(t *testing.T) {
	got := DistancePct(105, 100)
	if got != 0.05 {
		t.Fatalf("distance pct = %f, want 0.05", got)
	}
	if DistancePct(100, 0) != 0 {
		t.Fatal("zero level should return 0")
	}
}

func TestSetupATR(t *testing.T) {
	atr := ATR(
		[]float64{11, 13, 14},
		[]float64{9, 10, 12},
		[]float64{10, 12, 13},
		2,
	)
	if atr[0] != 0 {
		t.Fatalf("warmup atr = %f, want 0", atr[0])
	}
	if atr[1] != 2.5 {
		t.Fatalf("atr[1] = %f, want 2.5", atr[1])
	}
}

func TestVolatilityState(t *testing.T) {
	if VolatilityState(0.001) != VolatilityLow {
		t.Fatal("expected low volatility")
	}
	if VolatilityState(0.02) != VolatilityHigh {
		t.Fatal("expected high volatility")
	}
	if VolatilityState(0.005) != VolatilityNormal {
		t.Fatal("expected normal volatility")
	}
}

func TestInValueArea(t *testing.T) {
	if !InValueArea(100, 110, 90) {
		t.Fatal("expected price inside value area")
	}
	if InValueArea(120, 110, 90) {
		t.Fatal("expected price outside value area")
	}
}
