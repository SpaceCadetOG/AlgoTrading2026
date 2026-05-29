package backtest

import (
	"fmt"
	"time"

	"AlgoTrading2026/exchanges"
	"AlgoTrading2026/execution"
	"AlgoTrading2026/strategy"
)

type Fill struct {
	Venue     string
	Symbol    string
	Side      string
	Size      float64
	Price     float64
	Notional  float64
	Fee       float64
	Timestamp time.Time
}

func SimulateFill(
	order execution.OrderRequest,
	candle exchanges.Candle,
	liquidity string,
	feeModel FeeModel,
	slippageModel SlippageModel,
) (Fill, error) {
	basePrice := candle.CloseFloat()
	if basePrice <= 0 {
		return Fill{}, fmt.Errorf("invalid candle close price: %s", candle.Close)
	}

	size := strategy.SafeFloat(order.Size)
	if size <= 0 {
		return Fill{}, fmt.Errorf("invalid order size: %s", order.Size)
	}

	notionalBeforeSlippage := basePrice * size
	price := basePrice
	if slippageModel != nil {
		price = slippageModel.Apply(order.Venue, string(order.Side), basePrice, notionalBeforeSlippage, liquidity)
	}

	notional := price * size
	fee := 0.0
	if feeModel != nil {
		fee = feeModel.CalculateFee(order.Venue, notional, liquidity)
	}

	return Fill{
		Venue:     order.Venue,
		Symbol:    order.Symbol,
		Side:      string(order.Side),
		Size:      size,
		Price:     price,
		Notional:  notional,
		Fee:       fee,
		Timestamp: candle.EndUTC(),
	}, nil
}
