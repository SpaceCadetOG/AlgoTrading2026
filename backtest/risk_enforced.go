package backtest

import (
	"AlgoTrading2026/exchanges"
	"AlgoTrading2026/risk"
	"AlgoTrading2026/series"
)

type RiskEnforcedResult struct {
	Strategy         string               `json:"strategy"`
	Before           Report               `json:"before"`
	After            Report               `json:"after"`
	Violations       []risk.RiskViolation `json:"violations"`
	ViolationsByRule map[string]int       `json:"violationsByRule"`
	ImprovedPnL      bool                 `json:"improvedPnl"`
	ImprovedDrawdown bool                 `json:"improvedDrawdown"`
}

func RunRiskEnforcedWithSignals(
	candles []exchanges.Candle,
	signals series.SignalResult,
	config Config,
	controls risk.RiskControlConfig,
	strategyName string,
) RiskEnforcedResult {
	before := RunWithSignals(candles, signals, config)
	engine := NewEngine(config, nil)
	controlEngine := risk.NewRiskControlEngine(controls)
	violations := runWithRiskControls(engine, candles, signals, controlEngine, strategyName)
	after := engine.reportFromState(len(candles))

	return RiskEnforcedResult{
		Strategy:         strategyName,
		Before:           before,
		After:            after,
		Violations:       violations,
		ViolationsByRule: risk.CountViolationsByRule(violations),
		ImprovedPnL:      after.Metrics.NetPnL > before.Metrics.NetPnL,
		ImprovedDrawdown: after.Metrics.MaxDrawdownPct < before.Metrics.MaxDrawdownPct,
	}
}

func runWithRiskControls(
	engine *Engine,
	candles []exchanges.Candle,
	signals series.SignalResult,
	controlEngine *risk.RiskControlEngine,
	strategyName string,
) []risk.RiskViolation {
	replay := NewReplay(candles)
	signalByTime := signalsByTime(signals)
	positionByTime := signalPositionsByTime(signals)
	violations := make([]risk.RiskViolation, 0)
	openIndex := -1

	for i, candle := range replay.Candles {
		engine.markOpenPosition(candle)
		forcedExit := false

		if engine.OpenPosition != nil {
			holdBars := 0
			if openIndex >= 0 {
				holdBars = i - openIndex
			}
			positionViolations := controlEngine.CheckOpenPosition(risk.PositionCheck{
				Timestamp: candle.StartTime,
				Strategy:  strategyName,
				Symbol:    engine.OpenPosition.Symbol,
				Entry:     engine.OpenPosition.EntryPrice,
				Mark:      engine.OpenPosition.MarkPrice,
				HoldBars:  holdBars,
			})
			violations = append(violations, positionViolations...)
			if risk.MostSevereAction(positionViolations) == risk.ActionForceExit {
				engine.closeLong(candle, "risk force exit")
				openIndex = -1
				forcedExit = true
			}
		}

		signal := signalByTime[candle.StartTime]
		positionChange := positionByTime[candle.StartTime]
		switch {
		case positionChange > 0 && engine.OpenPosition == nil && !forcedExit:
			entryViolations := entryViolationsForCandle(controlEngine, engine, candle, strategyName)
			violations = append(violations, entryViolations...)
			if risk.MostSevereAction(entryViolations) != risk.ActionBlock {
				wasFlat := engine.OpenPosition == nil
				engine.openLong(candle)
				if wasFlat && engine.OpenPosition != nil {
					controlEngine.RecordTrade(candle.StartTime)
					openIndex = i
				}
			}
		case positionChange < 0 && engine.OpenPosition != nil:
			engine.closeLong(candle, "book signal exit")
			openIndex = -1
		}

		engine.markOpenPosition(candle)
		engine.appendPortfolioPoint(candle, signal, positionChange)
	}
	return violations
}

func entryViolationsForCandle(controlEngine *risk.RiskControlEngine, engine *Engine, candle exchanges.Candle, strategyName string) []risk.RiskViolation {
	price := candle.CloseFloat()
	if price <= 0 {
		return []risk.RiskViolation{{
			Timestamp: candle.StartTime,
			Strategy:  strategyName,
			Symbol:    engine.Config.Symbol,
			Rule:      "invalid_price",
			Severity:  risk.ViolationHigh,
			Message:   "invalid candle close price",
			Action:    risk.ActionBlock,
		}}
	}

	size := engine.Config.FixedNotional / price
	return controlEngine.CheckEntry(risk.EntryCheck{
		Timestamp: candle.StartTime,
		Strategy:  strategyName,
		Symbol:    engine.Config.Symbol,
		Size:      size,
		Notional:  engine.Config.FixedNotional,
		Volume:    candle.VolumeFloat(),
	})
}

func (e *Engine) reportFromState(candleCount int) Report {
	metrics := e.metrics()
	return Report{
		Venue:           e.Config.Venue,
		Symbol:          e.Config.Symbol,
		Interval:        e.Config.Interval,
		CandleCount:     candleCount,
		StartingBalance: e.StartingBalance,
		EndingBalance:   e.EquityCurve.EndingEquity(e.CurrentBalance),
		Metrics:         metrics,
		RejectReasons:   append([]string(nil), e.RejectReasons...),
		CSVPath:         e.Config.CSVPath,
		EquityCSVPath:   e.Config.EquityCSVPath,
		SignalCSVPath:   e.Config.SignalCSVPath,
	}
}
