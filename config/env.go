package config

import "os"

func TradingMode() string {
	mode := os.Getenv("TRADING_MODE")
	if mode == "" {
		return ""
	}
	return mode
}

func TradingEnv() string {
	env := os.Getenv("TRADING_ENV")
	if env == "" {
		return "mainnet"
	}

	return env
}

func IsTestnet() bool {
	return TradingEnv() == "testnet"
}
