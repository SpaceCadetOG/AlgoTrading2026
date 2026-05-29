package risk

import (
	"fmt"

	"AlgoTrading2026/execution"
	"AlgoTrading2026/strategy"
)

type Engine struct {
	Limits Limits
}

func NewEngine(limits Limits) *Engine {
	return &Engine{Limits: limits}
}

func (e *Engine) CheckOrder(
	account strategy.AccountState,
	positions []strategy.PositionState,
	order execution.OrderRequest,
) Decision {
	if order.ReduceOnly {
		if e.Limits.KillSwitchActive {
			return reject("kill switch active")
		}
		return approve("reduce-only order allowed")
	}

	notional, err := EstimatedOrderNotional(order)
	if err != nil {
		return reject(err.Error())
	}

	if !e.Limits.LiveTradingEnabled {
		return reject("live trading disabled")
	}

	if e.Limits.KillSwitchActive {
		return reject("kill switch active")
	}

	if e.Limits.MaxOpenPositions > 0 && len(positions) >= e.Limits.MaxOpenPositions {
		return reject(fmt.Sprintf("max open positions exceeded: %d >= %d", len(positions), e.Limits.MaxOpenPositions))
	}

	totalExposure := TotalExposure(positions)
	if e.Limits.MaxTotalExposureUSD > 0 && totalExposure+notional > e.Limits.MaxTotalExposureUSD {
		return reject(fmt.Sprintf("max total exposure exceeded: %.2f + %.2f > %.2f", totalExposure, notional, e.Limits.MaxTotalExposureUSD))
	}

	symbolExposure := SymbolExposure(positions, order.Symbol)
	if e.Limits.MaxSymbolExposureUSD > 0 && symbolExposure+notional > e.Limits.MaxSymbolExposureUSD {
		return reject(fmt.Sprintf("max symbol exposure exceeded: %.2f + %.2f > %.2f", symbolExposure, notional, e.Limits.MaxSymbolExposureUSD))
	}

	if account.AvailableEquity-notional < e.Limits.MinAvailableUSD {
		return reject(fmt.Sprintf("min available USD violated: %.2f - %.2f < %.2f", account.AvailableEquity, notional, e.Limits.MinAvailableUSD))
	}

	if e.Limits.MaxLeverage > 0 {
		equity := account.TotalEquity
		if equity <= 0 {
			return reject("cannot estimate leverage with non-positive total equity")
		}

		estimatedLeverage := (totalExposure + notional) / equity
		if estimatedLeverage > e.Limits.MaxLeverage {
			return reject(fmt.Sprintf("max leverage exceeded: %.2f > %.2f", estimatedLeverage, e.Limits.MaxLeverage))
		}
	}

	return approve("approved")
}
