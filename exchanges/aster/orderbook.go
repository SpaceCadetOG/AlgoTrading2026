package aster

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"

	"AlgoTrading2026/orderbook"
)

type AsterDepthSnapshot struct {
	LastUpdateID int64      `json:"lastUpdateId"`
	EventTime    int64      `json:"E"`
	TradeTime    int64      `json:"T"`
	Bids         [][]string `json:"bids"`
	Asks         [][]string `json:"asks"`
}

func (c *Client) GetOrderBook(symbol string) (orderbook.OrderBookSnapshot, error) {
	return GetOrderBook(symbol)
}

func GetOrderBook(symbol string) (orderbook.OrderBookSnapshot, error) {
	return getOrderBookAt(getBaseURL(), symbol)
}

func GetMainnetOrderBook(symbol string) (orderbook.OrderBookSnapshot, error) {
	return getOrderBookAt("https://fapi.asterdex.com", symbol)
}

func getOrderBookAt(baseURL string, symbol string) (orderbook.OrderBookSnapshot, error) {
	if symbol == "" {
		return orderbook.OrderBookSnapshot{}, fmt.Errorf("symbol is required")
	}

	params := url.Values{}
	params.Set("symbol", symbol)
	params.Set("limit", "100")

	resp, err := http.Get(baseURL + "/fapi/v1/depth?" + params.Encode())
	if err != nil {
		return orderbook.OrderBookSnapshot{}, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return orderbook.OrderBookSnapshot{}, err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return orderbook.OrderBookSnapshot{}, fmt.Errorf("aster depth bad status %d: %s", resp.StatusCode, string(body))
	}

	var raw AsterDepthSnapshot
	if err := json.Unmarshal(body, &raw); err != nil {
		return orderbook.OrderBookSnapshot{}, err
	}

	snapshot := orderbook.OrderBookSnapshot{
		Venue:      "aster",
		Symbol:     symbol,
		Time:       asterDepthTime(raw),
		Sequence:   strconv.FormatInt(raw.LastUpdateID, 10),
		Bids:       asterDepthLevels(raw.Bids),
		Asks:       asterDepthLevels(raw.Asks),
		IsSnapshot: true,
	}
	if err := orderbook.ValidateSnapshot(snapshot); err != nil {
		return orderbook.OrderBookSnapshot{}, err
	}
	return snapshot, nil
}

func asterDepthTime(raw AsterDepthSnapshot) int64 {
	if raw.EventTime > 0 {
		return raw.EventTime
	}
	return raw.TradeTime
}

func asterDepthLevels(levels [][]string) []orderbook.BookLevel {
	out := make([]orderbook.BookLevel, 0, len(levels))
	for _, level := range levels {
		if len(level) < 2 {
			continue
		}
		out = append(out, orderbook.BookLevel{Price: level[0], Size: level[1]})
	}
	return out
}
