package tradetape

import (
	"os"
	"strconv"
	"strings"

	appconfig "AlgoTrading2026/config"
)

type RecorderConfig struct {
	RunTradeTapeRecorder bool
	OutputPath           string
	IntervalSeconds      int
	MaxRounds            int
}

func DefaultRecorderConfig() RecorderConfig {
	return RecorderConfig{
		RunTradeTapeRecorder: strings.EqualFold(os.Getenv("RUN_TRADE_TAPE_RECORDER"), "true"),
		OutputPath:           envString("TRADE_TAPE_OUTPUT", appconfig.DataPath("trade_tape", "trades.csv")),
		IntervalSeconds:      envInt("TRADE_TAPE_INTERVAL_SECONDS", 5),
		MaxRounds:            envInt("TRADE_TAPE_MAX_ROUNDS", 12),
	}
}

func envString(name string, fallback string) string {
	value := strings.TrimSpace(os.Getenv(name))
	if value == "" {
		return fallback
	}
	return value
}

func envInt(name string, fallback int) int {
	value := strings.TrimSpace(os.Getenv(name))
	if value == "" {
		return fallback
	}
	parsed, err := strconv.Atoi(value)
	if err != nil || parsed <= 0 {
		return fallback
	}
	return parsed
}
