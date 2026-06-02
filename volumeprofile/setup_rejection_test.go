package volumeprofile

import "testing"

func TestBullishRejectionCreatesLongContext(t *testing.T) {
	points := []RejectionPricePoint{
		rejectionPoint(1, 105, 106, 94, 95, 20, "bearish", false, ""),
		rejectionPoint(2, 95, 101, 90, 100, 30, "none", true, RejectionDirectionBullish),
		rejectionPoint(3, 100, 111, 99, 110, 25, "bullish", false, ""),
		rejectionPoint(4, 108, 109, 94, 98, 10, "none", false, ""),
		rejectionPoint(10, 98, 116, 97, 115, 10, "none", false, ""),
	}
	setups := DetectRejectionSetups(points, ScopedPOCLevels{}, RejectionSetupConfig{ConfirmationLookahead: 3, RetestLookahead: 5})
	if len(setups) != 1 {
		t.Fatalf("setups=%+v", setups)
	}
	setup := setups[0]
	if setup.Direction != AccumulationDirectionLong || setup.SetupLevel != setup.POC || setup.RejectionLow != 90 {
		t.Fatalf("setup=%+v", setup)
	}
	if !setup.POCRetestDetected {
		t.Fatalf("expected POC retest: %+v", setup)
	}
	if setup.FollowThrough5 <= 0 {
		t.Fatalf("expected positive long follow-through: %+v", setup)
	}
}

func TestBearishRejectionCreatesShortContext(t *testing.T) {
	points := []RejectionPricePoint{
		rejectionPoint(1, 95, 106, 94, 105, 20, "bullish", false, ""),
		rejectionPoint(2, 105, 112, 101, 102, 30, "none", true, RejectionDirectionBearish),
		rejectionPoint(3, 102, 103, 91, 92, 25, "bearish", false, ""),
		rejectionPoint(4, 93, 106, 92, 105, 10, "none", false, ""),
		rejectionPoint(10, 105, 106, 80, 82, 10, "none", false, ""),
	}
	setups := DetectRejectionSetups(points, ScopedPOCLevels{}, RejectionSetupConfig{ConfirmationLookahead: 3, RetestLookahead: 5})
	if len(setups) != 1 {
		t.Fatalf("setups=%+v", setups)
	}
	setup := setups[0]
	if setup.Direction != AccumulationDirectionShort || setup.RejectionHigh != 112 {
		t.Fatalf("setup=%+v", setup)
	}
	if setup.FollowThrough5 <= 0 {
		t.Fatalf("expected positive short follow-through: %+v", setup)
	}
}

func TestRejectionHVNLVNAndInvalidation(t *testing.T) {
	points := []RejectionPricePoint{
		rejectionPoint(1, 100, 110, 90, 92, 10, "bearish", false, ""),
		rejectionPoint(2, 92, 105, 88, 103, 40, "none", true, RejectionDirectionBullish),
		rejectionPoint(3, 103, 120, 102, 118, 45, "bullish", false, ""),
		rejectionPoint(4, 118, 119, 86, 87, 10, "none", false, ""),
	}
	setups := DetectRejectionSetups(points, ScopedPOCLevels{}, RejectionSetupConfig{ConfirmationLookahead: 3, RetestLookahead: 5})
	if len(setups) != 1 {
		t.Fatalf("setups=%+v", setups)
	}
	setup := setups[0]
	if setup.NearestLVN < 0 || setup.NearestHVN < 0 {
		t.Fatalf("expected node fields recorded: %+v", setup)
	}
	if !setup.Invalidated {
		t.Fatalf("expected long invalidation below rejection low/VAL: %+v", setup)
	}
}

func TestRejectionVWAPAndPOCConfluence(t *testing.T) {
	points := []RejectionPricePoint{
		rejectionPoint(1, 105, 106, 94, 95, 20, "bearish", false, ""),
		rejectionPoint(2, 95, 101, 90, 100, 30, "none", true, RejectionDirectionBullish),
		rejectionPoint(3, 100, 111, 99, 110, 25, "bullish", false, ""),
		rejectionPoint(4, 108, 109, 96, 98, 10, "none", false, ""),
	}
	for i := range points {
		points[i].VWAPAlignment = "above_vwap"
		points[i].L2BookPressure = "bid_pressure"
	}
	setups := DetectRejectionSetups(points, ScopedPOCLevels{Daily: 95, Rolling3D: 95, Rolling7D: 95, Composite30D: 95}, RejectionSetupConfig{ConfirmationLookahead: 3, RetestLookahead: 5, ConfluenceThresholdPct: 0.10})
	if len(setups) != 1 {
		t.Fatalf("setups=%+v", setups)
	}
	setup := setups[0]
	if setup.VWAPAlignment != "above_vwap" || setup.L2BookPressure != "bid_pressure" || !setup.DailyPOCConfluence {
		t.Fatalf("setup=%+v", setup)
	}
}

func rejectionPoint(timestamp int64, open float64, high float64, low float64, close float64, volume float64, aggression string, rejected bool, rejectionDirection string) RejectionPricePoint {
	level := 0.0
	if rejectionDirection == RejectionDirectionBullish {
		level = low
	}
	if rejectionDirection == RejectionDirectionBearish {
		level = high
	}
	return RejectionPricePoint{
		Timestamp:           timestamp,
		Open:                open,
		High:                high,
		Low:                 low,
		Close:               close,
		Volume:              volume,
		AggressionDirection: aggression,
		AggressionScore:     1.5,
		RejectionDetected:   rejected,
		RejectionDirection:  rejectionDirection,
		RejectionLevel:      level,
	}
}
