package paper

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"

	appconfig "AlgoTrading2026/config"
)

type Config struct {
	Mode                                 string
	UniverseMode                         string
	InitialBalance                       float64
	MaxOpenPositions                     int
	MaxTradesPerSymbolPerDay             int
	Min24hVolumeUSD                      float64
	RequireVolumeForQualification        bool
	MaxSymbols                           int
	RiskPerTradePct                      float64
	MaxDailyLossPct                      float64
	MaxLeverage                          float64
	MinTP1RR                             float64
	AllowPartialFills                    bool
	EnableFunding                        bool
	EnableRecorders                      bool
	MakerFeeBps                          float64
	TakerFeeBps                          float64
	LiveEnabled                          bool
	StatePath                            string
	PositionsPath                        string
	TradesPath                           string
	EquityPath                           string
	EventsPath                           string
	FundingPath                          string
	UniverseCSVPath                      string
	UniverseJSONPath                     string
	DiscoveredUniversePath               string
	QualifiedUniversePath                string
	SelectedUniversePath                 string
	QualificationDiagnosticsPath         string
	QualifiedNotSelectedPath             string
	UniverseBiasDiagnosticsPath          string
	CandidateCoveragePath                string
	CandidateQualitySummaryPath          string
	CandidateQualityByVenuePath          string
	CandidateQualityBySymbolPath         string
	RankingPath                          string
	RankingDiagnosticsPath               string
	StrategyReviewCandidatesPath         string
	StrategyReviewRejectionsPath         string
	StrategyReviewSelectedPath           string
	StrategyReviewPositionsPath          string
	StrategyReviewCyclesPath             string
	OpportunityCoveragePath              string
	OpportunityCoverageByVenuePath       string
	OvernightStrategySummaryJSONPath     string
	OvernightStrategySummaryMarkdownPath string
	SummaryJSONPath                      string
	RejectSummaryJSONPath                string
	FundingHazardRate                    float64
	EndOfDayForceFlatHourUTC             int
	ScanIntervalSeconds                  int
	MaxRuntimeCycles                     int
	AllowSyntheticCandidate              bool
	AllowStrategyStacking                bool
	MaxPositionsPerSymbolVenue           int
	ResetState                           bool
	ResetCooldowns                       bool
	ResetDailyCounts                     bool
	ResetOpenPositions                   bool
	ResetRecentClosed                    bool
	ResetAll                             bool
	OvernightLogMode                     bool
	Symbols                              []string
	Venues                               []string
}

func DefaultConfig() Config {
	root := appconfig.PaperDir()
	return Config{
		Mode:                                 envString("TRADING_MODE", ""),
		UniverseMode:                         envString("PAPER_UNIVERSE_MODE", "dynamic"),
		InitialBalance:                       envFloat("PAPER_INITIAL_BALANCE", 1000),
		MaxOpenPositions:                     envIntAlias("MAX_OPEN_POSITIONS", "PAPER_MAX_OPEN_POSITIONS", 3),
		MaxTradesPerSymbolPerDay:             envInt("PAPER_MAX_TRADES_PER_SYMBOL_PER_DAY", 2),
		Min24hVolumeUSD:                      envFloatAlias("MIN_24H_VOLUME_USD", "PAPER_MIN_24H_VOLUME_USD", 5000000),
		RequireVolumeForQualification:        envBool("PAPER_REQUIRE_24H_VOLUME", false),
		MaxSymbols:                           envIntAlias("MAX_SYMBOLS", "PAPER_MAX_SYMBOLS", 25),
		RiskPerTradePct:                      envFloatAlias("RISK_PER_TRADE_PCT", "PAPER_RISK_PER_TRADE_PCT", 1),
		MaxDailyLossPct:                      envFloatAlias("MAX_DAILY_LOSS_PCT", "PAPER_MAX_DAILY_LOSS_PCT", 3),
		MaxLeverage:                          envFloat("PAPER_MAX_LEVERAGE", 10),
		MinTP1RR:                             envFloat("PAPER_MIN_TP1_RR", 1.0),
		AllowPartialFills:                    envBool("PAPER_ALLOW_PARTIAL_FILLS", true),
		EnableFunding:                        envBool("PAPER_ENABLE_FUNDING", true),
		EnableRecorders:                      envBool("PAPER_ENABLE_RECORDERS", true),
		MakerFeeBps:                          envFloat("PAPER_MAKER_FEE_BPS", 0.4),
		TakerFeeBps:                          envFloat("PAPER_TAKER_FEE_BPS", 2.8),
		LiveEnabled:                          false,
		StatePath:                            filepath.Join(root, "state.json"),
		PositionsPath:                        filepath.Join(root, "positions.json"),
		TradesPath:                           filepath.Join(root, "trades.csv"),
		EquityPath:                           filepath.Join(root, "equity.csv"),
		EventsPath:                           filepath.Join(root, "events.jsonl"),
		FundingPath:                          filepath.Join(root, "funding.csv"),
		UniverseCSVPath:                      filepath.Join(root, "universe.csv"),
		UniverseJSONPath:                     filepath.Join(root, "universe.json"),
		DiscoveredUniversePath:               filepath.Join(root, "discovered_universe.json"),
		QualifiedUniversePath:                filepath.Join(root, "qualified_universe.json"),
		SelectedUniversePath:                 filepath.Join(root, "selected_universe.json"),
		QualificationDiagnosticsPath:         filepath.Join(root, "qualification_diagnostics.json"),
		QualifiedNotSelectedPath:             filepath.Join(root, "qualified_not_selected.json"),
		UniverseBiasDiagnosticsPath:          filepath.Join(root, "universe_bias_diagnostics.json"),
		CandidateCoveragePath:                filepath.Join(root, "candidate_coverage.json"),
		CandidateQualitySummaryPath:          filepath.Join(root, "candidate_quality_summary.json"),
		CandidateQualityByVenuePath:          filepath.Join(root, "candidate_quality_by_venue.json"),
		CandidateQualityBySymbolPath:         filepath.Join(root, "candidate_quality_by_symbol.json"),
		RankingPath:                          filepath.Join(root, "ranking.json"),
		RankingDiagnosticsPath:               filepath.Join(root, "ranking_diagnostics.json"),
		StrategyReviewCandidatesPath:         filepath.Join(root, "strategy_review_candidates.jsonl"),
		StrategyReviewRejectionsPath:         filepath.Join(root, "strategy_review_rejections.jsonl"),
		StrategyReviewSelectedPath:           filepath.Join(root, "strategy_review_selected.jsonl"),
		StrategyReviewPositionsPath:          filepath.Join(root, "strategy_review_positions.jsonl"),
		StrategyReviewCyclesPath:             filepath.Join(root, "strategy_review_cycles.jsonl"),
		OpportunityCoveragePath:              filepath.Join(root, "opportunity_coverage.json"),
		OpportunityCoverageByVenuePath:       filepath.Join(root, "opportunity_coverage_by_venue.json"),
		OvernightStrategySummaryJSONPath:     filepath.Join(root, "overnight_strategy_summary.json"),
		OvernightStrategySummaryMarkdownPath: filepath.Join(root, "overnight_strategy_summary.md"),
		SummaryJSONPath:                      filepath.Join(root, "summary.json"),
		RejectSummaryJSONPath:                filepath.Join(root, "reject_summary.json"),
		FundingHazardRate:                    envFloat("PAPER_FUNDING_HAZARD_RATE", 0.0008),
		EndOfDayForceFlatHourUTC:             envInt("PAPER_FORCE_FLAT_UTC_HOUR", 23),
		ScanIntervalSeconds:                  envInt("PAPER_SCAN_INTERVAL_SECONDS", 30),
		MaxRuntimeCycles:                     envInt("PAPER_MAX_RUNTIME_CYCLES", 0),
		AllowSyntheticCandidate:              envBool("PAPER_ALLOW_SYNTHETIC_CANDIDATE", false),
		AllowStrategyStacking:                envBool("PAPER_ALLOW_STRATEGY_STACKING", false),
		MaxPositionsPerSymbolVenue:           envInt("PAPER_MAX_POSITIONS_PER_SYMBOL_VENUE", 1),
		ResetState:                           envBoolAlias("RESET_PAPER_STATE", "PAPER_RESET_STATE", false),
		ResetCooldowns:                       envBool("PAPER_RESET_COOLDOWNS", false),
		ResetDailyCounts:                     envBool("PAPER_RESET_DAILY_COUNTS", false),
		ResetOpenPositions:                   envBool("PAPER_RESET_OPEN_POSITIONS", false),
		ResetRecentClosed:                    envBool("PAPER_RESET_RECENT_CLOSED", false),
		ResetAll:                             envBool("PAPER_RESET_ALL", false),
		OvernightLogMode:                     envBool("PAPER_OVERNIGHT_LOG_MODE", false) || envBool("LIVE_OVERNIGHT_LOG_MODE", false),
		Symbols:                              splitCSV(envString("PAPER_SYMBOLS", "BTC,ETH,SOL")),
		Venues:                               splitCSV(envStringAlias("VENUES", "PAPER_VENUES", "aster,hyperliquid,lighter")),
	}
}

func (c Config) IsPaperMode() bool {
	return strings.EqualFold(strings.TrimSpace(c.Mode), "paper")
}

func envString(name string, fallback string) string {
	value := strings.TrimSpace(os.Getenv(name))
	if value == "" {
		return fallback
	}
	return value
}

func envStringAlias(primary string, legacy string, fallback string) string {
	value := strings.TrimSpace(os.Getenv(primary))
	if value != "" {
		return value
	}
	return envString(legacy, fallback)
}

func envBool(name string, fallback bool) bool {
	value := strings.TrimSpace(strings.ToLower(os.Getenv(name)))
	if value == "" {
		return fallback
	}
	switch value {
	case "1", "true", "yes", "y", "on":
		return true
	case "0", "false", "no", "n", "off":
		return false
	default:
		return fallback
	}
}

func envBoolAlias(primary string, legacy string, fallback bool) bool {
	value := strings.TrimSpace(os.Getenv(primary))
	if value != "" {
		return envBool(primary, fallback)
	}
	return envBool(legacy, fallback)
}

func envInt(name string, fallback int) int {
	value := strings.TrimSpace(os.Getenv(name))
	if value == "" {
		return fallback
	}
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}
	return parsed
}

func envIntAlias(primary string, legacy string, fallback int) int {
	value := strings.TrimSpace(os.Getenv(primary))
	if value != "" {
		return envInt(primary, fallback)
	}
	return envInt(legacy, fallback)
}

func envFloat(name string, fallback float64) float64 {
	value := strings.TrimSpace(os.Getenv(name))
	if value == "" {
		return fallback
	}
	parsed, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return fallback
	}
	return parsed
}

func envFloatAlias(primary string, legacy string, fallback float64) float64 {
	value := strings.TrimSpace(os.Getenv(primary))
	if value != "" {
		return envFloat(primary, fallback)
	}
	return envFloat(legacy, fallback)
}

func splitCSV(value string) []string {
	if strings.TrimSpace(value) == "" {
		return nil
	}
	parts := strings.Split(value, ",")
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
			out = append(out, trimmed)
		}
	}
	return out
}
