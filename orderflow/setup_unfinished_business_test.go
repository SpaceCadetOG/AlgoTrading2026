package orderflow

import "testing"

func TestUnfinishedHighCreatesSetup(t *testing.T) {
	inputs := []UnfinishedBusinessInput{
		unfinishedInput(0, 100, 95, 99, []FootprintLevel{{Price: 100, BidVolume: 1, AskVolume: 1}}),
		unfinishedInput(1, 101, 99, 99, nil),
	}
	got := DetectUnfinishedBusinessSetups(inputs, DefaultUnfinishedBusinessConfig())
	if len(got) != 1 {
		t.Fatalf("expected 1 setup, got %d", len(got))
	}
	if !got[0].UnfinishedHigh || got[0].Location != "HIGH" || !got[0].MagnetContext || !got[0].ShortContext {
		t.Fatalf("unexpected high setup: %+v", got[0])
	}
}

func TestUnfinishedLowCreatesSetup(t *testing.T) {
	inputs := []UnfinishedBusinessInput{
		unfinishedInput(0, 100, 95, 96, []FootprintLevel{{Price: 95, BidVolume: 1, AskVolume: 1}}),
		unfinishedInput(1, 96, 94, 96, nil),
	}
	got := DetectUnfinishedBusinessSetups(inputs, DefaultUnfinishedBusinessConfig())
	if len(got) != 1 {
		t.Fatalf("expected 1 setup, got %d", len(got))
	}
	if !got[0].UnfinishedLow || got[0].Location != "LOW" || !got[0].MagnetContext || !got[0].LongContext {
		t.Fatalf("unexpected low setup: %+v", got[0])
	}
}

func TestUnfinishedBusinessRevisitAcceptedRejectedNeutral(t *testing.T) {
	highAccepted := DetectUnfinishedBusinessSetups([]UnfinishedBusinessInput{
		unfinishedInput(0, 100, 95, 99, []FootprintLevel{{Price: 100, BidVolume: 1, AskVolume: 1}}),
		unfinishedInput(1, 101, 99, 101, nil),
	}, DefaultUnfinishedBusinessConfig())[0]
	if !highAccepted.Revisited || !highAccepted.Accepted || highAccepted.Rejected || highAccepted.Neutral {
		t.Fatalf("expected high accepted: %+v", highAccepted)
	}
	highRejected := DetectUnfinishedBusinessSetups([]UnfinishedBusinessInput{
		unfinishedInput(0, 100, 95, 99, []FootprintLevel{{Price: 100, BidVolume: 1, AskVolume: 1}}),
		unfinishedInput(1, 101, 99, 99, nil),
	}, DefaultUnfinishedBusinessConfig())[0]
	if !highRejected.Revisited || highRejected.Accepted || !highRejected.Rejected || highRejected.Neutral {
		t.Fatalf("expected high rejected: %+v", highRejected)
	}
	lowAccepted := DetectUnfinishedBusinessSetups([]UnfinishedBusinessInput{
		unfinishedInput(0, 100, 95, 96, []FootprintLevel{{Price: 95, BidVolume: 1, AskVolume: 1}}),
		unfinishedInput(1, 96, 94, 94, nil),
	}, DefaultUnfinishedBusinessConfig())[0]
	if !lowAccepted.Revisited || !lowAccepted.Accepted || lowAccepted.Rejected || lowAccepted.Neutral {
		t.Fatalf("expected low accepted: %+v", lowAccepted)
	}
	lowRejected := DetectUnfinishedBusinessSetups([]UnfinishedBusinessInput{
		unfinishedInput(0, 100, 95, 96, []FootprintLevel{{Price: 95, BidVolume: 1, AskVolume: 1}}),
		unfinishedInput(1, 96, 94, 96, nil),
	}, DefaultUnfinishedBusinessConfig())[0]
	if !lowRejected.Revisited || lowRejected.Accepted || !lowRejected.Rejected || lowRejected.Neutral {
		t.Fatalf("expected low rejected: %+v", lowRejected)
	}
	neutral := DetectUnfinishedBusinessSetups([]UnfinishedBusinessInput{
		unfinishedInput(0, 100, 95, 96, []FootprintLevel{{Price: 95, BidVolume: 1, AskVolume: 1}}),
	}, DefaultUnfinishedBusinessConfig())[0]
	if neutral.Revisited || neutral.Accepted || neutral.Rejected || !neutral.Neutral {
		t.Fatalf("expected neutral: %+v", neutral)
	}
}

func TestUnfinishedBusinessFollowThrough(t *testing.T) {
	inputs := []UnfinishedBusinessInput{
		unfinishedInput(0, 100, 95, 96, []FootprintLevel{{Price: 95, BidVolume: 1, AskVolume: 1}}),
	}
	for i := 1; i <= 20; i++ {
		inputs = append(inputs, unfinishedInput(int64(i), 100, 95, 95+float64(i), nil))
	}
	got := DetectUnfinishedBusinessSetups(inputs, DefaultUnfinishedBusinessConfig())[0]
	if got.FollowThrough5 != 5 || got.FollowThrough10 != 10 || got.FollowThrough20 != 20 {
		t.Fatalf("unexpected follow-through: %+v", got)
	}
}

func unfinishedInput(timestamp int64, high float64, low float64, close float64, levels []FootprintLevel) UnfinishedBusinessInput {
	return UnfinishedBusinessInput{
		Venue: "aster", VenueSymbol: "BTCUSDT", CanonicalSymbol: "BTC",
		StartTimestamp: timestamp, EndTimestamp: timestamp,
		High: high, Low: low, Close: close, Levels: levels,
		BarDelta: 1, BarVolume: 10, HVNPrice: close, HVNVolume: 5,
		ImbalanceCount: 2, StackedImbalanceCount: 1,
	}
}
