package strategy

type TargetRule struct {
	Name        string  `json:"name"`
	SourceBook  string  `json:"sourceBook"`
	TargetType  string  `json:"targetType"`
	Target1     string  `json:"target1"`
	Target2     string  `json:"target2"`
	Target3     string  `json:"target3"`
	MinRR       float64 `json:"minRR"`
	PartialPct1 float64 `json:"partialPct1"`
	PartialPct2 float64 `json:"partialPct2"`
	PartialPct3 float64 `json:"partialPct3"`
}
