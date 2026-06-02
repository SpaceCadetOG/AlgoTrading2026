package hyperliquid

import (
	"encoding/json"
	"fmt"
	"strconv"

	"AlgoTrading2026/orderbook"
)

type HLL2BookLevel struct {
	Price      string `json:"px"`
	Size       string `json:"sz"`
	OrderCount int    `json:"n"`
}

type HLL2BookResponse struct {
	Coin   string            `json:"coin"`
	Time   int64             `json:"time"`
	Levels [][]HLL2BookLevel `json:"levels"`
}

func (c *Client) GetL2OrderBookSnapshot(symbol string) (*orderbook.OrderBookSnapshot, error) {
	return getL2OrderBookSnapshotAt(getBaseURL(), symbol)
}

func GetMainnetL2OrderBookSnapshot(symbol string) (*orderbook.OrderBookSnapshot, error) {
	return getL2OrderBookSnapshotAt("https://api.hyperliquid.xyz", symbol)
}

func getL2OrderBookSnapshotAt(baseURL string, symbol string) (*orderbook.OrderBookSnapshot, error) {
	if symbol == "" {
		return nil, fmt.Errorf("symbol is required")
	}

	body := fmt.Sprintf(`{"type":"l2Book","coin":"%s"}`, symbol)
	var raw json.RawMessage
	if err := postInfoJSONAt(baseURL, body, &raw); err != nil {
		return nil, err
	}

	book, err := decodeL2Book(symbol, raw)
	if err != nil {
		return nil, err
	}
	if err := orderbook.ValidateSnapshot(*book); err != nil {
		return nil, err
	}
	return book, nil
}

func decodeL2Book(requestedSymbol string, raw json.RawMessage) (*orderbook.OrderBookSnapshot, error) {
	var objectResponse HLL2BookResponse
	if err := json.Unmarshal(raw, &objectResponse); err == nil && len(objectResponse.Levels) >= 2 {
		return normalizeL2Book(requestedSymbol, objectResponse), nil
	}

	var levels [][]HLL2BookLevel
	if err := json.Unmarshal(raw, &levels); err != nil {
		return nil, err
	}
	if len(levels) < 2 {
		return nil, fmt.Errorf("hyperliquid l2Book response missing bid/ask levels")
	}

	return normalizeL2Book(requestedSymbol, HLL2BookResponse{
		Coin:   requestedSymbol,
		Levels: levels,
	}), nil
}

func normalizeL2Book(requestedSymbol string, raw HLL2BookResponse) *orderbook.OrderBookSnapshot {
	symbol := raw.Coin
	if symbol == "" {
		symbol = requestedSymbol
	}

	return &orderbook.OrderBookSnapshot{
		Venue:      "hyperliquid",
		Symbol:     symbol,
		Time:       raw.Time,
		Sequence:   strconv.FormatInt(raw.Time, 10),
		Bids:       normalizeL2Levels(raw.Levels[0]),
		Asks:       normalizeL2Levels(raw.Levels[1]),
		IsSnapshot: true,
	}
}

func normalizeL2Levels(levels []HLL2BookLevel) []orderbook.BookLevel {
	out := make([]orderbook.BookLevel, 0, len(levels))
	for _, level := range levels {
		out = append(out, orderbook.BookLevel{
			Price: level.Price,
			Size:  level.Size,
		})
	}
	return out
}
