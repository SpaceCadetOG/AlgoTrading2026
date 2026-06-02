package research

type VolumeSetupAccumulationSummary struct {
	Setups                    int      `json:"setups"`
	LongContexts              int      `json:"longContexts"`
	ShortContexts             int      `json:"shortContexts"`
	Accepted                  int      `json:"accepted"`
	Rejected                  int      `json:"rejected"`
	Retests                   int      `json:"retests"`
	Invalidated               int      `json:"invalidated"`
	DailyPOCConfluence        int      `json:"dailyPocConfluence"`
	Rolling3DPOCConfluence    int      `json:"rolling3dPocConfluence"`
	Rolling7DPOCConfluence    int      `json:"rolling7dPocConfluence"`
	Composite30DPOCConfluence int      `json:"composite30dPocConfluence"`
	AverageFollowThrough5     float64  `json:"averageFollowThrough5"`
	AverageFollowThrough10    float64  `json:"averageFollowThrough10"`
	AverageFollowThrough20    float64  `json:"averageFollowThrough20"`
	Notes                     []string `json:"notes"`
}

func BuildVolumeSetupAccumulationSummary(rows []VolumeSetupAccumulationRow) VolumeSetupAccumulationSummary {
	summary := VolumeSetupAccumulationSummary{
		Setups: len(rows),
		Notes: []string{
			"Research-only setup study.",
			"No simulated execution.",
			"Volume Profile uses OHLCV candle-volume approximation.",
		},
	}
	var ft5, ft10, ft20 float64
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
		if row.RetestDetected {
			summary.Retests++
		}
		if row.Invalidated {
			summary.Invalidated++
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
		ft5 += row.FollowThrough5
		ft10 += row.FollowThrough10
		ft20 += row.FollowThrough20
	}
	if len(rows) > 0 {
		summary.AverageFollowThrough5 = ft5 / float64(len(rows))
		summary.AverageFollowThrough10 = ft10 / float64(len(rows))
		summary.AverageFollowThrough20 = ft20 / float64(len(rows))
	}
	return summary
}
