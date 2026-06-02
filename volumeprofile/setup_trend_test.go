package volumeprofile

import "testing"

func TestBullishTrendSetupDetection(t *testing.T) {
	points := []TrendPricePoint{
		trendPoint(1, 100, 101, 99, 100, "none", 0, false, ""),
		trendPoint(2, 101, 112, 100, 111, "bullish", 1.6, true, "up"),
		trendPoint(3, 111, 120, 110, 118, "bullish", 1.4, false, ""),
		trendPoint(4, 118, 121, 117, 120, "bullish", 1.3, false, ""),
		trendPoint(5, 120, 121, 109, 111, "none", 0.5, false, ""),
		trendPoint(6, 111, 116, 110, 115, "none", 0.5, false, ""),
	}
	setups := DetectTrendSetups(points, TrendSetupConfig{RetestLookahead: 5, MaxLegCandles: 3, MinDirectionalCount: 2})
	if len(setups) != 1 {
		t.Fatalf("setups=%+v", setups)
	}
	if setups[0].Direction != AccumulationDirectionLong || setups[0].TrendDirection != TrendDirectionBullish || !setups[0].POCRetestDetected {
		t.Fatalf("setup=%+v", setups[0])
	}
}

func TestBearishTrendSetupDetection(t *testing.T) {
	points := []TrendPricePoint{
		trendPoint(1, 120, 121, 119, 120, "none", 0, false, ""),
		trendPoint(2, 119, 120, 108, 109, "bearish", 1.6, true, "down"),
		trendPoint(3, 109, 110, 100, 102, "bearish", 1.4, false, ""),
		trendPoint(4, 102, 103, 99, 100, "bearish", 1.3, false, ""),
		trendPoint(5, 100, 111, 99, 109, "none", 0.5, false, ""),
	}
	setups := DetectTrendSetups(points, TrendSetupConfig{RetestLookahead: 5, MaxLegCandles: 3, MinDirectionalCount: 2})
	if len(setups) != 1 {
		t.Fatalf("setups=%+v", setups)
	}
	if setups[0].Direction != AccumulationDirectionShort || setups[0].TrendDirection != TrendDirectionBearish {
		t.Fatalf("setup=%+v", setups[0])
	}
}

func TestTrendHVNRetestAndInvalidation(t *testing.T) {
	points := []TrendPricePoint{
		trendPoint(1, 100, 112, 100, 111, "bullish", 1.8, true, "up"),
		trendPoint(2, 111, 122, 111, 121, "bullish", 1.6, false, ""),
		trendPoint(3, 121, 123, 120, 122, "bullish", 1.4, false, ""),
		trendPoint(4, 122, 123, 100, 101, "none", 0.5, false, ""),
	}
	setups := DetectTrendSetups(points, TrendSetupConfig{RetestLookahead: 5, MaxLegCandles: 3, MinDirectionalCount: 2, InvalidationPct: 0.001})
	if len(setups) != 1 {
		t.Fatalf("setups=%+v", setups)
	}
	if setups[0].NearestHVN > 0 && !setups[0].HVNRetestDetected {
		t.Fatalf("expected hvn retest: %+v", setups[0])
	}
	if !setups[0].Invalidated {
		t.Fatalf("expected invalidation: %+v", setups[0])
	}
}

func TestTrendFollowThrough(t *testing.T) {
	points := []TrendPricePoint{
		trendPoint(1, 100, 101, 99, 100, "bullish", 1.4, true, "up"),
		trendPoint(2, 100, 111, 99, 110, "bullish", 1.4, false, ""),
		trendPoint(3, 110, 121, 109, 120, "bullish", 1.4, false, ""),
		trendPoint(4, 120, 121, 119, 120, "none", 0, false, ""),
		trendPoint(5, 120, 126, 119, 125, "none", 0, false, ""),
	}
	if got := trendFollowThrough(points, 3, 1, AccumulationDirectionLong); got != 5 {
		t.Fatalf("ft=%v", got)
	}
	if got := trendFollowThrough(points, 3, 1, AccumulationDirectionShort); got != -5 {
		t.Fatalf("ft=%v", got)
	}
}

func trendPoint(ts int64, open, high, low, close float64, aggression string, score float64, initiation bool, initiationDirection string) TrendPricePoint {
	return TrendPricePoint{
		Timestamp:           ts,
		Open:                open,
		High:                high,
		Low:                 low,
		Close:               close,
		Volume:              10,
		AggressionDirection: aggression,
		AggressionScore:     score,
		InitiationDetected:  initiation,
		InitiationDirection: initiationDirection,
		VWAPAlignment:       "above_vwap",
		L2BookPressure:      "bid_pressure",
	}
}
