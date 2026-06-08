package orderflow

import "testing"

func TestCumulativeDeltaBullishDivergence(t *testing.T) {
	inputs := []CumulativeDeltaDivergenceInput{
		cumulativeInput(0, 100, 95, 98, -10),
		cumulativeInput(1, 101, 94, 97, 20),
		cumulativeInput(2, 102, 96, 99, 1),
	}
	got := DetectCumulativeDeltaDivergences(inputs, CumulativeDeltaDivergenceConfig{Window: 1})
	if len(got) == 0 || !got[0].BullishDivergence || !got[0].Accepted {
		t.Fatalf("expected accepted bullish divergence: %+v", got)
	}
}

func TestCumulativeDeltaBearishDivergence(t *testing.T) {
	inputs := []CumulativeDeltaDivergenceInput{
		cumulativeInput(0, 100, 95, 98, 20),
		cumulativeInput(1, 101, 96, 99, -30),
		cumulativeInput(2, 100, 94, 97, -1),
	}
	got := DetectCumulativeDeltaDivergences(inputs, CumulativeDeltaDivergenceConfig{Window: 1})
	if len(got) == 0 || !got[0].BearishDivergence || !got[0].Accepted {
		t.Fatalf("expected accepted bearish divergence: %+v", got)
	}
}

func TestCumulativeDeltaRollingWindowAndNeutral(t *testing.T) {
	inputs := []CumulativeDeltaDivergenceInput{
		cumulativeInput(0, 100, 95, 98, -10),
		cumulativeInput(1, 99, 94, 97, 20),
	}
	got := DetectCumulativeDeltaDivergences(inputs, CumulativeDeltaDivergenceConfig{Window: 1})
	if len(got) != 1 || !got[0].Neutral {
		t.Fatalf("expected one neutral divergence with insufficient future: %+v", got)
	}
	if got[0].CumulativeDelta != 10 || got[0].PriorCumulativeDelta != -10 {
		t.Fatalf("unexpected cumulative delta values: %+v", got[0])
	}
}

func cumulativeInput(timestamp int64, high float64, low float64, close float64, delta float64) CumulativeDeltaDivergenceInput {
	return CumulativeDeltaDivergenceInput{
		Venue: "aster", VenueSymbol: "BTCUSDT", CanonicalSymbol: "BTC",
		Timestamp: timestamp, High: high, Low: low, Close: close, Delta: delta,
		HVNPrice: 100, VolumeClusterCount: 1, NearMultipleHVN: true,
		ImbalanceCount: 2, StackedImbalanceCount: 1,
	}
}
