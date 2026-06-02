package volumeprofile

import (
	"testing"
	"time"

	"AlgoTrading2026/exchanges"
)

func TestCandlesBetween(t *testing.T) {
	candles := flexibleCandles(5)
	out := CandlesBetween(candles, candles[1].StartTime, candles[3].StartTime)
	if len(out) != 3 || out[0].StartTime != candles[1].StartTime || out[2].StartTime != candles[3].StartTime {
		t.Fatalf("bad slice: %+v", out)
	}
}

func TestBuildFlexibleProfile(t *testing.T) {
	candles := flexibleCandles(10)
	window := FlexibleProfileWindow{
		ProfileType: FlexibleSidewaysAccumulation,
		ProfileID:   "test_1",
		StartTime:   candles[1].StartTime,
		EndTime:     candles[4].StartTime,
	}
	profile := BuildFlexibleProfile(candles, window, BinConfig{BinSize: 1})
	if profile.ProfileType != FlexibleSidewaysAccumulation || profile.ProfileID != "test_1" || len(profile.Profile.Bins) == 0 {
		t.Fatalf("profile=%+v", profile)
	}
	if profile.AcceptanceState == "" {
		t.Fatalf("missing acceptance: %+v", profile)
	}
}

func TestClassifyAcceptance(t *testing.T) {
	candles := flexibleCandles(12)
	profile := VolumeProfile{VAL: 99, VAH: 103}
	if got := ClassifyAcceptance(candles, profile, candles[3].StartTime, 5); got != AcceptanceAccepted {
		t.Fatalf("acceptance=%s", got)
	}
	rejectCandles := flexibleCandles(12)
	rejectCandles[4].Close = "120"
	if got := ClassifyAcceptance(rejectCandles, profile, rejectCandles[3].StartTime, 5); got != AcceptanceRejected {
		t.Fatalf("acceptance=%s", got)
	}
}

func flexibleCandles(count int) []exchanges.Candle {
	start := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	candles := make([]exchanges.Candle, 0, count)
	for i := 0; i < count; i++ {
		ts := start.Add(time.Duration(i) * 15 * time.Minute).UnixMilli()
		candles = append(candles, vpCandle(100, 102, 98, 101, 10, ts))
	}
	return candles
}
