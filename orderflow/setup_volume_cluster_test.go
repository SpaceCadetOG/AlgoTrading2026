package orderflow

import "testing"

func TestDetectVolumeClusterSetupsLongAccepted(t *testing.T) {
	inputs := []VolumeClusterInput{
		clusterInput(0, 100, 20, 100, 99, 101),
		clusterInput(1, 100, 10, 103, 99.9, 104),
		clusterInput(2, 100, 10, 101, 99.95, 102),
		clusterInput(3, 100, 10, 105, 104, 106),
	}
	setups := DetectVolumeClusterSetups(inputs, DefaultVolumeClusterSetupConfig())
	if len(setups) != 1 {
		t.Fatalf("setups=%d want 1", len(setups))
	}
	got := setups[0]
	if !got.LongContext || got.ShortContext {
		t.Fatalf("expected long context: %+v", got)
	}
	if !got.Retested || !got.Accepted || got.Rejected {
		t.Fatalf("expected accepted retest: %+v", got)
	}
}

func TestDetectVolumeClusterSetupsShortRejected(t *testing.T) {
	inputs := []VolumeClusterInput{
		clusterInput(0, 100, 20, 100, 99, 101),
		clusterInput(1, 100, 10, 97, 96, 99.5),
		clusterInput(2, 100, 10, 101, 99.95, 102),
	}
	setups := DetectVolumeClusterSetups(inputs, DefaultVolumeClusterSetupConfig())
	if len(setups) != 1 {
		t.Fatalf("setups=%d want 1", len(setups))
	}
	got := setups[0]
	if !got.ShortContext || got.LongContext {
		t.Fatalf("expected short context: %+v", got)
	}
	if !got.Retested || got.Accepted || !got.Rejected {
		t.Fatalf("expected rejected short: %+v", got)
	}
}

func TestDetectVolumeClusterSetupsFollowThrough(t *testing.T) {
	inputs := []VolumeClusterInput{clusterInput(0, 100, 20, 100, 99, 101), clusterInput(1, 100, 10, 101, 100, 101)}
	for i := 2; i <= 20; i++ {
		inputs = append(inputs, clusterInput(int64(i), 100, 10, 100+float64(i), 100, 101+float64(i)))
	}
	setups := DetectVolumeClusterSetups(inputs, DefaultVolumeClusterSetupConfig())
	got := setups[0]
	if got.FollowThrough5 != 5 || got.FollowThrough10 != 10 || got.FollowThrough20 != 20 {
		t.Fatalf("unexpected follow-through: %+v", got)
	}
}

func TestDetectVolumeClusterSetupsRequiresThreshold(t *testing.T) {
	inputs := []VolumeClusterInput{clusterInput(0, 100, 1, 100, 99, 101)}
	setups := DetectVolumeClusterSetups(inputs, DefaultVolumeClusterSetupConfig())
	if len(setups) != 0 {
		t.Fatalf("expected no setups: %+v", setups)
	}
}

func clusterInput(timestamp int64, price float64, hvnVolume float64, close float64, low float64, high float64) VolumeClusterInput {
	return VolumeClusterInput{
		Venue: "aster", VenueSymbol: "BTCUSDT", CanonicalSymbol: "BTC",
		StartTimestamp: timestamp,
		High:           high, Low: low, Close: close,
		Volume: 100, Delta: 5, HVNPrice: price, HVNVolume: hvnVolume,
		VolumeClusterCount: 1, ImbalanceCount: 2, StackedImbalanceCount: 1, LevelCount: 10,
	}
}
