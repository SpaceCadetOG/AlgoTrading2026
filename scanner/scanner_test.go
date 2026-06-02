package scanner

import "testing"

func TestClassifyAsset(t *testing.T) {
	if ClassifyAsset("BTC") != AssetCrypto {
		t.Fatal("BTC should classify as crypto")
	}
	if ClassifyAsset("XAU") != AssetCommodity {
		t.Fatal("XAU should classify as commodity")
	}
	if ClassifyAsset("SPX") != AssetIndex {
		t.Fatal("SPX should classify as index")
	}
}

func TestRankInstruments(t *testing.T) {
	rows := RankInstruments([]Instrument{
		{Venue: "test", Symbol: "ILLQ", SpreadPct: 0.1},
		{Venue: "test", Symbol: "BTC", SpreadPct: 0.0005, L2Available: true, CandlesAvailable: true, TradesAvailable: true, ProfileReady: true},
	})
	if rows[0].Symbol != "BTC" || rows[0].Rank != 1 || rows[0].ResearchScore <= rows[1].ResearchScore {
		t.Fatalf("unexpected ranking: %+v", rows)
	}
}

func TestMergeUniverseFacts(t *testing.T) {
	merged := MergeUniverseFacts([]Instrument{{Venue: "aster", Symbol: "BTCUSDT"}}, []Instrument{{Venue: "aster", Symbol: "BTCUSDT", Price: 100, L2Available: true}})
	if len(merged) != 1 || merged[0].Price != 100 || !merged[0].L2Available {
		t.Fatalf("merged=%+v", merged)
	}
}
