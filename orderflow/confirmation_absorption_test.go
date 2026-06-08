package orderflow

import "testing"

func TestAbsorptionBullishDetection(t *testing.T) {
	inputs := []AbsorptionInput{
		absorptionInput(0, 95, 100, 90, 99, -10),
		absorptionInput(1, 99, 102, 98, 101, 1),
	}
	got := DetectAbsorptionConfirmations(inputs, DefaultAbsorptionConfig())
	if len(got) != 1 {
		t.Fatalf("expected one confirmation, got %d", len(got))
	}
	if !got[0].BullishAbsorption || got[0].BearishAbsorption || !got[0].Accepted || got[0].Rejected || got[0].Neutral {
		t.Fatalf("unexpected bullish absorption: %+v", got[0])
	}
}

func TestAbsorptionBearishDetection(t *testing.T) {
	inputs := []AbsorptionInput{
		absorptionInput(0, 96, 100, 90, 91, 10),
		absorptionInput(1, 91, 92, 88, 89, -1),
	}
	got := DetectAbsorptionConfirmations(inputs, DefaultAbsorptionConfig())
	if len(got) != 1 {
		t.Fatalf("expected one confirmation, got %d", len(got))
	}
	if !got[0].BearishAbsorption || got[0].BullishAbsorption || !got[0].Accepted || got[0].Rejected || got[0].Neutral {
		t.Fatalf("unexpected bearish absorption: %+v", got[0])
	}
}

func TestAbsorptionNoDetectionWhenDeltaWeak(t *testing.T) {
	got := DetectAbsorptionConfirmations([]AbsorptionInput{
		absorptionInput(0, 95, 100, 90, 99, -0.1),
	}, AbsorptionConfig{DeltaThreshold: 1})
	if len(got) != 0 {
		t.Fatalf("expected no weak-delta confirmation, got %+v", got)
	}
}

func TestAbsorptionNoDetectionWhenPriceConfirmsDelta(t *testing.T) {
	got := DetectAbsorptionConfirmations([]AbsorptionInput{
		absorptionInput(0, 99, 100, 90, 91, -10),
		absorptionInput(1, 91, 100, 90, 99, 10),
	}, AbsorptionConfig{DeltaThreshold: 1})
	if len(got) != 0 {
		t.Fatalf("expected no absorption when price follows delta, got %+v", got)
	}
}

func TestAbsorptionCloseLocation(t *testing.T) {
	got := absorptionCloseLocation(absorptionInput(0, 95, 100, 90, 97.5, -10))
	if got != 0.75 {
		t.Fatalf("expected close location 0.75, got %v", got)
	}
}

func TestAbsorptionRejectionAndNeutral(t *testing.T) {
	bullRejected := DetectAbsorptionConfirmations([]AbsorptionInput{
		absorptionInput(0, 95, 100, 90, 99, -10),
		absorptionInput(1, 99, 100, 90, 98, 1),
	}, DefaultAbsorptionConfig())[0]
	if bullRejected.Accepted || !bullRejected.Rejected || bullRejected.Neutral {
		t.Fatalf("expected bullish rejection: %+v", bullRejected)
	}
	bearRejected := DetectAbsorptionConfirmations([]AbsorptionInput{
		absorptionInput(0, 96, 100, 90, 91, 10),
		absorptionInput(1, 91, 100, 90, 92, 1),
	}, DefaultAbsorptionConfig())[0]
	if bearRejected.Accepted || !bearRejected.Rejected || bearRejected.Neutral {
		t.Fatalf("expected bearish rejection: %+v", bearRejected)
	}
	neutral := DetectAbsorptionConfirmations([]AbsorptionInput{
		absorptionInput(0, 95, 100, 90, 99, -10),
	}, DefaultAbsorptionConfig())[0]
	if !neutral.Neutral || neutral.Accepted || neutral.Rejected {
		t.Fatalf("expected insufficient-future neutral: %+v", neutral)
	}
}

func TestAbsorptionFollowThrough(t *testing.T) {
	inputs := []AbsorptionInput{absorptionInput(0, 95, 100, 90, 99, -10)}
	for i := 1; i <= 20; i++ {
		inputs = append(inputs, absorptionInput(int64(i), 99, 130, 90, 99+float64(i), 1))
	}
	got := DetectAbsorptionConfirmations(inputs, DefaultAbsorptionConfig())[0]
	if got.FollowThrough5 != 5 || got.FollowThrough10 != 10 || got.FollowThrough20 != 20 {
		t.Fatalf("unexpected follow through: %+v", got)
	}
}

func absorptionInput(timestamp int64, open float64, high float64, low float64, close float64, delta float64) AbsorptionInput {
	return AbsorptionInput{
		Venue: "aster", VenueSymbol: "BTCUSDT", CanonicalSymbol: "BTC",
		Timestamp: timestamp, Open: open, High: high, Low: low, Close: close,
		Delta: delta, Volume: 20,
		HVNPrice: 95, HVNVolume: 10, VolumeClusterCount: 1, NearMultipleHVN: true,
		ImbalanceCount: 2, StackedImbalanceCount: 1,
	}
}
