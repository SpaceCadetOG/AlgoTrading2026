package lighter

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"AlgoTrading2026/exchanges"
)

type LighterActiveOrder struct {
	MarketID         int64           `json:"market_id"`
	OrderIndex       int64           `json:"order_index"`
	ClientOrderIndex int64           `json:"client_order_index"`
	Status           string          `json:"status"`
	Raw              json.RawMessage `json:"-"`
}

type LighterPosition struct {
	MarketID int64           `json:"market_id"`
	Symbol   string          `json:"symbol"`
	Position string          `json:"position"`
	Raw      json.RawMessage `json:"-"`
}

func (c *Client) GetOpenOrders(symbol string) ([]LighterActiveOrder, error) {
	cfg, err := LoadExecutionConfig()
	if err != nil {
		return nil, err
	}

	txClient, _, _, err := newTxClient(cfg)
	if err != nil {
		return nil, err
	}

	auth, err := txClient.GetAuthToken(time.Now().Add(10 * time.Minute))
	if err != nil {
		return nil, fmt.Errorf("lighter auth token: %w", err)
	}

	params := url.Values{}
	params.Set("account_index", strconv.Itoa(cfg.AccountIndex))
	params.Set("market_type", "perp")
	if symbol != "" {
		marketID, _, err := lighterMarketID(symbol)
		if err != nil {
			return nil, err
		}
		params.Set("market_id", strconv.Itoa(int(marketID)))
	}

	req, err := http.NewRequest(http.MethodGet, getBaseURL()+"/api/v1/accountActiveOrders?"+params.Encode(), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("authorization", auth)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("lighter accountActiveOrders bad status %d: %s", resp.StatusCode, string(body))
	}

	return parseActiveOrders(body), nil
}

func (c *Client) GetPositions() ([]exchanges.Position, error) {
	// TODO: Replace this with a first-party Lighter positions endpoint once it is stable.
	// The explorer endpoint shape is not consistent enough for strategy risk.
	return []exchanges.Position{}, nil
}

func (c *Client) GetRawPositions() ([]LighterPosition, error) {
	cfg, err := LoadExecutionConfig()
	if err != nil {
		return nil, err
	}

	resp, err := http.Get("https://explorer.elliot.ai/api/accounts/" + strconv.Itoa(cfg.AccountIndex) + "/positions")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("lighter positions bad status %d: %s", resp.StatusCode, string(body))
	}

	return parsePositions(body), nil
}

func parseActiveOrders(body []byte) []LighterActiveOrder {
	var root any
	if err := json.Unmarshal(body, &root); err != nil {
		return nil
	}

	var out []LighterActiveOrder
	walkObjects(root, func(obj map[string]any) {
		orderIndex := int64FromField(obj, "order_index", "index", "order_id")
		clientOrderIndex := int64FromField(obj, "client_order_index", "client_order_id")
		if orderIndex == 0 && clientOrderIndex == 0 {
			return
		}

		raw, _ := json.Marshal(obj)
		out = append(out, LighterActiveOrder{
			MarketID:         int64FromField(obj, "market_id", "market_index"),
			OrderIndex:       orderIndex,
			ClientOrderIndex: clientOrderIndex,
			Status:           stringFromField(obj, "status", "state"),
			Raw:              raw,
		})
	})

	return out
}

func parsePositions(body []byte) []LighterPosition {
	var root any
	if err := json.Unmarshal(body, &root); err != nil {
		return nil
	}

	var out []LighterPosition
	walkObjects(root, func(obj map[string]any) {
		position := stringFromField(obj, "position", "size")
		if position == "" || position == "0" || position == "0.0" {
			return
		}

		raw, _ := json.Marshal(obj)
		out = append(out, LighterPosition{
			MarketID: int64FromField(obj, "market_id", "market_index"),
			Symbol:   stringFromField(obj, "symbol"),
			Position: position,
			Raw:      raw,
		})
	})

	return out
}

func walkObjects(value any, visit func(map[string]any)) {
	switch typed := value.(type) {
	case map[string]any:
		visit(typed)
		for _, child := range typed {
			walkObjects(child, visit)
		}
	case []any:
		for _, child := range typed {
			walkObjects(child, visit)
		}
	}
}

func int64FromField(obj map[string]any, names ...string) int64 {
	for _, name := range names {
		switch value := obj[name].(type) {
		case float64:
			return int64(value)
		case string:
			parsed, _ := strconv.ParseInt(value, 10, 64)
			return parsed
		case json.Number:
			parsed, _ := value.Int64()
			return parsed
		}
	}

	return 0
}

func stringFromField(obj map[string]any, names ...string) string {
	for _, name := range names {
		switch value := obj[name].(type) {
		case string:
			return value
		case float64:
			return strconv.FormatFloat(value, 'f', -1, 64)
		case json.Number:
			return value.String()
		}
	}

	return ""
}
