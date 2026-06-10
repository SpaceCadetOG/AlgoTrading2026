package lighter

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/websocket"

	"AlgoTrading2026/orderbook"
)

func (c *Client) GetOrderBook(symbol string) (orderbook.OrderBookSnapshot, error) {
	return GetOrderBook(symbol)
}

func GetOrderBook(symbol string) (orderbook.OrderBookSnapshot, error) {
	marketID, normalized, err := lighterMarketID(symbol)
	if err != nil {
		return orderbook.OrderBookSnapshot{}, err
	}
	snapshot, err := getOrderBookFromBaseURL(getBaseURL(), normalized, int(marketID))
	if err == nil {
		return snapshot, nil
	}
	return getOrderBookViaWSAt(getWSURL(), int(marketID), normalized)
}

func GetMainnetOrderBook(symbol string) (orderbook.OrderBookSnapshot, error) {
	marketID, normalized, err := lighterMarketID(symbol)
	if err != nil {
		return orderbook.OrderBookSnapshot{}, err
	}
	return GetMainnetOrderBookByMarketID(normalized, int(marketID))
}

func GetOrderBookByMarketID(symbol string, marketID int) (orderbook.OrderBookSnapshot, error) {
	normalized := strings.ToUpper(strings.TrimSpace(symbol))
	if normalized == "" {
		return orderbook.OrderBookSnapshot{}, fmt.Errorf("lighter symbol is required")
	}
	snapshot, err := getOrderBookFromBaseURL(getBaseURL(), normalized, marketID)
	if err == nil {
		return snapshot, nil
	}
	return getOrderBookViaWSAt(getWSURL(), marketID, normalized)
}

func GetMainnetOrderBookByMarketID(symbol string, marketID int) (orderbook.OrderBookSnapshot, error) {
	normalized := strings.ToUpper(strings.TrimSpace(symbol))
	if normalized == "" {
		return orderbook.OrderBookSnapshot{}, fmt.Errorf("lighter symbol is required")
	}
	snapshot, err := getOrderBookFromBaseURL("https://mainnet.zklighter.elliot.ai", normalized, int(marketID))
	if err == nil {
		return snapshot, nil
	}
	return getOrderBookViaWSAt("wss://mainnet.zklighter.elliot.ai/stream?readonly=true", int(marketID), normalized)
}

func getOrderBookFromBaseURL(baseURL string, symbol string, defaultMarketID int) (orderbook.OrderBookSnapshot, error) {
	var lastErr error
	for _, candidateID := range lighterOrderBookMarketIDs(symbol, defaultMarketID) {
		snapshot, err := getOrderBookAt(baseURL, candidateID, symbol)
		if err == nil {
			return snapshot, nil
		}
		lastErr = err
	}
	return orderbook.OrderBookSnapshot{}, lastErr
}

func getOrderBookViaWSAt(wsURL string, marketID int, symbol string) (orderbook.OrderBookSnapshot, error) {
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		return orderbook.OrderBookSnapshot{}, err
	}
	defer conn.Close()

	if err := conn.SetReadDeadline(time.Now().Add(10 * time.Second)); err != nil {
		return orderbook.OrderBookSnapshot{}, err
	}

	subscribe := fmt.Sprintf(`{"type":"subscribe","channel":"order_book/%d"}`, marketID)
	if err := conn.WriteMessage(websocket.TextMessage, []byte(subscribe)); err != nil {
		return orderbook.OrderBookSnapshot{}, err
	}

	var lastErr error
	for i := 0; i < 20; i++ {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			if lastErr != nil {
				return orderbook.OrderBookSnapshot{}, lastErr
			}
			return orderbook.OrderBookSnapshot{}, err
		}

		if !strings.Contains(string(msg), "order_book") {
			continue
		}
		snapshot, err := decodeLighterOrderBook(msg, symbol)
		if err != nil {
			lastErr = err
			continue
		}
		if err := orderbook.ValidateSnapshot(snapshot); err != nil {
			lastErr = err
			continue
		}
		return snapshot, nil
	}
	if lastErr != nil {
		return orderbook.OrderBookSnapshot{}, lastErr
	}
	return orderbook.OrderBookSnapshot{}, fmt.Errorf("lighter websocket order book snapshot not received")
}

func lighterOrderBookMarketIDs(symbol string, defaultID int) []int {
	seen := map[int]bool{}
	out := make([]int, 0, 3)
	add := func(id int) {
		if !seen[id] {
			out = append(out, id)
			seen[id] = true
		}
	}

	add(defaultID)
	switch strings.ToUpper(symbol) {
	case "BTC", "BTCUSDT", "BTC-USD":
		add(2)
		add(1)
	case "ETH", "ETHUSDT", "ETH-USD":
		add(1)
		add(0)
	case "SOL", "SOLUSDT", "SOL-USD":
		add(3)
	}
	return out
}

func getOrderBookAt(baseURL string, marketID int, symbol string) (orderbook.OrderBookSnapshot, error) {
	params := url.Values{}
	params.Set("market_id", strconv.Itoa(marketID))

	resp, err := http.Get(baseURL + "/api/v1/orderBook?" + params.Encode())
	if err != nil {
		return orderbook.OrderBookSnapshot{}, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return orderbook.OrderBookSnapshot{}, err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return getOrderBookDetailsAt(baseURL, marketID, symbol)
	}

	snapshot, err := decodeLighterOrderBook(body, symbol)
	if err != nil {
		return orderbook.OrderBookSnapshot{}, err
	}
	if err := orderbook.ValidateSnapshot(snapshot); err != nil {
		return getOrderBookDetailsAt(baseURL, marketID, symbol)
	}
	return snapshot, nil
}

func getOrderBookDetailsAt(baseURL string, marketID int, symbol string) (orderbook.OrderBookSnapshot, error) {
	params := url.Values{}
	params.Set("market_id", strconv.Itoa(marketID))
	params.Set("depth", "100")

	resp, err := http.Get(baseURL + "/api/v1/orderBookDetails?" + params.Encode())
	if err != nil {
		return orderbook.OrderBookSnapshot{}, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return orderbook.OrderBookSnapshot{}, err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return getOrderBookOrdersAt(baseURL, marketID, symbol)
	}

	snapshot, err := decodeLighterOrderBook(body, symbol)
	if err != nil {
		return orderbook.OrderBookSnapshot{}, err
	}
	if err := orderbook.ValidateSnapshot(snapshot); err != nil {
		return getOrderBookOrdersAt(baseURL, marketID, symbol)
	}
	return snapshot, nil
}

func getOrderBookOrdersAt(baseURL string, marketID int, symbol string) (orderbook.OrderBookSnapshot, error) {
	params := url.Values{}
	params.Set("market_id", strconv.Itoa(marketID))
	params.Set("limit", "250")

	resp, err := http.Get(baseURL + "/api/v1/orderBookOrders?" + params.Encode())
	if err != nil {
		return orderbook.OrderBookSnapshot{}, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return orderbook.OrderBookSnapshot{}, err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return orderbook.OrderBookSnapshot{}, fmt.Errorf("lighter orderBook/orderBookOrders bad status %d: %s", resp.StatusCode, string(body))
	}

	snapshot, err := decodeLighterOrderBookOrders(body, symbol)
	if err != nil {
		return orderbook.OrderBookSnapshot{}, err
	}
	if err := orderbook.ValidateSnapshot(snapshot); err != nil {
		return orderbook.OrderBookSnapshot{}, err
	}
	return snapshot, nil
}

func decodeLighterOrderBook(body []byte, symbol string) (orderbook.OrderBookSnapshot, error) {
	var root any
	if err := json.Unmarshal(body, &root); err != nil {
		return orderbook.OrderBookSnapshot{}, err
	}

	bids, asks := findLighterSides(root)
	snapshot := orderbook.OrderBookSnapshot{
		Venue:      "lighter",
		Symbol:     symbol,
		Time:       findLighterTime(root),
		Bids:       normalizeLighterLevels(bids),
		Asks:       normalizeLighterLevels(asks),
		IsSnapshot: true,
	}
	if snapshot.Time > 0 {
		snapshot.Sequence = strconv.FormatInt(snapshot.Time, 10)
	}
	return snapshot, nil
}

func decodeLighterOrderBookOrders(body []byte, symbol string) (orderbook.OrderBookSnapshot, error) {
	var root any
	if err := json.Unmarshal(body, &root); err != nil {
		return orderbook.OrderBookSnapshot{}, err
	}

	orders := findLighterOrders(root)
	bids := make(map[string]float64)
	asks := make(map[string]float64)
	for _, rawOrder := range orders {
		order, ok := rawOrder.(map[string]any)
		if !ok {
			continue
		}
		price := lighterFieldString(order, "price", "px", "p")
		size := lighterFieldString(order, "size", "sz", "qty", "quantity", "remaining_base_amount", "base_amount", "amount")
		if price == "" || size == "" {
			continue
		}
		sizeFloat, err := strconv.ParseFloat(size, 64)
		if err != nil || sizeFloat <= 0 {
			continue
		}
		isAsk, ok := lighterOrderSide(order)
		if !ok {
			continue
		}
		if isAsk {
			asks[price] += sizeFloat
		} else {
			bids[price] += sizeFloat
		}
	}

	return orderbook.OrderBookSnapshot{
		Venue:      "lighter",
		Symbol:     symbol,
		Time:       findLighterTime(root),
		Bids:       lighterAggregatedLevels(bids, true),
		Asks:       lighterAggregatedLevels(asks, false),
		IsSnapshot: true,
	}, nil
}

func findLighterOrders(root any) []any {
	if node, ok := root.(map[string]any); ok {
		for _, name := range []string{"orders", "order_book_orders", "orderBookOrders"} {
			if values, ok := findArrayField(node, name); ok {
				return values
			}
		}
		for _, value := range node {
			if orders := findLighterOrders(value); len(orders) > 0 {
				return orders
			}
		}
	}
	if values, ok := root.([]any); ok {
		if len(values) == 0 {
			return nil
		}
		if _, ok := values[0].(map[string]any); ok {
			return values
		}
		for _, value := range values {
			if orders := findLighterOrders(value); len(orders) > 0 {
				return orders
			}
		}
	}
	return nil
}

func lighterOrderSide(order map[string]any) (bool, bool) {
	for _, name := range []string{"is_ask", "isAsk", "ask"} {
		if value, ok := order[name]; ok {
			switch typed := value.(type) {
			case bool:
				return typed, true
			case float64:
				return typed != 0, true
			case string:
				normalized := strings.ToLower(strings.TrimSpace(typed))
				if normalized == "true" || normalized == "1" || normalized == "ask" || normalized == "sell" {
					return true, true
				}
				if normalized == "false" || normalized == "0" || normalized == "bid" || normalized == "buy" {
					return false, true
				}
			}
		}
	}
	side := strings.ToLower(lighterFieldString(order, "side", "direction", "order_side", "orderSide"))
	switch side {
	case "ask", "sell", "s", "short":
		return true, true
	case "bid", "buy", "b", "long":
		return false, true
	default:
		return false, false
	}
}

func lighterAggregatedLevels(levels map[string]float64, bids bool) []orderbook.BookLevel {
	out := make([]orderbook.BookLevel, 0, len(levels))
	for price, size := range levels {
		out = append(out, orderbook.BookLevel{
			Price: price,
			Size:  strconv.FormatFloat(size, 'f', -1, 64),
		})
	}
	sort.Slice(out, func(i int, j int) bool {
		if bids {
			return out[i].PriceFloat() > out[j].PriceFloat()
		}
		return out[i].PriceFloat() < out[j].PriceFloat()
	})
	return out
}

func findLighterSides(root any) ([]any, []any) {
	if node, ok := root.(map[string]any); ok {
		if bids, ok := findArrayField(node, "bids", "bid"); ok {
			if asks, ok := findArrayField(node, "asks", "ask"); ok {
				return bids, asks
			}
		}
		for _, value := range node {
			if bids, asks := findLighterSides(value); len(bids) > 0 || len(asks) > 0 {
				return bids, asks
			}
		}
	}
	if nodes, ok := root.([]any); ok {
		for _, node := range nodes {
			if bids, asks := findLighterSides(node); len(bids) > 0 || len(asks) > 0 {
				return bids, asks
			}
		}
	}
	return nil, nil
}

func findArrayField(node map[string]any, names ...string) ([]any, bool) {
	for _, name := range names {
		value, ok := node[name]
		if !ok {
			continue
		}
		values, ok := value.([]any)
		if ok {
			return values, true
		}
	}
	return nil, false
}

func normalizeLighterLevels(raw []any) []orderbook.BookLevel {
	out := make([]orderbook.BookLevel, 0, len(raw))
	for _, value := range raw {
		level, ok := normalizeLighterLevel(value)
		if ok {
			out = append(out, level)
		}
	}
	return out
}

func normalizeLighterLevel(value any) (orderbook.BookLevel, bool) {
	switch typed := value.(type) {
	case []any:
		if len(typed) < 2 {
			return orderbook.BookLevel{}, false
		}
		return orderbook.BookLevel{Price: lighterString(typed[0]), Size: lighterString(typed[1])}, true
	case map[string]any:
		price := lighterFieldString(typed, "price", "px", "p")
		size := lighterFieldString(typed, "size", "sz", "qty", "quantity", "amount")
		if price == "" || size == "" {
			return orderbook.BookLevel{}, false
		}
		return orderbook.BookLevel{Price: price, Size: size}, true
	default:
		return orderbook.BookLevel{}, false
	}
}

func lighterFieldString(node map[string]any, names ...string) string {
	for _, name := range names {
		if value, ok := node[name]; ok {
			out := lighterString(value)
			if out != "" {
				return out
			}
		}
	}
	return ""
}

func lighterString(value any) string {
	switch typed := value.(type) {
	case string:
		return typed
	case float64:
		return strconv.FormatFloat(typed, 'f', -1, 64)
	case json.Number:
		return typed.String()
	default:
		return strings.TrimSpace(fmt.Sprint(typed))
	}
}

func findLighterTime(root any) int64 {
	switch typed := root.(type) {
	case map[string]any:
		if value := int64Field(typed, "timestamp", "time", "ts"); value > 0 {
			return value
		}
		for _, child := range typed {
			if value := findLighterTime(child); value > 0 {
				return value
			}
		}
	case []any:
		for _, child := range typed {
			if value := findLighterTime(child); value > 0 {
				return value
			}
		}
	}
	return 0
}

func int64Field(node map[string]any, names ...string) int64 {
	for _, name := range names {
		value, ok := node[name]
		if !ok {
			continue
		}
		switch typed := value.(type) {
		case float64:
			return int64(typed)
		case string:
			parsed, _ := strconv.ParseInt(typed, 10, 64)
			return parsed
		case json.Number:
			parsed, _ := typed.Int64()
			return parsed
		}
	}
	return 0
}
