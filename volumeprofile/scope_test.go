package volumeprofile

import (
	"testing"
	"time"

	"AlgoTrading2026/exchanges"
)

func TestCandlesInLookback(t *testing.T) {
	start := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	candles := make([]exchanges.Candle, 0, 5)
	for i := 0; i < 5; i++ {
		candles = append(candles, vpCandle(100, 110, 90, 100, 10, start.Add(time.Duration(i)*24*time.Hour).UnixMilli()))
	}
	window := CandlesInLookback(candles, 2*24*time.Hour)
	if len(window) != 2 {
		t.Fatalf("window len=%d want 2", len(window))
	}
	if window[0].StartTime != candles[3].StartTime {
		t.Fatalf("window starts at %d want %d", window[0].StartTime, candles[3].StartTime)
	}
}

func TestBuildScopedProfiles(t *testing.T) {
	start := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	candles := make([]exchanges.Candle, 0, 8)
	for i := 0; i < 8; i++ {
		candles = append(candles, vpCandle(100+float64(i), 110+float64(i), 90+float64(i), 100+float64(i), 10, start.Add(time.Duration(i)*24*time.Hour).UnixMilli()))
	}
	scoped := BuildScopedProfiles(candles, BinConfig{BinSize: 10})
	if len(scoped) != 4 {
		t.Fatalf("scopes=%d", len(scoped))
	}
	daily, ok := ScopedProfileByName(scoped, ScopeDailySession)
	if !ok || daily.Scope != ScopeDailySession || len(daily.Profile.Bins) == 0 {
		t.Fatalf("missing daily profile: %+v", scoped)
	}
	composite, ok := ScopedProfileByName(scoped, ScopeComposite30D)
	if !ok || composite.Profile.StartTime != candles[0].StartTime {
		t.Fatalf("bad composite profile: %+v", composite)
	}
}
