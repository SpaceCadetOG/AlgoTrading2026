package paper

type ConfigKnob struct {
	Name        string `json:"name"`
	Class       string `json:"class"`
	Replacement string `json:"replacement,omitempty"`
	Note        string `json:"note,omitempty"`
}

func ConfigKnobs() []ConfigKnob {
	return []ConfigKnob{
		{Name: "TRADING_ENV", Class: "operator"},
		{Name: "DATA_ROOT", Class: "operator"},
		{Name: "MAX_SYMBOLS", Class: "operator"},
		{Name: "VENUES", Class: "operator"},
		{Name: "MIN_24H_VOLUME_USD", Class: "operator"},
		{Name: "RESET_PAPER_STATE", Class: "operator"},
		{Name: "PAPER_INITIAL_BALANCE", Class: "operator"},
		{Name: "MAX_OPEN_POSITIONS", Class: "operator"},
		{Name: "RISK_PER_TRADE_PCT", Class: "operator"},
		{Name: "MAX_DAILY_LOSS_PCT", Class: "operator"},

		{Name: "PAPER_MAX_SYMBOLS", Class: "deprecated_alias", Replacement: "MAX_SYMBOLS"},
		{Name: "PAPER_VENUES", Class: "deprecated_alias", Replacement: "VENUES"},
		{Name: "PAPER_MIN_24H_VOLUME_USD", Class: "deprecated_alias", Replacement: "MIN_24H_VOLUME_USD"},
		{Name: "PAPER_RESET_STATE", Class: "deprecated_alias", Replacement: "RESET_PAPER_STATE"},
		{Name: "PAPER_MAX_OPEN_POSITIONS", Class: "deprecated_alias", Replacement: "MAX_OPEN_POSITIONS"},
		{Name: "PAPER_RISK_PER_TRADE_PCT", Class: "deprecated_alias", Replacement: "RISK_PER_TRADE_PCT"},
		{Name: "PAPER_MAX_DAILY_LOSS_PCT", Class: "deprecated_alias", Replacement: "MAX_DAILY_LOSS_PCT"},

		{Name: "TRADING_MODE", Class: "code_default", Note: "operator launch prompt should select mode"},
		{Name: "PAPER_UNIVERSE_MODE", Class: "code_default", Note: "dynamic by default; manual only for diagnostics"},
		{Name: "PAPER_MAX_TRADES_PER_SYMBOL_PER_DAY", Class: "code_default"},
		{Name: "PAPER_REQUIRE_24H_VOLUME", Class: "code_default"},
		{Name: "PAPER_MAX_LEVERAGE", Class: "code_default"},
		{Name: "PAPER_MIN_TP1_RR", Class: "code_default"},
		{Name: "PAPER_ALLOW_PARTIAL_FILLS", Class: "code_default"},
		{Name: "PAPER_ENABLE_FUNDING", Class: "code_default"},
		{Name: "PAPER_ENABLE_RECORDERS", Class: "code_default"},
		{Name: "PAPER_MAKER_FEE_BPS", Class: "code_default"},
		{Name: "PAPER_TAKER_FEE_BPS", Class: "code_default"},
		{Name: "PAPER_FUNDING_HAZARD_RATE", Class: "code_default"},
		{Name: "PAPER_FORCE_FLAT_UTC_HOUR", Class: "code_default"},
		{Name: "PAPER_SCAN_INTERVAL_SECONDS", Class: "code_default"},
		{Name: "PAPER_MAX_POSITIONS_PER_SYMBOL_VENUE", Class: "code_default"},
		{Name: "PAPER_SYMBOLS", Class: "code_default", Note: "manual universe only"},

		{Name: "PAPER_MAX_RUNTIME_CYCLES", Class: "debug_dev"},
		{Name: "PAPER_RESET_COOLDOWNS", Class: "debug_dev"},
		{Name: "PAPER_RESET_DAILY_COUNTS", Class: "debug_dev"},
		{Name: "PAPER_RESET_OPEN_POSITIONS", Class: "debug_dev"},
		{Name: "PAPER_RESET_RECENT_CLOSED", Class: "debug_dev"},
		{Name: "PAPER_RESET_ALL", Class: "debug_dev"},
		{Name: "PAPER_OVERNIGHT_LOG_MODE", Class: "debug_dev", Note: "temporary overnight review"},
		{Name: "LIVE_OVERNIGHT_LOG_MODE", Class: "debug_dev", Note: "temporary overnight review"},
		{Name: "RUN_L2_RECORDER", Class: "debug_dev"},
		{Name: "RUN_TRADE_TAPE_RECORDER", Class: "debug_dev"},
		{Name: "TRADE_TAPE_OUTPUT", Class: "debug_dev"},
		{Name: "TRADE_TAPE_INTERVAL_SECONDS", Class: "debug_dev"},
		{Name: "TRADE_TAPE_MAX_ROUNDS", Class: "debug_dev"},
		{Name: "RUN_ASTER_AGGTRADES_BACKFILL", Class: "debug_dev"},
		{Name: "ASTER_BACKFILL_SYMBOL", Class: "debug_dev"},
		{Name: "ASTER_BACKFILL_OUTPUT", Class: "debug_dev"},
		{Name: "ASTER_BACKFILL_START_MS", Class: "debug_dev"},
		{Name: "ASTER_BACKFILL_END_MS", Class: "debug_dev"},
		{Name: "RUN_FULL_RESEARCH_HARNESS", Class: "debug_dev"},
		{Name: "ASTER_DEBUG_SIGNING", Class: "debug_dev"},
		{Name: "HL_DEBUG_SIGNING", Class: "debug_dev"},
		{Name: "LIGHTER_DEBUG_SIGNING", Class: "debug_dev"},

		{Name: "LIVE_ENABLE_LIVE_TRADING", Class: "live_safety"},
		{Name: "ENABLE_LIVE_ORDERS", Class: "live_safety"},
		{Name: "LIVE_ALLOWED_VENUES", Class: "live_safety"},
		{Name: "LIVE_KILL_SWITCH", Class: "live_safety"},
		{Name: "LIVE_VENUE_HEALTHY", Class: "live_safety"},
		{Name: "LIVE_ACCOUNT_READY", Class: "live_safety"},
		{Name: "LIVE_ENABLE_WEBSOCKET_RECONCILIATION", Class: "live_safety"},
		{Name: "LIVE_REQUIRE_WEBSOCKET_RECONCILIATION", Class: "live_safety"},
		{Name: "ENABLE_LIGHTER_SENDTX", Class: "live_safety"},

		{Name: "WALLET_ADDRESS", Class: "credential"},
		{Name: "HYPERLIQUID_PRIVATE_KEY", Class: "credential"},
		{Name: "ASTER_ACCOUNT_INDEX", Class: "credential"},
		{Name: "ASTER_API_KEY_INDEX", Class: "credential"},
		{Name: "ASTER_API_PRIVATE_KEY", Class: "credential"},
		{Name: "ASTER_TESTNET_ACCOUNT_INDEX", Class: "credential"},
		{Name: "ASTER_TESTNET_API_KEY_INDEX", Class: "credential"},
		{Name: "ASTER_TESTNET_API_PRIVATE_KEY", Class: "credential"},
		{Name: "LIGHTER_MAINNET_ACCOUNT_INDEX", Class: "credential"},
		{Name: "LIGHTER_MAINNET_API_KEY_INDEX", Class: "credential"},
		{Name: "LIGHTER_MAINNET_API_PRIVATE_KEY", Class: "credential"},
		{Name: "LIGHTER_TESTNET_ACCOUNT_INDEX", Class: "credential"},
		{Name: "LIGHTER_TESTNET_API_KEY_INDEX", Class: "credential"},
		{Name: "LIGHTER_TESTNET_API_PRIVATE_KEY", Class: "credential"},

		{Name: "PAPER_ALLOW_SYNTHETIC_CANDIDATE", Class: "delete", Note: "synthetic candidate must not be normal runtime path"},
		{Name: "PAPER_ALLOW_STRATEGY_STACKING", Class: "delete", Note: "dedupe should be deterministic default"},
	}
}

func ConfigKnobsByClass(class string) []ConfigKnob {
	var out []ConfigKnob
	for _, knob := range ConfigKnobs() {
		if knob.Class == class {
			out = append(out, knob)
		}
	}
	return out
}
