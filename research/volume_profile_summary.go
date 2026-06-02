package research

import (
	"encoding/json"
	"os"

	"AlgoTrading2026/volumeprofile"
)

type VolumeProfileSummary struct {
	Profiles        int      `json:"profiles"`
	Bins            int      `json:"bins"`
	TotalVolume     float64  `json:"totalVolume"`
	POC             float64  `json:"poc"`
	VAH             float64  `json:"vah"`
	VAL             float64  `json:"val"`
	HVNs            int      `json:"hvns"`
	LVNs            int      `json:"lvns"`
	DProfiles       int      `json:"dProfiles"`
	PProfiles       int      `json:"pProfiles"`
	BProfiles       int      `json:"bProfiles"`
	ThinProfiles    int      `json:"thinProfiles"`
	UnknownProfiles int      `json:"unknownProfiles"`
	Notes           []string `json:"notes"`
}

func BuildVolumeProfileSummary(profiles []volumeprofile.VolumeProfile) VolumeProfileSummary {
	summary := VolumeProfileSummary{
		Profiles: len(profiles),
		Notes: []string{
			"Phase 1 uses OHLCV candle approximation.",
			"Candle volume is distributed across price bins between high and low.",
			"This is not tick-accurate volume-at-price yet.",
			"Future improvements can use trade tape volume-at-price, historical L2 context, and tick-level profile reconstruction.",
		},
	}
	for _, profile := range profiles {
		summary.Bins += len(profile.Bins)
		summary.TotalVolume += profile.TotalVolume
		summary.POC = profile.POC
		summary.VAH = profile.VAH
		summary.VAL = profile.VAL
		summary.HVNs += len(profile.HVNs)
		summary.LVNs += len(profile.LVNs)
		switch profile.Shape {
		case volumeprofile.ShapeDProfile:
			summary.DProfiles++
		case volumeprofile.ShapePProfile:
			summary.PProfiles++
		case volumeprofile.ShapeBProfile:
			summary.BProfiles++
		case volumeprofile.ShapeThinProfile:
			summary.ThinProfiles++
		default:
			summary.UnknownProfiles++
		}
	}
	return summary
}

func WriteVolumeProfileSummaryJSON(path string, summary VolumeProfileSummary) error {
	if err := ensureDir(path); err != nil {
		return err
	}
	body, err := json.MarshalIndent(summary, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, append(body, '\n'), 0644)
}
