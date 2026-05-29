package lighter

import "AlgoTrading2026/config"

func getBaseURL() string {
	if config.IsTestnet() {
		return "https://testnet.zklighter.elliot.ai"
	}

	return "https://mainnet.zklighter.elliot.ai"
}

func getWSURL() string {
	if config.IsTestnet() {
		return "wss://testnet.zklighter.elliot.ai/stream"
	}

	return "wss://mainnet.zklighter.elliot.ai/stream?readonly=true"
}