package hyperliquid

import (
	"context"
	"crypto/ecdsa"
	"encoding/json"
	"fmt"
	"os"
	"strconv"

	"github.com/ethereum/go-ethereum/crypto"
	hl "github.com/sonirico/go-hyperliquid"

	"AlgoTrading2026/execution"
)

type HLOpenOrder = hl.OpenOrder

type HLCancelResult struct {
	Success bool
	Status  string
	OrderID int64
	Raw     any
}

func (c *Client) PlaceOrder(order execution.OrderRequest) (*execution.OrderResult, error) {
	if os.Getenv("ENABLE_LIVE_ORDERS") != "true" {
		fmt.Println("LIVE ORDERS DISABLED")

		fmt.Printf(
			"[DRY RUN] %s %s %s %s size=%s price=%s\n",
			order.Venue,
			order.Symbol,
			order.Side,
			order.Type,
			order.Size,
			order.Price,
		)

		return &execution.OrderResult{
			Success: true,
			Venue:   "hyperliquid",
			Symbol:  order.Symbol,
			Status:  "DRY_RUN",
			Message: "live orders disabled",
		}, nil
	}

	exchange, err := c.exchange(context.Background())
	if err != nil {
		return nil, err
	}

	size, err := strconv.ParseFloat(order.Size, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid order size: %w", err)
	}

	price, err := strconv.ParseFloat(order.Price, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid order price: %w", err)
	}

	hlOrder := hl.CreateOrderRequest{
		Coin:       order.Symbol,
		IsBuy:      order.Side == execution.Buy,
		Price:      price,
		Size:       size,
		ReduceOnly: order.ReduceOnly,
		OrderType: hl.OrderType{
			Limit: &hl.LimitOrderType{
				Tif: hl.TifGtc,
			},
		},
	}

	c.debugHLAction("order", nil)

	resp, err := exchange.Order(context.Background(), hlOrder, nil)
	c.debugHLResponse("order", resp)
	if err != nil {
		return &execution.OrderResult{
			Success: false,
			Venue:   "hyperliquid",
			Symbol:  order.Symbol,
			Status:  "ERROR",
			Message: err.Error(),
			Raw:     resp,
		}, err
	}

	status, orderID := hlOrderStatus(resp)
	return &execution.OrderResult{
		Success: true,
		Venue:   "hyperliquid",
		Symbol:  order.Symbol,
		OrderID: orderID,
		Status:  status,
		Message: "hyperliquid order accepted",
		Raw:     resp,
	}, nil
}

func (c *Client) CancelOrder(symbol string, orderID int64) (*HLCancelResult, error) {
	exchange, err := c.exchange(context.Background())
	if err != nil {
		return nil, err
	}

	c.debugHLAction("cancel", nil)

	resp, err := exchange.Cancel(context.Background(), symbol, orderID)
	c.debugHLResponse("cancel", resp)
	if err != nil {
		return &HLCancelResult{
			Success: false,
			Status:  "error",
			OrderID: orderID,
			Raw:     resp,
		}, err
	}

	return &HLCancelResult{
		Success: true,
		Status:  "success",
		OrderID: orderID,
		Raw:     resp,
	}, nil
}

func (c *Client) GetOpenOrders() ([]HLOpenOrder, error) {
	info := hl.NewInfo(context.Background(), getBaseURL(), true, nil, nil, nil)
	orders, err := info.OpenOrders(context.Background(), c.Address)
	if err != nil {
		return nil, err
	}

	c.debugHLResponse("openOrders", orders)
	return orders, nil
}

func (c *Client) exchange(ctx context.Context) (*hl.Exchange, error) {
	privateKey, err := c.privateKey()
	if err != nil {
		return nil, err
	}

	return hl.NewExchange(
		ctx,
		privateKey,
		getBaseURL(),
		nil,
		"",
		c.Address,
		nil,
		nil,
		hl.ExchangeOptL1Signer(&debugL1Signer{
			privateKey: privateKey,
			client:     c,
		}),
	), nil
}

func (c *Client) privateKey() (*ecdsa.PrivateKey, error) {
	if c.PrivateKey == "" {
		return nil, fmt.Errorf("missing Hyperliquid private key")
	}

	privateKey, err := crypto.HexToECDSA(c.PrivateKey)
	if err != nil {
		return nil, fmt.Errorf("invalid Hyperliquid private key: %w", err)
	}

	return privateKey, nil
}

func hlOrderStatus(status hl.OrderStatus) (string, string) {
	switch {
	case status.Resting != nil:
		if status.Resting.Status != "" {
			return status.Resting.Status, fmt.Sprintf("%d", status.Resting.Oid)
		}
		return "resting", fmt.Sprintf("%d", status.Resting.Oid)
	case status.Filled != nil:
		return "filled", fmt.Sprintf("%d", status.Filled.Oid)
	case status.Error != nil:
		return "error", ""
	default:
		raw, _ := json.Marshal(status)
		return string(raw), ""
	}
}
