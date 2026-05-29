package pairs

import (
	"math"
	"os"
	"strings"
	"testing"

	"AlgoTrading2026/exchanges"
)

func TestAlignByTimeMatchesSharedTimestamps(t *testing.T) {
	a := []exchanges.Candle{candle(1, "10"), candle(2, "11")}
	b := []exchanges.Candle{candle(2, "20"), candle(3, "21")}

	got, err := AlignByTime(a, b)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got[0].Time != 2 || got[0].AClose != 11 || got[0].BClose != 20 {
		t.Fatalf("unexpected alignment: %+v", got)
	}
}

func TestPearson(t *testing.T) {
	got := Pearson([]float64{1, 2, 3}, []float64{2, 4, 6})
	if math.Abs(got-1) > 1e-9 {
		t.Fatalf("pearson = %f, want 1", got)
	}
}

func TestSpread(t *testing.T) {
	got := Spread([]float64{5, 7}, []float64{2, 3})
	if got[0] != 3 || got[1] != 4 {
		t.Fatalf("spread = %v, want [3 4]", got)
	}
}

func TestZScoreFinite(t *testing.T) {
	got := ZScore([]float64{1, 2, 3}, 3)
	if math.IsNaN(got[2]) || math.IsInf(got[2], 0) {
		t.Fatalf("zscore not finite: %v", got)
	}
	if got[2] <= 0 {
		t.Fatalf("zscore = %f, want positive", got[2])
	}
}

func TestBuildPairCandidate(t *testing.T) {
	a := []exchanges.Candle{candle(1, "10"), candle(2, "11"), candle(3, "12")}
	b := []exchanges.Candle{candle(1, "20"), candle(2, "22"), candle(3, "24")}

	got, err := BuildPairCandidate("a", "BTC", a, "b", "BTC", b, 2)
	if err != nil {
		t.Fatal(err)
	}
	if got.Samples != 3 || got.LatestSpread != -12 {
		t.Fatalf("unexpected candidate: %+v", got)
	}
}

func TestWriteCandidatesCSV(t *testing.T) {
	path := t.TempDir() + "/pairs.csv"
	rows := []PairCandidate{{VenueA: "aster", SymbolA: "BTCUSDT", VenueB: "hyperliquid", SymbolB: "BTC", Samples: 10}}

	if err := WriteCandidatesCSV(path, rows); err != nil {
		t.Fatal(err)
	}
	body, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	text := string(body)
	if !strings.Contains(text, "venue_a,symbol_a,venue_b,symbol_b") {
		t.Fatalf("missing header: %s", text)
	}
	if !strings.Contains(text, "aster,BTCUSDT,hyperliquid,BTC") {
		t.Fatalf("missing row: %s", text)
	}
}

func candle(t int64, close string) exchanges.Candle {
	return exchanges.Candle{StartTime: t, Close: close}
}
