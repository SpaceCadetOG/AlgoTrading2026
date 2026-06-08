package orderflow

import "testing"

func TestBullishAbsorptionDetection(t *testing.T) {
	inputs := []BigLimitOrderInput{
		bigLimitInput(0, 100, 90, 99, -10),
		bigLimitInput(1, 101, 95, 100, 1),
	}
	got := DetectBigLimitOrderConfirmations(inputs, DefaultBigLimitOrderConfig())
	if len(got) != 1 {
		t.Fatalf("expected one confirmation, got %d", len(got))
	}
	if !got[0].BullishAbsorption || got[0].BearishAbsorption || !got[0].Accepted || got[0].Rejected {
		t.Fatalf("unexpected bullish confirmation: %+v", got[0])
	}
}

func TestBearishAbsorptionDetection(t *testing.T) {
	inputs := []BigLimitOrderInput{
		bigLimitInput(0, 100, 90, 91, 10),
		bigLimitInput(1, 95, 89, 90, -1),
	}
	got := DetectBigLimitOrderConfirmations(inputs, DefaultBigLimitOrderConfig())
	if len(got) != 1 {
		t.Fatalf("expected one confirmation, got %d", len(got))
	}
	if !got[0].BearishAbsorption || got[0].BullishAbsorption || !got[0].Accepted || got[0].Rejected {
		t.Fatalf("unexpected bearish confirmation: %+v", got[0])
	}
}

func TestBigLimitOrderRejection(t *testing.T) {
	bullRejected := DetectBigLimitOrderConfirmations([]BigLimitOrderInput{
		bigLimitInput(0, 100, 90, 99, -10),
		bigLimitInput(1, 100, 90, 98, 1),
	}, DefaultBigLimitOrderConfig())[0]
	if bullRejected.Accepted || !bullRejected.Rejected {
		t.Fatalf("expected bullish rejection: %+v", bullRejected)
	}
	bearRejected := DetectBigLimitOrderConfirmations([]BigLimitOrderInput{
		bigLimitInput(0, 100, 90, 91, 10),
		bigLimitInput(1, 100, 90, 92, 1),
	}, DefaultBigLimitOrderConfig())[0]
	if bearRejected.Accepted || !bearRejected.Rejected {
		t.Fatalf("expected bearish rejection: %+v", bearRejected)
	}
}

func TestBigLimitOrderFollowThrough(t *testing.T) {
	inputs := []BigLimitOrderInput{bigLimitInput(0, 100, 90, 99, -10)}
	for i := 1; i <= 20; i++ {
		inputs = append(inputs, bigLimitInput(int64(i), 120, 90, 99+float64(i), 1))
	}
	got := DetectBigLimitOrderConfirmations(inputs, DefaultBigLimitOrderConfig())[0]
	if got.FollowThrough5 != 5 || got.FollowThrough10 != 10 || got.FollowThrough20 != 20 {
		t.Fatalf("unexpected follow-through: %+v", got)
	}
}

func bigLimitInput(timestamp int64, high float64, low float64, close float64, delta float64) BigLimitOrderInput {
	return BigLimitOrderInput{
		Venue: "aster", VenueSymbol: "BTCUSDT", CanonicalSymbol: "BTC",
		Timestamp: timestamp, High: high, Low: low, Close: close,
		Delta: delta, Volume: 20,
		AggressiveBuyVolume:  10,
		AggressiveSellVolume: 10,
		NearHVN:              true, NearVolumeCluster: true, NearMultipleHVN: true,
		ImbalanceCount: 2, StackedImbalanceCount: 1,
	}
}
