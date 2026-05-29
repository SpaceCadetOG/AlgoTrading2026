package lighter

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"AlgoTrading2026/config"
)

type LighterExecutionConfig struct {
	AccountIndex  int
	APIKeyIndex   int
	APIPrivateKey string
}

func LoadExecutionConfig() (*LighterExecutionConfig, error) {
	accountIndex, err := strconv.Atoi(lighterEnv("ACCOUNT_INDEX"))
	if err != nil {
		return nil, fmt.Errorf("invalid Lighter account index: %w", err)
	}

	apiKeyIndex, err := strconv.Atoi(lighterEnv("API_KEY_INDEX"))
	if err != nil {
		return nil, fmt.Errorf("invalid Lighter API key index: %w", err)
	}

	apiPrivateKey := strings.TrimSpace(lighterEnv("API_PRIVATE_KEY"))
	if apiPrivateKey == "" {
		return nil, fmt.Errorf("missing Lighter API private key")
	}

	cleanKey := strings.TrimPrefix(apiPrivateKey, "0x")
	if len(cleanKey) != 80 {
		return nil, fmt.Errorf(
			"wrong Lighter API private key length: got %d chars, expected 80 hex chars or 82 with 0x; use Lighter API signer key, not EVM private key",
			len(cleanKey),
		)
	}

	return &LighterExecutionConfig{
		AccountIndex:  accountIndex,
		APIKeyIndex:   apiKeyIndex,
		APIPrivateKey: apiPrivateKey,
	}, nil
}

func lighterEnv(name string) string {
	if config.IsTestnet() {
		return os.Getenv("LIGHTER_TESTNET_" + name)
	}

	return os.Getenv("LIGHTER_MAINNET_" + name)
}
