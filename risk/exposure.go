package risk

import (
	"fmt"
	"math"
	"strings"

	"AlgoTrading2026/execution"
	"AlgoTrading2026/strategy"
)

func EstimatedOrderNotional(order execution.OrderRequest) (float64, error) {
	price := strategy.SafeFloat(order.Price)
	if price <= 0 {
		return 0, fmt.Errorf("invalid order price: %q", order.Price)
	}

	size := strategy.SafeFloat(order.Size)
	if size <= 0 {
		return 0, fmt.Errorf("invalid order size: %q", order.Size)
	}

	return price * size, nil
}

func TotalExposure(positions []strategy.PositionState) float64 {
	total := 0.0
	for _, position := range positions {
		total += math.Abs(position.ExposureUSD)
	}

	return total
}

func SymbolExposure(positions []strategy.PositionState, symbol string) float64 {
	total := 0.0
	target := strings.ToUpper(strings.TrimSpace(symbol))
	for _, position := range positions {
		if strings.ToUpper(strings.TrimSpace(position.Symbol)) == target {
			total += math.Abs(position.ExposureUSD)
		}
	}

	return total
}
