package orderflow

import "testing"

func TestStackedBuyImbalanceCreatesLongContext(t *testing.T) {
	inputs := []StackedImbalanceInput{
		stackedInput(0, []FootprintLevel{
			{Price: 100, BidVolume: 1, AskVolume: 5},
			{Price: 101, BidVolume: 1, AskVolume: 5},
			{Price: 102, BidVolume: 1, AskVolume: 5},
		}, 100, 99, 103, 10),
		stackedInput(1, []FootprintLevel{{Price: 103, BidVolume: 1, AskVolume: 1}}, 103, 101, 104, 1),
	}
	got := DetectStackedImbalanceSetups(inputs, nil, DefaultStackedImbalanceConfig())
	if len(got) != 1 {
		t.Fatalf("expected one setup, got %d", len(got))
	}
	if got[0].ImbalanceDirection != "BUY" || !got[0].LongContext || got[0].ShortContext {
		t.Fatalf("unexpected buy setup: %+v", got[0])
	}
	if got[0].ZoneLow != 100 || got[0].ZoneHigh != 102 || got[0].StackedLevels != 3 {
		t.Fatalf("unexpected zone: %+v", got[0])
	}
}

func TestStackedSellImbalanceCreatesShortContext(t *testing.T) {
	inputs := []StackedImbalanceInput{
		stackedInput(0, []FootprintLevel{
			{Price: 100, BidVolume: 5, AskVolume: 1},
			{Price: 101, BidVolume: 5, AskVolume: 1},
			{Price: 102, BidVolume: 5, AskVolume: 1},
		}, 100, 99, 103, -10),
		stackedInput(1, []FootprintLevel{{Price: 99, BidVolume: 1, AskVolume: 1}}, 99, 98, 101, 1),
	}
	got := DetectStackedImbalanceSetups(inputs, nil, DefaultStackedImbalanceConfig())
	if len(got) != 1 {
		t.Fatalf("expected one setup, got %d", len(got))
	}
	if got[0].ImbalanceDirection != "SELL" || !got[0].ShortContext || got[0].LongContext {
		t.Fatalf("unexpected sell setup: %+v", got[0])
	}
}

func TestStackedImbalanceAcceptanceRejectionNeutral(t *testing.T) {
	longAccepted := DetectStackedImbalanceSetups([]StackedImbalanceInput{
		stackedBuyInput(0, 100),
		stackedInput(1, nil, 103, 101, 104, 1),
	}, nil, DefaultStackedImbalanceConfig())[0]
	if !longAccepted.Accepted || longAccepted.Rejected {
		t.Fatalf("expected long accepted: %+v", longAccepted)
	}
	longRejected := DetectStackedImbalanceSetups([]StackedImbalanceInput{
		stackedBuyInput(0, 100),
		stackedInput(1, nil, 99, 98, 101, 1),
	}, nil, DefaultStackedImbalanceConfig())[0]
	if longRejected.Accepted || !longRejected.Rejected {
		t.Fatalf("expected long rejected: %+v", longRejected)
	}
	shortAccepted := DetectStackedImbalanceSetups([]StackedImbalanceInput{
		stackedSellInput(0, 100),
		stackedInput(1, nil, 99, 98, 101, 1),
	}, nil, DefaultStackedImbalanceConfig())[0]
	if !shortAccepted.Accepted || shortAccepted.Rejected {
		t.Fatalf("expected short accepted: %+v", shortAccepted)
	}
	shortRejected := DetectStackedImbalanceSetups([]StackedImbalanceInput{
		stackedSellInput(0, 100),
		stackedInput(1, nil, 103, 101, 104, 1),
	}, nil, DefaultStackedImbalanceConfig())[0]
	if shortRejected.Accepted || !shortRejected.Rejected {
		t.Fatalf("expected short rejected: %+v", shortRejected)
	}
	neutral := DetectStackedImbalanceSetups([]StackedImbalanceInput{stackedBuyInput(0, 100)}, nil, DefaultStackedImbalanceConfig())[0]
	if neutral.Accepted || neutral.Rejected {
		t.Fatalf("expected neutral without future bar: %+v", neutral)
	}
}

func TestStackedImbalanceRetestAndConfluence(t *testing.T) {
	inputs := []StackedImbalanceInput{
		stackedBuyInput(0, 100),
		stackedInput(1, nil, 104, 100.5, 105, 1),
	}
	zones := []StackedImbalanceZone{{
		Venue: "aster", VenueSymbol: "BTCUSDT", CanonicalSymbol: "BTC",
		ZoneLow: 99.9, ZoneHigh: 102.1,
	}}
	got := DetectStackedImbalanceSetups(inputs, zones, DefaultStackedImbalanceConfig())
	if len(got) != 1 {
		t.Fatalf("expected setup")
	}
	if !got[0].Retested || !got[0].NearHVN || !got[0].NearVolumeCluster || !got[0].NearMultipleHVN {
		t.Fatalf("expected retest and confluence: %+v", got[0])
	}
}

func TestStackedImbalanceFollowThrough(t *testing.T) {
	inputs := []StackedImbalanceInput{stackedBuyInput(0, 100)}
	for i := 1; i <= 20; i++ {
		inputs = append(inputs, stackedInput(int64(i), nil, 101+float64(i), 100, 125, 1))
	}
	got := DetectStackedImbalanceSetups(inputs, nil, DefaultStackedImbalanceConfig())[0]
	if got.FollowThrough5 != 5 || got.FollowThrough10 != 10 || got.FollowThrough20 != 20 {
		t.Fatalf("unexpected follow through: %+v", got)
	}
}

func stackedBuyInput(timestamp int64, price float64) StackedImbalanceInput {
	return stackedInput(timestamp, []FootprintLevel{
		{Price: price, BidVolume: 1, AskVolume: 5},
		{Price: price + 1, BidVolume: 1, AskVolume: 5},
		{Price: price + 2, BidVolume: 1, AskVolume: 5},
	}, price, price-1, price+3, 10)
}

func stackedSellInput(timestamp int64, price float64) StackedImbalanceInput {
	return stackedInput(timestamp, []FootprintLevel{
		{Price: price, BidVolume: 5, AskVolume: 1},
		{Price: price + 1, BidVolume: 5, AskVolume: 1},
		{Price: price + 2, BidVolume: 5, AskVolume: 1},
	}, price, price-1, price+3, -10)
}

func stackedInput(timestamp int64, levels []FootprintLevel, close float64, low float64, high float64, delta float64) StackedImbalanceInput {
	return StackedImbalanceInput{
		Venue: "aster", VenueSymbol: "BTCUSDT", CanonicalSymbol: "BTC",
		StartTimestamp: timestamp, EndTimestamp: timestamp,
		High: high, Low: low, Close: close, Levels: levels,
		BarDelta: delta, BarVolume: 30, HVNPrice: 101, HVNVolume: 10, VolumeClusterCount: 1,
	}
}
