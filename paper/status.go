package paper

type StatusPayload struct {
	Paper struct {
		Mode                string           `json:"mode"`
		Balance             float64          `json:"balance"`
		Equity              float64          `json:"equity"`
		Reserve             float64          `json:"reserve"`
		OpenPnL             float64          `json:"openPnl"`
		RealizedToday       float64          `json:"realizedToday"`
		OpenCount           int              `json:"openCount"`
		RecentDecisionCount int              `json:"recentDecisionCount"`
		RecentClosedCount   int              `json:"recentClosedCount"`
		OpenPositions       []PaperPosition  `json:"openPositions"`
		RecentClosed        []PaperPosition  `json:"recentClosed"`
		RecentDecisions     []TelemetryEvent `json:"recentDecisions"`
		RecordersStatus     RecorderStatus   `json:"recordersStatus"`
	} `json:"paper"`
}

func BuildStatusPayload(state EngineState, cfg Config) StatusPayload {
	var payload StatusPayload
	payload.Paper.Mode = cfg.Mode
	payload.Paper.Balance = state.Balance
	payload.Paper.Equity = state.Equity
	payload.Paper.Reserve = state.Reserve
	payload.Paper.OpenPnL = state.OpenPnL
	payload.Paper.RealizedToday = state.RealizedToday
	payload.Paper.OpenCount = len(state.OpenPositions)
	payload.Paper.RecentDecisionCount = len(state.RecentDecisions)
	payload.Paper.RecentClosedCount = len(state.RecentClosed)
	payload.Paper.OpenPositions = append([]PaperPosition(nil), state.OpenPositions...)
	payload.Paper.RecentClosed = append([]PaperPosition(nil), state.RecentClosed...)
	payload.Paper.RecentDecisions = append([]TelemetryEvent(nil), state.RecentDecisions...)
	payload.Paper.RecordersStatus = state.RecordersStatus
	return payload
}
