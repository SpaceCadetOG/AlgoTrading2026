package strategy

type ExecutablePlaybook struct {
	Name                  string              `json:"name"`
	DirectionBias         string              `json:"directionBias"`
	Setup                 string              `json:"setup"`
	Entry                 EntryRule           `json:"entry"`
	Stop                  StopRule            `json:"stop"`
	Target                TargetRule          `json:"target"`
	Management            TradeManagementRule `json:"management"`
	Risk                  RiskRule            `json:"risk"`
	RequiredContext       []string            `json:"requiredContext"`
	RequiredConfirmations []string            `json:"requiredConfirmations"`
	RejectIf              []string            `json:"rejectIf"`
}

type TradeSetup = ExecutablePlaybook

type BookTradeRulesPacket struct {
	Playbooks           []ExecutablePlaybook `json:"playbooks"`
	ExecutionEnabled    bool                 `json:"executionEnabled"`
	PaperTradingEnabled bool                 `json:"paperTradingEnabled"`
	Status              string               `json:"status"`
}

type PlaybookPacket = BookTradeRulesPacket
