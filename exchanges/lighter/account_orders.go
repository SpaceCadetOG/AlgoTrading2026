package lighter

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"AlgoTrading2026/exchanges"
)

type LighterActiveOrder struct {
	MarketID            int64           `json:"market_id"`
	OrderIndex          int64           `json:"order_index"`
	ClientOrderIndex    int64           `json:"client_order_index"`
	Status              string          `json:"status"`
	Side                string          `json:"side"`
	Type                string          `json:"type"`
	InitialBaseAmount   string          `json:"initial_base_amount"`
	RemainingBaseAmount string          `json:"remaining_base_amount"`
	FilledBaseAmount    string          `json:"filled_base_amount"`
	Price               string          `json:"price"`
	ReduceOnly          bool            `json:"reduce_only"`
	Raw                 json.RawMessage `json:"-"`
}

type LighterPosition struct {
	MarketID int64           `json:"market_id"`
	Symbol   string          `json:"symbol"`
	Position string          `json:"position"`
	Side     string          `json:"side"`
	Entry    string          `json:"entry_price"`
	PnL      string          `json:"pnl"`
	Raw      json.RawMessage `json:"-"`
}

type LighterTrade struct {
	TradeID     string          `json:"trade_id"`
	MarketID    int64           `json:"market_id"`
	AskID       string          `json:"ask_id"`
	BidID       string          `json:"bid_id"`
	AskClientID string          `json:"ask_client_id"`
	BidClientID string          `json:"bid_client_id"`
	Type        string          `json:"type"`
	Side        string          `json:"side"`
	Size        string          `json:"size"`
	Price       string          `json:"price"`
	Timestamp   int64           `json:"timestamp"`
	Raw         json.RawMessage `json:"-"`
}

type LighterAccountWSState struct {
	Orders    []LighterActiveOrder `json:"orders"`
	Trades    []LighterTrade       `json:"trades"`
	Positions []LighterPosition    `json:"positions"`
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

func (c *Client) GetInactiveOrders(symbol string, limit int) ([]LighterActiveOrder, error) {
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
	if limit <= 0 {
		limit = 100
	}
	params := url.Values{}
	params.Set("account_index", strconv.Itoa(cfg.AccountIndex))
	params.Set("market_type", "perp")
	params.Set("limit", strconv.Itoa(limit))
	if symbol != "" {
		marketID, _, err := lighterMarketID(symbol)
		if err != nil {
			return nil, err
		}
		params.Set("market_id", strconv.Itoa(int(marketID)))
	}
	req, err := http.NewRequest(http.MethodGet, getBaseURL()+"/api/v1/accountInactiveOrders?"+params.Encode(), nil)
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
		return nil, fmt.Errorf("lighter accountInactiveOrders bad status %d: %s", resp.StatusCode, string(body))
	}
	return parseActiveOrders(body), nil
}

func (c *Client) GetPositions() ([]exchanges.Position, error) {
	raw, err := c.GetRawPositions()
	if err != nil {
		return nil, err
	}
	out := make([]exchanges.Position, 0, len(raw))
	for _, p := range raw {
		side := strings.ToUpper(strings.TrimSpace(p.Side))
		if side == "" {
			side = "BOTH"
		}
		symbol := strings.ToUpper(strings.TrimSpace(p.Symbol))
		if symbol == "" && p.MarketID >= 0 {
			symbol = lighterSymbolFromMarketID(p.MarketID)
		}
		out = append(out, exchanges.Position{
			Venue:  "lighter",
			Symbol: symbol,
			Side:   side,
			Size:   p.Position,
			Entry:  p.Entry,
			PnL:    p.PnL,
		})
	}
	return out, nil
}

func (c *Client) GetRawPositions() ([]LighterPosition, error) {
	cfg, err := LoadExecutionConfig()
	if err != nil {
		return nil, err
	}

	param := c.positionsLookupParam(cfg)
	resp, err := http.Get("https://explorer.elliot.ai/api/accounts/" + url.PathEscape(param) + "/positions")
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

func (c *Client) positionsLookupParam(cfg *LighterExecutionConfig) string {
	if cfg.AccountIndex > 0 {
		return strconv.Itoa(cfg.AccountIndex)
	}
	if strings.TrimSpace(c.Address) != "" {
		return strings.TrimSpace(c.Address)
	}
	return strconv.Itoa(cfg.AccountIndex)
}

func (c *Client) GetAccountTrades(symbol string, limit int) ([]LighterTrade, error) {
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
	if limit <= 0 {
		limit = 100
	}
	params := url.Values{}
	params.Set("account_index", strconv.Itoa(cfg.AccountIndex))
	params.Set("market_type", "perp")
	params.Set("type", "trade")
	params.Set("limit", strconv.Itoa(limit))
	params.Set("sort_by", "timestamp")
	params.Set("order", "desc")
	if symbol != "" {
		marketID, _, err := lighterMarketID(symbol)
		if err != nil {
			return nil, err
		}
		params.Set("market_id", strconv.Itoa(int(marketID)))
	}
	req, err := http.NewRequest(http.MethodGet, getBaseURL()+"/api/v1/trades?"+params.Encode(), nil)
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
		return nil, fmt.Errorf("lighter trades bad status %d: %s", resp.StatusCode, string(body))
	}
	return parseTrades(body), nil
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
			MarketID:            int64FromField(obj, "market_id", "market_index"),
			OrderIndex:          orderIndex,
			ClientOrderIndex:    clientOrderIndex,
			Status:              stringFromField(obj, "status", "state"),
			Side:                stringFromField(obj, "side"),
			Type:                stringFromField(obj, "type"),
			InitialBaseAmount:   stringFromField(obj, "initial_base_amount"),
			RemainingBaseAmount: stringFromField(obj, "remaining_base_amount"),
			FilledBaseAmount:    stringFromField(obj, "filled_base_amount"),
			Price:               stringFromField(obj, "price"),
			ReduceOnly:          boolFromField(obj, "reduce_only"),
			Raw:                 raw,
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
		position := stringFromField(obj, "position", "position_size")
		if position == "" || position == "0" || position == "0.0" {
			return
		}

		raw, _ := json.Marshal(obj)
		out = append(out, LighterPosition{
			MarketID: int64FromField(obj, "market_id", "market_index"),
			Symbol:   stringFromField(obj, "symbol"),
			Position: position,
			Side:     stringFromField(obj, "side"),
			Entry:    stringFromField(obj, "entry_price", "avg_entry_price"),
			PnL:      stringFromField(obj, "pnl", "unrealized_pnl"),
			Raw:      raw,
		})
	})

	return out
}

func parseTrades(body []byte) []LighterTrade {
	var root any
	if err := json.Unmarshal(body, &root); err != nil {
		return nil
	}
	var out []LighterTrade
	walkObjects(root, func(obj map[string]any) {
		tradeID := stringFromField(obj, "trade_id_str", "trade_id")
		size := stringFromField(obj, "size")
		price := stringFromField(obj, "price")
		if tradeID == "" || size == "" || price == "" {
			return
		}
		raw, _ := json.Marshal(obj)
		out = append(out, LighterTrade{
			TradeID:     tradeID,
			MarketID:    int64FromField(obj, "market_id", "market_index"),
			AskID:       stringFromField(obj, "ask_id_str", "ask_id"),
			BidID:       stringFromField(obj, "bid_id_str", "bid_id"),
			AskClientID: stringFromField(obj, "ask_client_id_str", "ask_client_id"),
			BidClientID: stringFromField(obj, "bid_client_id_str", "bid_client_id"),
			Type:        stringFromField(obj, "type"),
			Side:        stringFromField(obj, "side"),
			Size:        size,
			Price:       price,
			Timestamp:   int64FromField(obj, "timestamp", "transaction_time"),
			Raw:         raw,
		})
	})
	return out
}

func ParseAccountWSState(body []byte) (LighterAccountWSState, error) {
	var root any
	if err := json.Unmarshal(body, &root); err != nil {
		return LighterAccountWSState{}, err
	}
	return LighterAccountWSState{
		Orders:    parseActiveOrders(body),
		Trades:    parseTrades(body),
		Positions: parsePositions(body),
	}, nil
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

func boolFromField(obj map[string]any, names ...string) bool {
	for _, name := range names {
		switch value := obj[name].(type) {
		case bool:
			return value
		case string:
			return strings.EqualFold(strings.TrimSpace(value), "true") || strings.TrimSpace(value) == "1"
		case float64:
			return value != 0
		}
	}
	return false
}

func lighterSymbolFromMarketID(marketID int64) string {
	switch marketID {
	case 0:
		return "ETH"
	case 1:
		return "BTC"
	default:
		return strconv.FormatInt(marketID, 10)
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
