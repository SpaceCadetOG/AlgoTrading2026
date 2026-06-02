package risk

import (
	"fmt"
	"math"
	"time"
)

type RiskControlConfig struct {
	MaxTradesPerDay           int
	MaxTradeSize              float64
	MaxNotional               float64
	MaxHoldBars               int
	StopLossPct               float64
	MaxVolumeParticipationPct float64
	EnableViolationLogging    bool
}

type RiskControlEngine struct {
	Config      RiskControlConfig
	tradesByDay map[string]int
}

type EntryCheck struct {
	Timestamp int64
	Strategy  string
	Symbol    string
	Size      float64
	Notional  float64
	Volume    float64
}

type PositionCheck struct {
	Timestamp int64
	Strategy  string
	Symbol    string
	Entry     float64
	Mark      float64
	HoldBars  int
}

func DefaultRiskControlConfig() RiskControlConfig {
	return RiskControlConfig{
		MaxTradesPerDay:           20,
		MaxTradeSize:              1,
		MaxNotional:               100,
		MaxHoldBars:               96,
		StopLossPct:               0.02,
		MaxVolumeParticipationPct: 1,
		EnableViolationLogging:    true,
	}
}

func NewRiskControlEngine(config RiskControlConfig) *RiskControlEngine {
	if config.MaxTradesPerDay <= 0 {
		config.MaxTradesPerDay = DefaultRiskControlConfig().MaxTradesPerDay
	}
	if config.MaxTradeSize <= 0 {
		config.MaxTradeSize = DefaultRiskControlConfig().MaxTradeSize
	}
	if config.MaxNotional <= 0 {
		config.MaxNotional = DefaultRiskControlConfig().MaxNotional
	}
	if config.MaxHoldBars <= 0 {
		config.MaxHoldBars = DefaultRiskControlConfig().MaxHoldBars
	}
	if config.StopLossPct <= 0 {
		config.StopLossPct = DefaultRiskControlConfig().StopLossPct
	}
	if config.MaxVolumeParticipationPct <= 0 {
		config.MaxVolumeParticipationPct = DefaultRiskControlConfig().MaxVolumeParticipationPct
	}
	return &RiskControlEngine{Config: config, tradesByDay: make(map[string]int)}
}

func (e *RiskControlEngine) CheckEntry(check EntryCheck) []RiskViolation {
	var violations []RiskViolation
	day := dayKey(check.Timestamp)
	if e.tradesByDay[day] >= e.Config.MaxTradesPerDay {
		violations = append(violations, newViolation(check.Timestamp, check.Strategy, check.Symbol, RuleMaxTradesPerDay, ViolationHigh, ActionBlock, "max trades per day exceeded"))
	}
	if check.Size > e.Config.MaxTradeSize {
		violations = append(violations, newViolation(check.Timestamp, check.Strategy, check.Symbol, RuleMaxTradeSize, ViolationMedium, ActionBlock, fmt.Sprintf("trade size %.6f exceeds max %.6f", check.Size, e.Config.MaxTradeSize)))
	}
	if check.Notional > e.Config.MaxNotional {
		violations = append(violations, newViolation(check.Timestamp, check.Strategy, check.Symbol, RuleMaxNotional, ViolationHigh, ActionBlock, fmt.Sprintf("notional %.2f exceeds max %.2f", check.Notional, e.Config.MaxNotional)))
	}
	if check.Volume > 0 {
		participationPct := check.Size / check.Volume * 100
		if participationPct > e.Config.MaxVolumeParticipationPct {
			violations = append(violations, newViolation(check.Timestamp, check.Strategy, check.Symbol, RuleVolumeParticipation, ViolationMedium, ActionBlock, fmt.Sprintf("volume participation %.4f%% exceeds max %.4f%%", participationPct, e.Config.MaxVolumeParticipationPct)))
		}
	}
	return violations
}

func (e *RiskControlEngine) CheckOpenPosition(check PositionCheck) []RiskViolation {
	var violations []RiskViolation
	if check.Entry > 0 && check.Mark > 0 {
		pnlPct := (check.Mark - check.Entry) / check.Entry
		if pnlPct <= -math.Abs(e.Config.StopLossPct) {
			violations = append(violations, newViolation(check.Timestamp, check.Strategy, check.Symbol, RuleStopLoss, ViolationHigh, ActionForceExit, fmt.Sprintf("stop loss breached %.4f <= -%.4f", pnlPct, math.Abs(e.Config.StopLossPct))))
		}
	}
	if check.HoldBars >= e.Config.MaxHoldBars {
		violations = append(violations, newViolation(check.Timestamp, check.Strategy, check.Symbol, RuleMaxHoldBars, ViolationMedium, ActionForceExit, fmt.Sprintf("hold bars %d reached max %d", check.HoldBars, e.Config.MaxHoldBars)))
	}
	return violations
}

func (e *RiskControlEngine) RecordTrade(timestamp int64) {
	e.tradesByDay[dayKey(timestamp)]++
}

func MostSevereAction(violations []RiskViolation) ViolationAction {
	action := ActionAllow
	for _, violation := range violations {
		switch violation.Action {
		case ActionForceExit:
			return ActionForceExit
		case ActionBlock:
			action = ActionBlock
		case ActionReduce:
			if action == ActionAllow {
				action = ActionReduce
			}
		}
	}
	return action
}

func dayKey(timestamp int64) string {
	return time.UnixMilli(timestamp).UTC().Format("2006-01-02")
}

func newViolation(timestamp int64, strategy string, symbol string, rule string, severity ViolationSeverity, action ViolationAction, message string) RiskViolation {
	return RiskViolation{
		Timestamp: timestamp,
		Strategy:  strategy,
		Symbol:    symbol,
		Rule:      rule,
		Severity:  severity,
		Message:   message,
		Action:    action,
	}
}
