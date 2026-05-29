package statarb

import (
	"math"
	"os"
	"strings"
	"testing"

	"AlgoTrading2026/exchanges"
)

func TestHedgeRatioKnownValues(t *testing.T) {
	got := HedgeRatio([]float64{2, 4, 6}, []float64{1, 2, 3})
	assertClose(t, got, 2)
}

func TestNormalizedSpread(t *testing.T) {
	got := NormalizedSpread([]float64{10, 12}, []float64{4, 5}, 2)
	if got[0] != 2 || got[1] != 2 {
		t.Fatalf("spread = %v, want [2 2]", got)
	}
}

func TestRollingZScoreZeroStdDev(t *testing.T) {
	got := RollingZScore([]float64{1, 1, 1}, 3)
	if got[2] != 0 {
		t.Fatalf("zscore = %v, want zero at constant series", got)
	}
}

func TestCandidateNotesAndScore(t *testing.T) {
	candidate := Candidate{
		Samples:      40,
		Correlation:  0.4,
		LatestZScore: 2.5,
		AbsZScore:    2.5,
	}
	candidate.Score = math.Abs(candidate.Correlation) * candidate.AbsZScore
	candidate.Notes = candidateNotes(candidate)

	assertClose(t, candidate.Score, 1.0)
	notes := strings.Join(candidate.Notes, ",")
	for _, want := range []string{"insufficient_samples", "weak_correlation", "interesting_divergence"} {
		if !strings.Contains(notes, want) {
			t.Fatalf("notes %q missing %s", notes, want)
		}
	}
}

func TestBuildCandidate(t *testing.T) {
	x := []exchanges.Candle{candle(1, "10"), candle(2, "12"), candle(3, "14"), candle(4, "16")}
	y := []exchanges.Candle{candle(1, "5"), candle(2, "6"), candle(3, "7"), candle(4, "8")}

	got, err := BuildCandidate("a", "BTC", x, "b", "BTC", y, 2)
	if err != nil {
		t.Fatal(err)
	}
	if got.Samples != 4 {
		t.Fatalf("samples = %d, want 4", got.Samples)
	}
	assertClose(t, got.HedgeRatio, 2)
}

func TestWriteCandidatesCSV(t *testing.T) {
	path := t.TempDir() + "/statarb.csv"
	rows := []Candidate{{
		VenueA:      "aster",
		SymbolA:     "BTCUSDT",
		VenueB:      "hyperliquid",
		SymbolB:     "BTC",
		Samples:     100,
		Correlation: 0.8,
		HedgeRatio:  1.1,
		Notes:       []string{"interesting_divergence"},
	}}

	if err := WriteCandidatesCSV(path, rows); err != nil {
		t.Fatal(err)
	}
	body, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	text := string(body)
	if !strings.Contains(text, "venue_a,symbol_a,venue_b,symbol_b,samples,correlation,hedge_ratio") {
		t.Fatalf("missing header: %s", text)
	}
	if !strings.Contains(text, "aster,BTCUSDT,hyperliquid,BTC") {
		t.Fatalf("missing row: %s", text)
	}
}

func candle(t int64, close string) exchanges.Candle {
	return exchanges.Candle{StartTime: t, Close: close}
}

func assertClose(t *testing.T, got float64, want float64) {
	t.Helper()
	if math.Abs(got-want) > 1e-9 {
		t.Fatalf("got %f, want %f", got, want)
	}
}
