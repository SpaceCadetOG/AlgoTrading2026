package runtime

import (
	"fmt"
	"strings"

	"AlgoTrading2026/execution"
)

type RoutedOrderPlacer struct {
	Placers map[string]execution.OrderPlacer
}

func (r RoutedOrderPlacer) PlaceOrder(order execution.OrderRequest) (*execution.OrderResult, error) {
	venue := strings.ToLower(strings.TrimSpace(order.Venue))
	placer := r.Placers[venue]
	if placer == nil {
		return &execution.OrderResult{
			Success: false,
			Venue:   order.Venue,
			Symbol:  order.Symbol,
			Status:  "REJECTED",
			Message: "venue_order_placer_unavailable",
		}, nil
	}
	result, err := placer.PlaceOrder(order)
	if err != nil {
		return nil, fmt.Errorf("%s place order: %w", venue, err)
	}
	if result.Venue == "" {
		result.Venue = venue
	}
	if result.Symbol == "" {
		result.Symbol = order.Symbol
	}
	return result, nil
}
