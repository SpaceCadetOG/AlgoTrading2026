package aster

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"

	"AlgoTrading2026/execution"
)

type AsterOrderResponse struct {
	ClientOrderID string `json:"clientOrderId"`
	CumQty        string `json:"cumQty"`
	CumQuote      string `json:"cumQuote"`
	ExecutedQty   string `json:"executedQty"`
	OrderID       int64  `json:"orderId"`
	AvgPrice      string `json:"avgPrice"`
	OrigQty       string `json:"origQty"`
	Price         string `json:"price"`
	ReduceOnly    bool   `json:"reduceOnly"`
	Side          string `json:"side"`
	PositionSide  string `json:"positionSide"`
	Status        string `json:"status"`
	StopPrice     string `json:"stopPrice"`
	ClosePosition bool   `json:"closePosition"`
	Symbol        string `json:"symbol"`
	TimeInForce   string `json:"timeInForce"`
	Type          string `json:"type"`
	OrigType      string `json:"origType"`
	UpdateTime    int64  `json:"updateTime"`
	WorkingType   string `json:"workingType"`
	PriceProtect  bool   `json:"priceProtect"`
}

type AsterCancelAllOpenOrdersResponse struct {
	Code string `json:"code"`
	Msg  string `json:"msg"`
}

func (c *Client) PlaceOrder(
	order execution.OrderRequest,
) (*execution.OrderResult, error) {

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
			Venue:   "aster",
			Symbol:  order.Symbol,
			Status:  "DRY_RUN",
			Message: "live orders disabled",
		}, nil
	}

	params := url.Values{}

	params.Set("symbol", order.Symbol)
	params.Set("side", string(order.Side))
	params.Set("type", string(order.Type))
	params.Set("quantity", order.Size)

	if order.Type == execution.Limit {
		params.Set("price", order.Price)
		params.Set("timeInForce", "GTC")
	}
	if order.ReduceOnly {
		params.Set("reduceOnly", "true")
	}

	body, err := c.signedRequest(
		http.MethodPost,
		"/fapi/v3/order",
		params,
	)

	if err != nil {
		return nil, err
	}

	var result AsterOrderResponse

	err = json.Unmarshal(body, &result)
	if err != nil {
		return nil, err
	}

	return &execution.OrderResult{
		Success: true,
		Venue:   "aster",
		Symbol:  result.Symbol,
		OrderID: fmt.Sprintf("%d", result.OrderID),
		Status:  result.Status,
		Raw:     result,
	}, nil
}

func (c *Client) QueryOrder(symbol string, orderID string) (*AsterOrderResponse, error) {
	params := url.Values{}
	params.Set("symbol", symbol)
	params.Set("orderId", orderID)

	var result AsterOrderResponse
	err := c.signedJSON(http.MethodGet, "/fapi/v3/order", params, &result)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

func (c *Client) CancelOrder(symbol string, orderID string) (*AsterOrderResponse, error) {
	params := url.Values{}
	params.Set("symbol", symbol)
	params.Set("orderId", orderID)

	var result AsterOrderResponse
	err := c.signedJSON(http.MethodDelete, "/fapi/v3/order", params, &result)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

func (c *Client) CancelAllOpenOrders(symbol string) (*AsterCancelAllOpenOrdersResponse, error) {
	params := url.Values{}
	params.Set("symbol", symbol)

	var result AsterCancelAllOpenOrdersResponse
	err := c.signedJSON(http.MethodDelete, "/fapi/v3/allOpenOrders", params, &result)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

func (c *Client) GetOpenOrders(symbol string) ([]AsterOrderResponse, error) {
	params := url.Values{}
	if symbol != "" {
		params.Set("symbol", symbol)
	}

	var result []AsterOrderResponse
	err := c.signedJSON(http.MethodGet, "/fapi/v3/openOrders", params, &result)
	if err != nil {
		return nil, err
	}

	return result, nil
}
