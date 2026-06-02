package realism

type FillAssumptionModel struct {
	FillRatio                float64             `json:"fillRatio"`
	PartialFillSupported     bool                `json:"partialFillSupported"`
	CancelAmendSupported     bool                `json:"cancelAmendSupported"`
	StaleBookRisk            bool                `json:"staleBookRisk"`
	CrossedBookArtificiality bool                `json:"crossedBookArtificiality"`
	DeterministicFillWarning bool                `json:"deterministicFillWarning"`
	Risk                     DislocationSeverity `json:"risk"`
}

func NewFillAssumptionModel(fillRatio float64, partialFillSupported bool, cancelAmendSupported bool, staleBookRisk bool, crossedBookArtificiality bool, deterministicFillWarning bool) FillAssumptionModel {
	model := FillAssumptionModel{
		FillRatio:                clamp01(fillRatio),
		PartialFillSupported:     partialFillSupported,
		CancelAmendSupported:     cancelAmendSupported,
		StaleBookRisk:            staleBookRisk,
		CrossedBookArtificiality: crossedBookArtificiality,
		DeterministicFillWarning: deterministicFillWarning,
	}
	model.Risk = ClassifyFillAssumptionRisk(model)
	return model
}

func DefaultFillAssumptionModel() FillAssumptionModel {
	return NewFillAssumptionModel(1, false, true, true, true, true)
}

func ClassifyFillAssumptionRisk(model FillAssumptionModel) DislocationSeverity {
	if model.CrossedBookArtificiality || model.DeterministicFillWarning || model.StaleBookRisk {
		return SeverityHigh
	}
	if !model.PartialFillSupported || model.FillRatio >= 0.95 {
		return SeverityMedium
	}
	return SeverityLow
}

func clamp01(value float64) float64 {
	switch {
	case value < 0:
		return 0
	case value > 1:
		return 1
	default:
		return value
	}
}

type AggregateRealismModel struct {
	Latency        LatencyModel        `json:"latency"`
	PlaceInLine    PlaceInLineEstimate `json:"placeInLine"`
	MarketImpact   MarketImpactModel   `json:"marketImpact"`
	FillAssumption FillAssumptionModel `json:"fillAssumption"`
	OverallRisk    DislocationSeverity `json:"overallRisk"`
	Warnings       []string            `json:"warnings"`
}

func NewAggregateRealismModel(latency LatencyModel, placeInLine PlaceInLineEstimate, marketImpact MarketImpactModel, fillAssumption FillAssumptionModel) AggregateRealismModel {
	model := AggregateRealismModel{
		Latency:        latency,
		PlaceInLine:    placeInLine,
		MarketImpact:   marketImpact,
		FillAssumption: fillAssumption,
	}
	model.Warnings = aggregateWarnings(model)
	model.OverallRisk = ClassifyAggregateRealismRisk(model)
	return model
}

func DefaultAggregateRealismModel() AggregateRealismModel {
	return NewAggregateRealismModel(
		DefaultLatencyModel(),
		DefaultPlaceInLineEstimate(),
		DefaultMarketImpactModel(),
		DefaultFillAssumptionModel(),
	)
}

func ClassifyAggregateRealismRisk(model AggregateRealismModel) DislocationSeverity {
	for _, risk := range []DislocationSeverity{
		model.Latency.Risk,
		model.PlaceInLine.Risk,
		model.MarketImpact.Risk,
		model.FillAssumption.Risk,
	} {
		if risk == SeverityHigh {
			return SeverityHigh
		}
	}
	for _, risk := range []DislocationSeverity{
		model.Latency.Risk,
		model.PlaceInLine.Risk,
		model.MarketImpact.Risk,
		model.FillAssumption.Risk,
	} {
		if risk == SeverityMedium {
			return SeverityMedium
		}
	}
	return SeverityLow
}

func aggregateWarnings(model AggregateRealismModel) []string {
	var warnings []string
	if model.Latency.Risk == SeverityHigh {
		warnings = append(warnings, "latency assumptions are high risk")
	}
	if model.PlaceInLine.Risk == SeverityHigh {
		warnings = append(warnings, "place-in-line assumptions are high risk")
	}
	if model.MarketImpact.Risk == SeverityHigh {
		warnings = append(warnings, "market impact assumptions are high risk")
	}
	if model.FillAssumption.CrossedBookArtificiality {
		warnings = append(warnings, "crossed-book artificiality remains enabled in the validation harness")
	}
	if model.FillAssumption.DeterministicFillWarning {
		warnings = append(warnings, "deterministic full-fill assumption remains high risk")
	}
	if model.FillAssumption.StaleBookRisk {
		warnings = append(warnings, "stale book risk remains high")
	}
	return warnings
}
