package volumeprofile

import (
	"time"

	"AlgoTrading2026/exchanges"
)

const (
	ScopeDailySession = "daily_session"
	ScopeRolling3D    = "rolling_3d"
	ScopeRolling7D    = "rolling_7d"
	ScopeComposite30D = "composite_30d"
)

type ScopedProfile struct {
	Scope   string
	Profile VolumeProfile
}

func BuildScopedProfiles(candles []exchanges.Candle, cfg BinConfig) []ScopedProfile {
	scopes := []struct {
		Name     string
		Lookback time.Duration
	}{
		{Name: ScopeDailySession, Lookback: 24 * time.Hour},
		{Name: ScopeRolling3D, Lookback: 3 * 24 * time.Hour},
		{Name: ScopeRolling7D, Lookback: 7 * 24 * time.Hour},
		{Name: ScopeComposite30D, Lookback: 30 * 24 * time.Hour},
	}
	out := make([]ScopedProfile, 0, len(scopes))
	for _, scope := range scopes {
		window := CandlesInLookback(candles, scope.Lookback)
		if len(window) == 0 {
			out = append(out, ScopedProfile{Scope: scope.Name})
			continue
		}
		out = append(out, ScopedProfile{
			Scope:   scope.Name,
			Profile: BuildProfile(window, cfg),
		})
	}
	return out
}

func CandlesInLookback(candles []exchanges.Candle, lookback time.Duration) []exchanges.Candle {
	if len(candles) == 0 {
		return nil
	}
	if lookback <= 0 {
		return append([]exchanges.Candle(nil), candles...)
	}
	end := candles[len(candles)-1].EndUTC()
	start := end.Add(-lookback)
	first := len(candles)
	for i, candle := range candles {
		if !candle.StartUTC().Before(start) {
			first = i
			break
		}
	}
	if first >= len(candles) {
		first = len(candles) - 1
	}
	return append([]exchanges.Candle(nil), candles[first:]...)
}

func ScopedProfileByName(profiles []ScopedProfile, scope string) (ScopedProfile, bool) {
	for _, profile := range profiles {
		if profile.Scope == scope {
			return profile, true
		}
	}
	return ScopedProfile{}, false
}
