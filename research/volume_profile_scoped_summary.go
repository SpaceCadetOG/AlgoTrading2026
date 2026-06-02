package research

import (
	"encoding/json"
	"os"

	"AlgoTrading2026/volumeprofile"
)

type ScopedVolumeProfileSummaryItem struct {
	Scope    string  `json:"scope"`
	Profiles int     `json:"profiles"`
	Bins     int     `json:"bins"`
	POC      float64 `json:"poc"`
	VAH      float64 `json:"vah"`
	VAL      float64 `json:"val"`
	HVNs     int     `json:"hvns"`
	LVNs     int     `json:"lvns"`
	Shape    string  `json:"shape"`
}

type ScopedVolumeProfileSummary struct {
	Scopes []ScopedVolumeProfileSummaryItem `json:"scopes"`
	Notes  []string                         `json:"notes"`
}

func BuildScopedVolumeProfileSummary(scoped []volumeprofile.ScopedProfile) ScopedVolumeProfileSummary {
	summary := ScopedVolumeProfileSummary{
		Scopes: make([]ScopedVolumeProfileSummaryItem, 0, len(scoped)),
		Notes: []string{
			"Scoped profiles allow chart comparison by timeframe/window.",
			"Phase uses OHLCV candle volume approximation, not tick-accurate volume-at-price.",
			"daily_session, rolling_3d, rolling_7d, and composite_30d are built from the current candle dataset.",
		},
	}
	for _, item := range scoped {
		profile := item.Profile
		profiles := 0
		if len(profile.Bins) > 0 {
			profiles = 1
		}
		summary.Scopes = append(summary.Scopes, ScopedVolumeProfileSummaryItem{
			Scope:    item.Scope,
			Profiles: profiles,
			Bins:     len(profile.Bins),
			POC:      profile.POC,
			VAH:      profile.VAH,
			VAL:      profile.VAL,
			HVNs:     len(profile.HVNs),
			LVNs:     len(profile.LVNs),
			Shape:    profile.Shape,
		})
	}
	return summary
}

func WriteScopedVolumeProfileSummaryJSON(path string, summary ScopedVolumeProfileSummary) error {
	if err := ensureDir(path); err != nil {
		return err
	}
	body, err := json.MarshalIndent(summary, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, append(body, '\n'), 0644)
}
