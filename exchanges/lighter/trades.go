package lighter

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"AlgoTrading2026/tradetape"
)

type LighterRawTrade map[string]any

func GetRecentTrades(symbol string) ([]tradetape.TradeTapePrint, error) {
	marketID, normalized, err := lighterMarketID(symbol)
	if err != nil {
		return nil, err
	}
	return getRecentTradesAt(getBaseURL(), int(marketID), normalized, 100)
}

func GetTrades(symbol string, limit int) ([]tradetape.TradeTapePrint, error) {
	marketID, normalized, err := lighterMarketID(symbol)
	if err != nil {
		return nil, err
	}
	if limit <= 0 || limit > 100 {
		limit = 100
	}
	return getTradesAt(getBaseURL(), int(marketID), normalized, limit)
}

func getRecentTradesAt(baseURL string, marketID int, symbol string, limit int) ([]tradetape.TradeTapePrint, error) {
	if limit <= 0 || limit > 100 {
		limit = 100
	}
	params := url.Values{}
	params.Set("market_id", strconv.Itoa(marketID))
	params.Set("limit", strconv.Itoa(limit))
	return getLighterTradesEndpoint(baseURL, "/api/v1/recentTrades", params, marketID, symbol)
}

func getTradesAt(baseURL string, marketID int, symbol string, limit int) ([]tradetape.TradeTapePrint, error) {
	if limit <= 0 || limit > 100 {
		limit = 100
	}
	params := url.Values{}
	params.Set("market_id", strconv.Itoa(marketID))
	params.Set("limit", strconv.Itoa(limit))
	params.Set("sort_by", "timestamp")
	params.Set("sort_dir", "desc")
	params.Set("type", "trade")
	return getLighterTradesEndpoint(baseURL, "/api/v1/trades", params, marketID, symbol)
}

func getLighterTradesEndpoint(baseURL string, path string, params url.Values, marketID int, symbol string) ([]tradetape.TradeTapePrint, error) {
	resp, err := http.Get(baseURL + path + "?" + params.Encode())
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
	return decodeLighterTrades(body, marketID, symbol)
}

func NormalizeLighterTrade(raw LighterRawTrade) (tradetape.TradeTapePrint, error) {
	print := normalizeLighterTradeWithContext(raw, 0, "")
	if err := tradetape.ValidatePrint(print); err != nil {
		return tradetape.TradeTapePrint{}, err
	}
	return print, nil
}

func decodeLighterTrades(body []byte, marketID int, symbol string) ([]tradetape.TradeTapePrint, error) {
	var root any
	if err := json.Unmarshal(body, &root); err != nil {
		return nil, err
	}
	rawTrades := findLighterTradeNodes(root)
	out := make([]tradetape.TradeTapePrint, 0, len(rawTrades))
	for _, raw := range rawTrades {
		print := normalizeLighterTradeWithContext(raw, marketID, symbol)
		if err := tradetape.ValidatePrint(print); err != nil {
			return nil, err
		}
		out = append(out, print)
	}
	return out, nil
}

func normalizeLighterTradeWithContext(raw LighterRawTrade, marketID int, symbol string) tradetape.TradeTapePrint {
	marketIDText := lighterFieldString(raw, "market_id", "market_index", "marketID", "market")
	if marketIDText == "" && marketID > 0 {
		marketIDText = strconv.Itoa(marketID)
	}
	tradeSymbol := strings.ToUpper(lighterFieldString(raw, "symbol", "coin", "market_symbol", "marketSymbol", "ticker"))
	if tradeSymbol == "" {
		tradeSymbol = strings.ToUpper(symbol)
	}
	side := strings.ToUpper(lighterFieldString(raw, "side", "direction", "trade_side", "tradeSide", "ask_bid", "ask_filter", "is_ask", "isAsk"))
	return tradetape.TradeTapePrint{
		Venue:         "lighter",
		Symbol:        tradeSymbol,
		MarketID:      marketIDText,
		Timestamp:     int64Field(raw, "timestamp", "time", "ts", "created_at", "createdAt", "executed_at", "executedAt"),
		Price:         floatField(raw, "price", "px", "p", "execution_price"),
		Size:          floatField(raw, "size", "sz", "qty", "quantity", "amount", "base_amount", "baseAmount", "trade_size"),
		Side:          side,
		AggressorSide: tradetape.UnknownAggressorSide,
		TradeID:       lighterFieldString(raw, "trade_id", "tradeID", "id", "order_id", "orderIndex", "order_index"),
	}
}

func findLighterTradeNodes(root any) []LighterRawTrade {
	switch typed := root.(type) {
	case []any:
		return lighterTradeNodesFromArray(typed)
	case map[string]any:
		for _, name := range []string{"trades", "recent_trades", "recentTrades", "data", "items", "results"} {
			if values, ok := findArrayField(typed, name); ok {
				if trades := lighterTradeNodesFromArray(values); len(trades) > 0 {
					return trades
				}
			}
		}
		if looksLikeLighterTrade(typed) {
			return []LighterRawTrade{typed}
		}
		for _, child := range typed {
			if trades := findLighterTradeNodes(child); len(trades) > 0 {
				return trades
			}
		}
	}
	return nil
}

func lighterTradeNodesFromArray(values []any) []LighterRawTrade {
	out := make([]LighterRawTrade, 0, len(values))
	for _, value := range values {
		node, ok := value.(map[string]any)
		if !ok || !looksLikeLighterTrade(node) {
			continue
		}
		out = append(out, node)
	}
	return out
}

func looksLikeLighterTrade(node map[string]any) bool {
	hasPrice := lighterFieldString(node, "price", "px", "p", "execution_price") != ""
	hasSize := lighterFieldString(node, "size", "sz", "qty", "quantity", "amount", "base_amount", "baseAmount", "trade_size") != ""
	hasTime := int64Field(node, "timestamp", "time", "ts", "created_at", "createdAt", "executed_at", "executedAt") > 0
	return hasPrice && hasSize && hasTime
}

func floatField(node map[string]any, names ...string) float64 {
	for _, name := range names {
		if value, ok := node[name]; ok {
			switch typed := value.(type) {
			case float64:
				return typed
			case string:
				parsed, _ := strconv.ParseFloat(strings.TrimSpace(typed), 64)
				return parsed
			case json.Number:
				parsed, _ := typed.Float64()
				return parsed
			}
		}
	}
	return 0
}
