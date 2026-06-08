package paper

type EngineState struct {
	Balance           float64          `json:"balance"`
	Equity            float64          `json:"equity"`
	Reserve           float64          `json:"reserve"`
	RealizedPnL       float64          `json:"realizedPnl"`
	OpenPnL           float64          `json:"openPnl"`
	RealizedToday     float64          `json:"realizedToday"`
	OpenCount         int              `json:"openCount"`
	TradeBudgetUsed   int              `json:"tradeBudgetUsed"`
	DailyTradeCount   map[string]int   `json:"dailyTradeCount"`
	SymbolCooldowns   map[string]int64 `json:"symbolCooldowns"`
	SymbolLocks       map[string]int64 `json:"symbolLocks"`
	LossCooldownUntil int64            `json:"lossCooldownUntil"`
	OpenPositions     []PaperPosition  `json:"openPositions"`
	RecentDecisions   []TelemetryEvent `json:"recentDecisions"`
	RecentClosed      []PaperPosition  `json:"recentClosed"`
	RecordersStatus   RecorderStatus   `json:"recordersStatus"`
	LiveEnabled       bool             `json:"liveEnabled"`
	ExecutionEnabled  bool             `json:"executionEnabled"`
	PaperEnabled      bool             `json:"paperEnabled"`
}

func NewState(cfg Config) EngineState {
	return EngineState{
		Balance:          cfg.InitialBalance,
		Equity:           cfg.InitialBalance,
		DailyTradeCount:  map[string]int{},
		SymbolCooldowns:  map[string]int64{},
		SymbolLocks:      map[string]int64{},
		OpenPositions:    []PaperPosition{},
		RecentDecisions:  []TelemetryEvent{},
		RecentClosed:     []PaperPosition{},
		LiveEnabled:      false,
		ExecutionEnabled: false,
		PaperEnabled:     true,
	}
}
