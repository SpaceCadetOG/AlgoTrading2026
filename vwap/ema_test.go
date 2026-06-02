package vwap

import "testing"

func TestEMA(t *testing.T) {
	got := EMA([]float64{10, 20, 30}, 2)
	assertClose(t, got[0], 10)
	assertClose(t, got[1], 16.666666666666668)
	assertClose(t, got[2], 25.555555555555557)
}

func TestEMARelationships(t *testing.T) {
	rows := EMARelationships([]float64{10, 11, 12, 13, 14})
	if len(rows) != 5 {
		t.Fatalf("rows=%d want 5", len(rows))
	}
	if rows[4].Alignment != EMAAlignmentBullish {
		t.Fatalf("alignment=%s want bullish", rows[4].Alignment)
	}
	if ClassifyEMAAlignment(9, 10, 11) != EMAAlignmentBearish {
		t.Fatal("expected bearish alignment")
	}
	if ClassifyEMAAlignment(10, 9, 11) != EMAAlignmentMixed {
		t.Fatal("expected mixed alignment")
	}
}
