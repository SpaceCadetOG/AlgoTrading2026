package volumeprofile

import "testing"

func TestAccumulationLongAfterBullishInitiation(t *testing.T) {
	profiles := []AccumulationProfileInput{accumulationProfile("a1", 1, 2, 100, 110, 90, AcceptanceAccepted)}
	points := []AccumulationPricePoint{
		{Timestamp: 1, High: 101, Low: 99, Close: 100},
		{Timestamp: 3, High: 120, Low: 111, Close: 118, InitiationDetected: true, InitiationDirection: "up"},
		{Timestamp: 4, High: 101, Low: 99, Close: 100},
		{Timestamp: 5, High: 103, Low: 100, Close: 102},
		{Timestamp: 6, High: 104, Low: 101, Close: 103},
	}
	setups := DetectAccumulationSetups(profiles, points, ScopedPOCLevels{Daily: 100.1}, AccumulationSetupConfig{InitiationLookahead: 5, RetestLookahead: 5})
	if len(setups) != 1 {
		t.Fatalf("setups=%+v", setups)
	}
	setup := setups[0]
	if setup.Direction != AccumulationDirectionLong || !setup.RetestDetected || !setup.DailyPOCConfluence || !setup.Accepted {
		t.Fatalf("setup=%+v", setup)
	}
	if setup.FollowThrough5 <= 0 {
		t.Fatalf("expected positive follow-through: %+v", setup)
	}
}

func TestAccumulationShortAfterBearishInitiation(t *testing.T) {
	profiles := []AccumulationProfileInput{accumulationProfile("a1", 1, 2, 100, 110, 90, AcceptanceRejected)}
	points := []AccumulationPricePoint{
		{Timestamp: 1, High: 101, Low: 99, Close: 100},
		{Timestamp: 3, High: 89, Low: 80, Close: 82, InitiationDetected: true, InitiationDirection: "down"},
		{Timestamp: 4, High: 101, Low: 99, Close: 100},
		{Timestamp: 5, High: 99, Low: 96, Close: 97},
	}
	setups := DetectAccumulationSetups(profiles, points, ScopedPOCLevels{}, AccumulationSetupConfig{InitiationLookahead: 5, RetestLookahead: 5})
	if len(setups) != 1 {
		t.Fatalf("setups=%+v", setups)
	}
	if setups[0].Direction != AccumulationDirectionShort || !setups[0].Rejected {
		t.Fatalf("setup=%+v", setups[0])
	}
}

func TestAccumulationInvalidation(t *testing.T) {
	longProfile := []AccumulationProfileInput{accumulationProfile("long", 1, 2, 100, 110, 90, AcceptanceNeutral)}
	longPoints := []AccumulationPricePoint{
		{Timestamp: 3, High: 120, Low: 111, Close: 118, InitiationDetected: true, InitiationDirection: "up"},
		{Timestamp: 4, High: 101, Low: 99, Close: 89},
	}
	longSetups := DetectAccumulationSetups(longProfile, longPoints, ScopedPOCLevels{}, AccumulationSetupConfig{InitiationLookahead: 5, RetestLookahead: 5, InvalidationPct: 0.001})
	if len(longSetups) != 1 || !longSetups[0].Invalidated {
		t.Fatalf("expected long invalidation: %+v", longSetups)
	}

	shortProfile := []AccumulationProfileInput{accumulationProfile("short", 1, 2, 100, 110, 90, AcceptanceNeutral)}
	shortPoints := []AccumulationPricePoint{
		{Timestamp: 3, High: 89, Low: 80, Close: 82, InitiationDetected: true, InitiationDirection: "down"},
		{Timestamp: 4, High: 101, Low: 99, Close: 111},
	}
	shortSetups := DetectAccumulationSetups(shortProfile, shortPoints, ScopedPOCLevels{}, AccumulationSetupConfig{InitiationLookahead: 5, RetestLookahead: 5, InvalidationPct: 0.001})
	if len(shortSetups) != 1 || !shortSetups[0].Invalidated {
		t.Fatalf("expected short invalidation: %+v", shortSetups)
	}
}

func TestNearPOC(t *testing.T) {
	if !NearPOC(100, 100.2, 0.0025) {
		t.Fatalf("expected near")
	}
	if NearPOC(100, 101, 0.0025) {
		t.Fatalf("expected not near")
	}
}

func accumulationProfile(id string, start, end int64, poc, vah, val float64, state string) AccumulationProfileInput {
	return AccumulationProfileInput{
		ProfileID:       id,
		StartTime:       start,
		EndTime:         end,
		Symbol:          "BTCUSDT",
		POC:             poc,
		VAH:             vah,
		VAL:             val,
		ProfileShape:    ShapeDProfile,
		ShapeConfidence: 0.8,
		ProfileVolume:   1000,
		AcceptanceState: state,
	}
}
