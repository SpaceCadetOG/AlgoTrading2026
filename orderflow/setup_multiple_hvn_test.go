package orderflow

import "testing"

func TestDetectMultipleHVNSetupsCreatesZone(t *testing.T) {
	inputs := []MultipleHVNInput{
		multipleHVNInput(0, 100, 10, 100, 99, 101),
		multipleHVNInput(1, 100.1, 20, 100, 99, 101),
		multipleHVNInput(2, 100.2, 5, 102, 100, 103),
	}
	setups := DetectMultipleHVNSetups(inputs, DefaultMultipleHVNConfig())
	if len(setups) != 1 {
		t.Fatalf("setups=%d want 1: %+v", len(setups), setups)
	}
	got := setups[0]
	if got.HVNCount != 3 || got.ZoneLow != 100 || got.ZoneHigh != 100.2 {
		t.Fatalf("unexpected zone: %+v", got)
	}
	if got.AverageHVNVolume != 35.0/3.0 {
		t.Fatalf("unexpected average volume: %+v", got)
	}
}

func TestDetectMultipleHVNSetupsLongAccepted(t *testing.T) {
	inputs := []MultipleHVNInput{
		multipleHVNInput(0, 100, 10, 100, 99, 101),
		multipleHVNInput(1, 100.1, 20, 100, 99, 101),
		multipleHVNInput(2, 110, 1, 103, 102, 104),
		multipleHVNInput(3, 110, 1, 101, 99.9, 102),
	}
	got := DetectMultipleHVNSetups(inputs, DefaultMultipleHVNConfig())[0]
	if !got.LongContext || got.ShortContext {
		t.Fatalf("expected long context: %+v", got)
	}
	if !got.Retested || !got.Accepted || got.Rejected {
		t.Fatalf("expected long accepted retest: %+v", got)
	}
}

func TestDetectMultipleHVNSetupsShortRejected(t *testing.T) {
	inputs := []MultipleHVNInput{
		multipleHVNInput(0, 100, 10, 100, 99, 101),
		multipleHVNInput(1, 100.1, 20, 100, 99, 101),
		multipleHVNInput(2, 110, 1, 97, 96, 99),
		multipleHVNInput(3, 110, 1, 102, 99.9, 103),
	}
	got := DetectMultipleHVNSetups(inputs, DefaultMultipleHVNConfig())[0]
	if !got.ShortContext || got.LongContext {
		t.Fatalf("expected short context: %+v", got)
	}
	if !got.Retested || got.Accepted || !got.Rejected {
		t.Fatalf("expected short rejected retest: %+v", got)
	}
}

func TestDetectMultipleHVNSetupsFollowThrough(t *testing.T) {
	inputs := []MultipleHVNInput{
		multipleHVNInput(0, 100, 10, 100, 99, 101),
		multipleHVNInput(1, 100, 20, 100, 99, 101),
	}
	for i := 2; i <= 21; i++ {
		inputs = append(inputs, multipleHVNInput(int64(i), 120, 1, 100+float64(i), 100, 101+float64(i)))
	}
	got := DetectMultipleHVNSetups(inputs, DefaultMultipleHVNConfig())[0]
	if got.FollowThrough5 != 6 || got.FollowThrough10 != 11 || got.FollowThrough20 != 21 {
		t.Fatalf("unexpected follow-through: %+v", got)
	}
}

func TestDetectMultipleHVNSetupsRejectsDistantHVNs(t *testing.T) {
	inputs := []MultipleHVNInput{
		multipleHVNInput(0, 100, 10, 100, 99, 101),
		multipleHVNInput(1, 110, 20, 100, 99, 101),
	}
	setups := DetectMultipleHVNSetups(inputs, DefaultMultipleHVNConfig())
	if len(setups) != 0 {
		t.Fatalf("expected no setups: %+v", setups)
	}
}

func multipleHVNInput(timestamp int64, hvnPrice float64, hvnVolume float64, close float64, low float64, high float64) MultipleHVNInput {
	return MultipleHVNInput{
		Venue: "aster", VenueSymbol: "BTCUSDT", CanonicalSymbol: "BTC",
		StartTimestamp: timestamp,
		High:           high, Low: low, Close: close,
		HVNPrice: hvnPrice, HVNVolume: hvnVolume,
		Delta: 5, ImbalanceCount: 2, StackedImbalanceCount: 1,
	}
}
