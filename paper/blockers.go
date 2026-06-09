package paper

func BuildStateBlockers(state EngineState, cfg Config, nowMS int64) StateBlockers {
	cooldowns := 0
	for _, until := range state.SymbolCooldowns {
		if until > nowMS {
			cooldowns++
		}
	}
	for _, until := range state.SymbolLocks {
		if until > nowMS {
			cooldowns++
		}
	}
	atDailyMax := 0
	for _, count := range state.DailyTradeCount {
		if cfg.MaxTradesPerSymbolPerDay > 0 && count >= cfg.MaxTradesPerSymbolPerDay {
			atDailyMax++
		}
	}
	openSlots := cfg.MaxOpenPositions - len(state.OpenPositions)
	if openSlots < 0 {
		openSlots = 0
	}
	blockers := StateBlockers{
		CooldownSymbols:     cooldowns,
		SymbolsAtDailyMax:   atDailyMax,
		LossCooldownActive:  state.LossCooldownUntil > nowMS,
		OpenPositionSlots:   openSlots,
		MaxOpenPositions:    cfg.MaxOpenPositions,
		OpenPositionBlocked: cfg.MaxOpenPositions > 0 && len(state.OpenPositions) >= cfg.MaxOpenPositions,
	}
	blockers.Active = blockers.CooldownSymbols > 0 ||
		blockers.SymbolsAtDailyMax > 0 ||
		blockers.LossCooldownActive ||
		blockers.OpenPositionBlocked
	return blockers
}
