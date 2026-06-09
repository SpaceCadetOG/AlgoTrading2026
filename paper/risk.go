package paper

import (
	"math"
	"strings"

	runtime "AlgoTrading2026/internal/runtime"
)

type Candidate = runtime.Candidate

type RiskDecision struct {
	Allowed bool
	Reasons []string
}

func CheckRisk(state EngineState, candidate Candidate, cfg Config, nowMS int64) RiskDecision {
	reasons := []string{}
	key := candidateStateKey(candidate)
	if len(state.OpenPositions) >= cfg.MaxOpenPositions {
		reasons = append(reasons, "max_open_positions")
	}
	if state.LossCooldownUntil > nowMS {
		reasons = append(reasons, "loss_cooldown")
	}
	if until := state.SymbolCooldowns[key]; until > nowMS {
		reasons = append(reasons, "symbol_cooldown")
	}
	if until := state.SymbolLocks[key]; until > nowMS {
		reasons = append(reasons, "symbol_lock")
	}
	if state.DailyTradeCount[key] >= cfg.MaxOpenPositions {
		reasons = append(reasons, "trade_budget_exceeded")
	}
	if state.DailyTradeCount[key] >= cfg.MaxTradesPerSymbolPerDay {
		reasons = append(reasons, "max_trades_per_symbol_per_day")
	}
	if !validBracket(candidate) {
		reasons = append(reasons, "invalid_bracket_geometry")
	}
	if rewardToRisk(candidate) < cfg.MinTP1RR || rewardToRisk(candidate) < candidate.RequiredRR {
		reasons = append(reasons, "poor_rr")
	}
	if FundingHazardBlocked(candidate.FundingRate, candidate.Side, cfg) {
		reasons = append(reasons, "funding_hazard")
	}
	if candidate.SpreadPct > 0.5 {
		reasons = append(reasons, "spread_too_wide")
	}
	if candidate.Liquidity <= 0 {
		reasons = append(reasons, "insufficient_liquidity")
	}
	return RiskDecision{Allowed: len(reasons) == 0, Reasons: reasons}
}

func candidateStateKey(candidate Candidate) string {
	return venueSymbolKey(candidate.Venue, candidate.Symbol)
}

func positionStateKey(position PaperPosition) string {
	return venueSymbolKey(position.Venue, position.Symbol)
}

func venueSymbolKey(venue string, symbol string) string {
	venue = strings.ToLower(strings.TrimSpace(venue))
	symbol = strings.ToUpper(strings.TrimSpace(symbol))
	if venue == "" {
		return symbol
	}
	return venue + ":" + symbol
}

func rewardToRisk(candidate Candidate) float64 {
	risk := math.Abs(candidate.EntryPrice - candidate.StopPrice)
	reward := math.Abs(candidate.TP1 - candidate.EntryPrice)
	if risk == 0 {
		return 0
	}
	return reward / risk
}

func validBracket(candidate Candidate) bool {
	switch normalizeSide(candidate.Side) {
	case "LONG":
		return candidate.StopPrice < candidate.EntryPrice && candidate.TP1 > candidate.EntryPrice
	case "SHORT":
		return candidate.StopPrice > candidate.EntryPrice && candidate.TP1 < candidate.EntryPrice
	default:
		return false
	}
}
