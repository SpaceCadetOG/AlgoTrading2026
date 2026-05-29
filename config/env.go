package config

import "os"

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