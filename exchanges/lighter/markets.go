package lighter

import (
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

type LighterMarket struct {
	ID            int16
	Symbol        string
	PriceDecimals int
	SizeDecimals  int
	Raw           json.RawMessage
}

func (c *Client) ResolveMarket(symbol string) (*LighterMarket, error) {
	marketID, normalized, err := lighterMarketID(symbol)
	if err != nil {
		return nil, err
	}

	market := &LighterMarket{
		ID:            marketID,
		Symbol:        normalized,
		PriceDecimals: 2,
		SizeDecimals:  5,
	}

	details, err := c.GetOrderBookDetails(marketID)
	if err != nil {
		return market, nil
	}
	if details.PriceDecimals > 0 {
		market.PriceDecimals = details.PriceDecimals
	}
	if details.SizeDecimals > 0 {
		market.SizeDecimals = details.SizeDecimals
	}
	market.Raw = details.Raw

	return market, nil
}

func (m *LighterMarket) WirePriceSize(price string, size string) (uint32, int64, error) {
	priceFloat, err := strconv.ParseFloat(price, 64)
	if err != nil {
		return 0, 0, fmt.Errorf("invalid price: %w", err)
	}

	sizeFloat, err := strconv.ParseFloat(size, 64)
	if err != nil {
		return 0, 0, fmt.Errorf("invalid size: %w", err)
	}

	priceInt := uint32(math.Round(priceFloat * pow10Int(m.PriceDecimals)))
	baseAmount := int64(math.Round(sizeFloat * pow10Int(m.SizeDecimals)))
	if priceInt == 0 {
		return 0, 0, fmt.Errorf("price %s rounds to zero using %d decimals", price, m.PriceDecimals)
	}
	if baseAmount == 0 {
		return 0, 0, fmt.Errorf("size %s rounds to zero using %d decimals", size, m.SizeDecimals)
	}

	return priceInt, baseAmount, nil
}

func (c *Client) GetOrderBookDetails(marketID int16) (*LighterMarket, error) {
	params := url.Values{}
	params.Set("market_id", strconv.Itoa(int(marketID)))
	params.Set("filter", "all")

	resp, err := http.Get(getBaseURL() + "/api/v1/orderBooks?" + params.Encode())
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("lighter orderBooks bad status %d: %s", resp.StatusCode, string(body))
	}

	var root any
	if err := json.Unmarshal(body, &root); err != nil {
		return nil, err
	}

	marketNode := findMarketNode(root, float64(marketID))
	if marketNode == nil {
		marketNode = root
	}
	raw, _ := json.Marshal(marketNode)

	return &LighterMarket{
		ID:            marketID,
		PriceDecimals: intFromAny(findField(marketNode, "supported_price_decimals", "price_decimals", "price_decimals_count")),
		SizeDecimals:  intFromAny(findField(marketNode, "supported_size_decimals", "size_decimals", "base_decimals")),
		Raw:           raw,
	}, nil
}

func lighterMarketID(symbol string) (int16, string, error) {
	switch strings.ToUpper(strings.TrimSpace(symbol)) {
	case "ETH", "ETHUSDT":
		return 0, "ETH", nil
	case "BTC", "BTCUSDT":
		return 1, "BTC", nil
	default:
		return 0, "", fmt.Errorf("unsupported lighter symbol: %s", symbol)
	}
}

func findMarketNode(value any, marketID float64) any {
	switch typed := value.(type) {
	case map[string]any:
		if id := findField(typed, "market_id", "market_index", "id", "index"); id != nil {
			if floatFromAny(id) == marketID {
				return typed
			}
		}
		for _, child := range typed {
			if found := findMarketNode(child, marketID); found != nil {
				return found
			}
		}
	case []any:
		for _, child := range typed {
			if found := findMarketNode(child, marketID); found != nil {
				return found
			}
		}
	}

	return nil
}

func findField(value any, names ...string) any {
	typed, ok := value.(map[string]any)
	if !ok {
		return nil
	}

	for _, name := range names {
		if value, ok := typed[name]; ok {
			return value
		}
	}

	return nil
}

func intFromAny(value any) int {
	switch typed := value.(type) {
	case float64:
		return int(typed)
	case string:
		parsed, _ := strconv.Atoi(typed)
		return parsed
	case json.Number:
		parsed, _ := typed.Int64()
		return int(parsed)
	default:
		return 0
	}
}

func floatFromAny(value any) float64 {
	switch typed := value.(type) {
	case float64:
		return typed
	case string:
		parsed, _ := strconv.ParseFloat(typed, 64)
		return parsed
	case json.Number:
		parsed, _ := typed.Float64()
		return parsed
	default:
		return math.NaN()
	}
}
