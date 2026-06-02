package research

import (
	"encoding/json"
	"os"

	"AlgoTrading2026/volumeprofile"
)

type FlexibleVolumeProfileSummary struct {
	SidewaysAccumulationProfiles int            `json:"sidewaysAccumulationProfiles"`
	OpenDriveProfiles            int            `json:"openDriveProfiles"`
	FailedAuctionProfiles        int            `json:"failedAuctionProfiles"`
	HighLowRetestProfiles        int            `json:"highLowRetestProfiles"`
	AcceptedProfiles             int            `json:"acceptedProfiles"`
	RejectedProfiles             int            `json:"rejectedProfiles"`
	ShapeCounts                  map[string]int `json:"shapeCounts"`
}

func BuildFlexibleVolumeProfileSummary(rows []FlexibleVolumeProfileRow) FlexibleVolumeProfileSummary {
	summary := FlexibleVolumeProfileSummary{
		ShapeCounts: map[string]int{
			volumeprofile.ShapeDProfile:    0,
			volumeprofile.ShapePProfile:    0,
			volumeprofile.ShapeBProfile:    0,
			volumeprofile.ShapeThinProfile: 0,
		},
	}
	for _, row := range rows {
		switch row.ProfileType {
		case volumeprofile.FlexibleSidewaysAccumulation:
			summary.SidewaysAccumulationProfiles++
		case volumeprofile.FlexibleOpenDrive:
			summary.OpenDriveProfiles++
		case volumeprofile.FlexibleFailedAuction:
			summary.FailedAuctionProfiles++
		case volumeprofile.FlexibleHighLowRetest:
			summary.HighLowRetestProfiles++
		}
		switch row.AcceptanceState {
		case volumeprofile.AcceptanceAccepted:
			summary.AcceptedProfiles++
		case volumeprofile.AcceptanceRejected:
			summary.RejectedProfiles++
		}
		if _, ok := summary.ShapeCounts[row.ProfileShape]; ok {
			summary.ShapeCounts[row.ProfileShape]++
		}
	}
	return summary
}

func WriteFlexibleVolumeProfileSummaryJSON(path string, summary FlexibleVolumeProfileSummary) error {
	if err := ensureDir(path); err != nil {
		return err
	}
	body, err := json.MarshalIndent(summary, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, append(body, '\n'), 0644)
}
