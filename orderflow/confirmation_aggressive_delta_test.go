package orderflow

import "testing"

func TestAggressiveDeltaBuySellDetection(t *testing.T) {
	inputs := []AggressiveDeltaInput{
		aggressiveDeltaInput(0, 100, 10),
		aggressiveDeltaInput(1, 101, 1),
		aggressiveDeltaInput(2, 99, -10),
		aggressiveDeltaInput(3, 98, -1),
	}
	got := DetectAggressiveDeltaConfirmations(inputs, AggressiveDeltaConfig{DeltaThreshold: 5})
	if len(got) != 2 {
		t.Fatalf("expected two confirmations, got %d", len(got))
	}
	if !got[0].AggressiveBuy || !got[0].Accepted {
		t.Fatalf("expected accepted buy: %+v", got[0])
	}
	if !got[1].AggressiveSell || !got[1].Accepted {
		t.Fatalf("expected accepted sell: %+v", got[1])
	}
}

func TestAggressiveDeltaBucketAndNeutral(t *testing.T) {
	got := DetectAggressiveDeltaConfirmations([]AggressiveDeltaInput{
		aggressiveDeltaInput(0, 100, 30),
	}, AggressiveDeltaConfig{DeltaThreshold: 10})
	if len(got) != 1 {
		t.Fatalf("expected one confirmation")
	}
	if got[0].DeltaStrengthBucket != "extreme" || !got[0].Neutral {
		t.Fatalf("unexpected bucket/neutral: %+v", got[0])
	}
}

func TestAggressiveDeltaRejection(t *testing.T) {
	got := DetectAggressiveDeltaConfirmations([]AggressiveDeltaInput{
		aggressiveDeltaInput(0, 100, 10),
		aggressiveDeltaInput(1, 99, 1),
	}, AggressiveDeltaConfig{DeltaThreshold: 5})
	if !got[0].Rejected || got[0].Accepted || got[0].Neutral {
		t.Fatalf("expected rejected aggressive buy: %+v", got[0])
	}
}

func aggressiveDeltaInput(timestamp int64, close float64, delta float64) AggressiveDeltaInput {
	return AggressiveDeltaInput{
		Venue: "aster", VenueSymbol: "BTCUSDT", CanonicalSymbol: "BTC",
		Timestamp: timestamp, Open: close - 1, High: close + 1, Low: close - 2, Close: close,
		Delta: delta, Volume: 20,
		HVNPrice: 100, VolumeClusterCount: 1, NearMultipleHVN: true,
		ImbalanceCount: 2, StackedImbalanceCount: 1,
	}
}
