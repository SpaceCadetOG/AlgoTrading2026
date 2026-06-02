package labels

import "testing"

func TestSetupBpsNormalization(t *testing.T) {
	got := Bps(2, 100)
	if got != 200 {
		t.Fatalf("bps = %f, want 200", got)
	}
}

func TestSetupATRNormalization(t *testing.T) {
	got := ATRMultiple(4, 2)
	if got != 2 {
		t.Fatalf("atr multiple = %f, want 2", got)
	}
}

func TestAcceptanceLabel(t *testing.T) {
	if AcceptanceLabel(true, false) != AcceptanceAccepted {
		t.Fatal("expected accepted")
	}
	if AcceptanceLabel(false, true) != AcceptanceRejected {
		t.Fatal("expected rejected")
	}
	if AcceptanceLabel(false, false) != AcceptanceNeutral {
		t.Fatal("expected neutral")
	}
}

func TestDirectionalLabel20(t *testing.T) {
	if DirectionalLabel20(1) != DirectionalFavorable {
		t.Fatal("expected favorable")
	}
	if DirectionalLabel20(-1) != DirectionalUnfavorable {
		t.Fatal("expected unfavorable")
	}
	if DirectionalLabel20(0) != DirectionalFlat {
		t.Fatal("expected flat")
	}
}

func TestTripleBarrierUpperHit(t *testing.T) {
	got := TripleBarrierLabel(
		[]float64{100, 103},
		[]float64{99, 100},
		[]float64{100, 102},
		0,
		2,
		2,
		"long_context",
	)
	if got != TripleBarrierUpperHit {
		t.Fatalf("triple barrier = %s, want upper hit", got)
	}
}

func TestTripleBarrierLowerHit(t *testing.T) {
	got := TripleBarrierLabel(
		[]float64{100, 101},
		[]float64{99, 97},
		[]float64{100, 98},
		0,
		2,
		2,
		"long_context",
	)
	if got != TripleBarrierLowerHit {
		t.Fatalf("triple barrier = %s, want lower hit", got)
	}
}

func TestTripleBarrierTimeExpired(t *testing.T) {
	got := TripleBarrierLabel(
		[]float64{100, 101},
		[]float64{99, 99},
		[]float64{100, 100},
		0,
		1,
		5,
		"long_context",
	)
	if got != TripleBarrierTimeExpired {
		t.Fatalf("triple barrier = %s, want time expired", got)
	}
}
