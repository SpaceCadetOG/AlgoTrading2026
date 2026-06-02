package research

type VolumeSetupTrendSummary struct {
	Setups                 int      `json:"setups"`
	LongContexts           int      `json:"longContexts"`
	ShortContexts          int      `json:"shortContexts"`
	Accepted               int      `json:"accepted"`
	Rejected               int      `json:"rejected"`
	POCRetests             int      `json:"pocRetests"`
	HVNRetests             int      `json:"hvnRetests"`
	VWAPAligned            int      `json:"vwapAligned"`
	BidPressureAligned     int      `json:"bidPressureAligned"`
	AskPressureAligned     int      `json:"askPressureAligned"`
	Invalidated            int      `json:"invalidated"`
	AverageTrendStrength   float64  `json:"averageTrendStrength"`
	AverageFollowThrough5  float64  `json:"averageFollowThrough5"`
	AverageFollowThrough10 float64  `json:"averageFollowThrough10"`
	AverageFollowThrough20 float64  `json:"averageFollowThrough20"`
	Notes                  []string `json:"notes"`
}

func BuildVolumeSetupTrendSummary(rows []VolumeSetupTrendRow) VolumeSetupTrendSummary {
	summary := VolumeSetupTrendSummary{
		Setups: len(rows),
		Notes: []string{
			"Research-only setup study.",
			"No simulated execution.",
			"Volume Profile uses OHLCV candle-volume approximation.",
		},
	}
	var strength, ft5, ft10, ft20 float64
	for _, row := range rows {
		switch row.Direction {
		case "long_context":
			summary.LongContexts++
		case "short_context":
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
		if row.Invalidated {
			summary.Invalidated++
		}
		strength += row.TrendStrengthScore
		ft5 += row.FollowThrough5
		ft10 += row.FollowThrough10
		ft20 += row.FollowThrough20
	}
	if len(rows) > 0 {
		count := float64(len(rows))
		summary.AverageTrendStrength = strength / count
		summary.AverageFollowThrough5 = ft5 / count
		summary.AverageFollowThrough10 = ft10 / count
		summary.AverageFollowThrough20 = ft20 / count
	}
	return summary
}

func vwapAligned(direction string, alignment string) bool {
	return direction == "long_context" && alignment == "above_vwap" ||
		direction == "short_context" && alignment == "below_vwap"
}
