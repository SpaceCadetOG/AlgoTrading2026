package hyperliquid

import "AlgoTrading2026/config"

func getBaseURL() string {
	if config.IsTestnet() {
		return "https://api.hyperliquid-testnet.xyz"
	}

	return "https://api.hyperliquid.xyz"
}

func getWSURL() string {
	if config.IsTestnet() {
		return "wss://api.hyperliquid-testnet.xyz/ws"
	}

	return "wss://api.hyperliquid.xyz/ws"
}