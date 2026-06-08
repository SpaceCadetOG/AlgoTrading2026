package strategy

type TradeManagementRule struct {
	Name           string   `json:"name"`
	SourceBook     string   `json:"sourceBook"`
	BreakEvenAfter string   `json:"breakEvenAfter"`
	TrailAfter     string   `json:"trailAfter"`
	EarlyExitIf    []string `json:"earlyExitIf"`
	ForceFlatIf    []string `json:"forceFlatIf"`
}

type PositionManagementRule = TradeManagementRule
