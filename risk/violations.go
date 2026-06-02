package risk

type ViolationSeverity string

const (
	ViolationInfo   ViolationSeverity = "INFO"
	ViolationMedium ViolationSeverity = "MEDIUM"
	ViolationHigh   ViolationSeverity = "HIGH"
)

type ViolationAction string

const (
	ActionAllow     ViolationAction = "ALLOW"
	ActionReduce    ViolationAction = "REDUCE"
	ActionBlock     ViolationAction = "BLOCK"
	ActionForceExit ViolationAction = "FORCE_EXIT"
)

const (
	RuleStopLoss            = "stop_loss"
	RuleMaxHoldBars         = "max_hold_bars"
	RuleMaxTradesPerDay     = "max_trades_per_day"
	RuleMaxTradeSize        = "max_trade_size"
	RuleMaxNotional         = "max_notional"
	RuleVolumeParticipation = "volume_participation"
)

type RiskViolation struct {
	Timestamp int64             `json:"timestamp"`
	Strategy  string            `json:"strategy"`
	Symbol    string            `json:"symbol"`
	Rule      string            `json:"rule"`
	Severity  ViolationSeverity `json:"severity"`
	Message   string            `json:"message"`
	Action    ViolationAction   `json:"action"`
}

func CountViolationsByRule(violations []RiskViolation) map[string]int {
	counts := make(map[string]int)
	for _, violation := range violations {
		counts[violation.Rule]++
	}
	return counts
}

func HasBlockingViolation(violations []RiskViolation) bool {
	for _, violation := range violations {
		if violation.Action == ActionBlock || violation.Action == ActionForceExit {
			return true
		}
	}
	return false
}
