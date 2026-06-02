package volumeprofile

import "testing"

func TestFailedHighAuctionCreatesShortReversal(t *testing.T) {
	rejections := []ReversalRejectionInput{{
		Timestamp:      10,
		Direction:      AccumulationDirectionShort,
		POC:            110,
		VAH:            112,
		VAL:            100,
		NearestHVN:     111,
		ProfileShape:   ShapeDProfile,
		VWAPAlignment:  "below_vwap",
		FollowThrough5: 1,
	}}
	auctions := []ReversalFailedAuctionInput{{
		Timestamp:       11,
		Strategy:        FailedHighAuction,
		Level:           110,
		FollowThrough5:  2,
		FollowThrough10: 3,
		FollowThrough20: 4,
	}}
	setups := DetectReversalSetups(rejections, auctions, ReversalSetupConfig{MaxEventDistanceMillis: 10, ConfluenceThresholdPct: 0.01})
	if len(setups) != 1 {
		t.Fatalf("setups=%+v", setups)
	}
	setup := setups[0]
	if setup.Direction != AccumulationDirectionShort || setup.FailedAuctionType != FailedHighAuction || !setup.POCConfluence || !setup.HVNConfluence || setup.FollowThrough20 != 4 || !setup.Accepted || setup.Rejected {
		t.Fatalf("setup=%+v", setup)
	}
}

func TestFailedLowAuctionCreatesLongReversal(t *testing.T) {
	rejections := []ReversalRejectionInput{{
		Timestamp:      20,
		Direction:      AccumulationDirectionLong,
		POC:            90,
		VAH:            100,
		VAL:            88,
		NearestHVN:     91,
		NearestLVN:     89,
		ProfileShape:   ShapeBProfile,
		VWAPAlignment:  "above_vwap",
		RejectionLow:   87,
		FollowThrough5: 1,
	}}
	auctions := []ReversalFailedAuctionInput{{
		Timestamp:       21,
		Strategy:        FailedLowAuction,
		Level:           90,
		FollowThrough20: 5,
	}}
	setups := DetectReversalSetups(rejections, auctions, ReversalSetupConfig{MaxEventDistanceMillis: 10, ConfluenceThresholdPct: 0.02})
	if len(setups) != 1 {
		t.Fatalf("setups=%+v", setups)
	}
	setup := setups[0]
	if setup.Direction != AccumulationDirectionLong || setup.ReversalLevel != 90 || !setup.POCConfluence || setup.NearestLVN != 89 {
		t.Fatalf("setup=%+v", setup)
	}
}

func TestReversalVAHVALAndInvalidation(t *testing.T) {
	rejections := []ReversalRejectionInput{{
		Timestamp:      30,
		Direction:      AccumulationDirectionShort,
		RejectionHigh:  115,
		RejectionLevel: 112,
		POC:            105,
		VAH:            112,
		VAL:            95,
	}}
	auctions := []ReversalFailedAuctionInput{{
		Timestamp:   31,
		Strategy:    FailedHighAuction,
		Level:       112,
		Invalidated: true,
	}}
	setups := DetectReversalSetups(rejections, auctions, ReversalSetupConfig{MaxEventDistanceMillis: 10, ConfluenceThresholdPct: 0.01})
	if len(setups) != 1 {
		t.Fatalf("setups=%+v", setups)
	}
	if !setups[0].VAHVALRejection || !setups[0].Invalidated || !setups[0].Rejected {
		t.Fatalf("setup=%+v", setups[0])
	}
}

func TestReversalOutcomeNeutral(t *testing.T) {
	rejections := []ReversalRejectionInput{{
		Timestamp: 10,
		Direction: AccumulationDirectionLong,
		POC:       100,
	}}
	auctions := []ReversalFailedAuctionInput{{
		Timestamp:       11,
		Strategy:        FailedLowAuction,
		Level:           100,
		FollowThrough20: 0,
	}}
	setups := DetectReversalSetups(rejections, auctions, ReversalSetupConfig{MaxEventDistanceMillis: 10, ConfluenceThresholdPct: 0.01})
	if len(setups) != 1 {
		t.Fatalf("setups=%+v", setups)
	}
	if setups[0].Accepted || setups[0].Rejected {
		t.Fatalf("expected neutral setup: %+v", setups[0])
	}
}
