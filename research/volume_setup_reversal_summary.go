package research

type VolumeSetupReversalSummary struct {
	Setups                 int      `json:"setups"`
	LongContexts           int      `json:"longContexts"`
	ShortContexts          int      `json:"shortContexts"`
	Accepted               int      `json:"accepted"`
	Rejected               int      `json:"rejected"`
	Neutral                int      `json:"neutral"`
	VWAPAligned            int      `json:"vwapAligned"`
	POCConfluence          int      `json:"pocConfluence"`
	HVNConfluence          int      `json:"hvnConfluence"`
	VAHVALRejections       int      `json:"vahValRejections"`
	Invalidated            int      `json:"invalidated"`
	AverageFollowThrough5  float64  `json:"averageFollowThrough5"`
	AverageFollowThrough10 float64  `json:"averageFollowThrough10"`
	AverageFollowThrough20 float64  `json:"averageFollowThrough20"`
	Notes                  []string `json:"notes"`
}

func BuildVolumeSetupReversalSummary(rows []VolumeSetupReversalRow) VolumeSetupReversalSummary {
	summary := VolumeSetupReversalSummary{
		Setups: len(rows),
		Notes: []string{
			"Research-only reversal study.",
			"No simulated execution.",
			"Volume Profile uses OHLCV approximation.",
		},
	}
	for _, row := range rows {
		if row.Direction == "long_context" {
			summary.LongContexts++
		}
		if row.Direction == "short_context" {
			summary.ShortContexts++
		}
		if row.Accepted {
			summary.Accepted++
		}
		if row.Rejected {
			summary.Rejected++
		}
		if !row.Accepted && !row.Rejected {
			summary.Neutral++
		}
		if vwapAligned(row.Direction, row.VWAPAlignment) {
			summary.VWAPAligned++
		}
		if row.POCConfluence {
			summary.POCConfluence++
		}
		if row.HVNConfluence {
			summary.HVNConfluence++
		}
		if row.VAHVALRejection {
			summary.VAHVALRejections++
		}
		if row.Invalidated {
			summary.Invalidated++
		}
		summary.AverageFollowThrough5 += row.FollowThrough5
		summary.AverageFollowThrough10 += row.FollowThrough10
		summary.AverageFollowThrough20 += row.FollowThrough20
	}
	if len(rows) > 0 {
		count := float64(len(rows))
		summary.AverageFollowThrough5 /= count
		summary.AverageFollowThrough10 /= count
		summary.AverageFollowThrough20 /= count
	}
	return summary
}
