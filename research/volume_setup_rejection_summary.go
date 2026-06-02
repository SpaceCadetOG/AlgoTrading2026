package research

type VolumeSetupRejectionSummary struct {
	Setups                    int      `json:"setups"`
	LongContexts              int      `json:"longContexts"`
	ShortContexts             int      `json:"shortContexts"`
	Accepted                  int      `json:"accepted"`
	Rejected                  int      `json:"rejected"`
	POCRetests                int      `json:"pocRetests"`
	HVNRetests                int      `json:"hvnRetests"`
	VWAPAligned               int      `json:"vwapAligned"`
	BidPressureAligned        int      `json:"bidPressureAligned"`
	AskPressureAligned        int      `json:"askPressureAligned"`
	DailyPOCConfluence        int      `json:"dailyPocConfluence"`
	Rolling3DPOCConfluence    int      `json:"rolling3dPocConfluence"`
	Rolling7DPOCConfluence    int      `json:"rolling7dPocConfluence"`
	Composite30DPOCConfluence int      `json:"composite30dPocConfluence"`
	Invalidated               int      `json:"invalidated"`
	AverageFollowThrough5     float64  `json:"averageFollowThrough5"`
	AverageFollowThrough10    float64  `json:"averageFollowThrough10"`
	AverageFollowThrough20    float64  `json:"averageFollowThrough20"`
	Notes                     []string `json:"notes"`
}

func BuildVolumeSetupRejectionSummary(rows []VolumeSetupRejectionRow) VolumeSetupRejectionSummary {
	summary := VolumeSetupRejectionSummary{
		Setups: len(rows),
		Notes: []string{
			"Research-only setup study.",
			"No simulated execution.",
			"Volume Profile uses OHLCV candle-volume approximation.",
			"L2 context is latest snapshot context unless historical L2 replay is explicitly available.",
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
		if row.POCRetestDetected {
			summary.POCRetests++
		}
		if row.HVNRetestDetected {
			summary.HVNRetests++
		}
		if vwapAligned(row.Direction, row.VWAPAlignment) {
			summary.VWAPAligned++
		}
		if row.Direction == "long_context" && row.L2BookPressure == "bid_pressure" {
			summary.BidPressureAligned++
		}
		if row.Direction == "short_context" && row.L2BookPressure == "ask_pressure" {
			summary.AskPressureAligned++
		}
		if row.DailyPOCConfluence {
			summary.DailyPOCConfluence++
		}
		if row.Rolling3DPOCConfluence {
			summary.Rolling3DPOCConfluence++
		}
		if row.Rolling7DPOCConfluence {
			summary.Rolling7DPOCConfluence++
		}
		if row.Composite30DPOCConfluence {
			summary.Composite30DPOCConfluence++
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
