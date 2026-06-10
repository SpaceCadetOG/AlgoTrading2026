package hyperliquid

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	"AlgoTrading2026/tradetape"
)

type HyperliquidRawTrade struct {
	Coin string `json:"coin"`
	Side string `json:"side"`
	Px   string `json:"px"`
	Sz   string `json:"sz"`
	Time int64  `json:"time"`
	Hash string `json:"hash"`
	TID  any    `json:"tid"`
}

type UserFill struct {
	ClosedPnl     string `json:"closedPnl"`
	Coin          string `json:"coin"`
	Crossed       bool   `json:"crossed"`
	Dir           string `json:"dir"`
	Hash          string `json:"hash"`
	Oid           int64  `json:"oid"`
	Price         string `json:"px"`
	Side          string `json:"side"`
	StartPosition string `json:"startPosition"`
	Size          string `json:"sz"`
	Time          int64  `json:"time"`
	Fee           string `json:"fee"`
	FeeToken      string `json:"feeToken"`
	Tid           int64  `json:"tid"`
}

func parseTradeFloat(value string) (float64, error) {
	return strconv.ParseFloat(strings.TrimSpace(value), 64)
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" && strings.TrimSpace(value) != "<nil>" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}

func GetRecentTrades(symbol string) ([]tradetape.TradeTapePrint, error) {
	return getRecentTradesAt(getBaseURL(), symbol)
}

func (c *Client) GetUserFills() ([]UserFill, error) {
	var result []UserFill
	jsonBody := fmt.Sprintf(`{"type":"userFills","user":"%s"}`, c.Address)
	if err := postInfoJSON(jsonBody, &result); err != nil {
		return nil, err
	}
	return result, nil
}

func getRecentTradesAt(baseURL string, symbol string) ([]tradetape.TradeTapePrint, error) {
	body := fmt.Sprintf(`{"type":"recentTrades","coin":"%s"}`, symbol)
	resp, err := http.Post(baseURL+"/info", "application/json", bytes.NewBufferString(body))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("hyperliquid recent trades bad status %d: %s", resp.StatusCode, string(respBody))
	}

	var raw []HyperliquidRawTrade
	if err := json.Unmarshal(respBody, &raw); err != nil {
		return nil, err
	}
	out := make([]tradetape.TradeTapePrint, 0, len(raw))
	for _, trade := range raw {
		print, err := NormalizeHyperliquidTrade(trade)
		if err != nil {
			return nil, err
		}
		out = append(out, print)
	}
	return out, nil
}

func NormalizeHyperliquidTrade(raw HyperliquidRawTrade) (tradetape.TradeTapePrint, error) {
	side := strings.ToUpper(strings.TrimSpace(raw.Side))
	price, err := parseTradeFloat(raw.Px)
	if err != nil {
		return tradetape.TradeTapePrint{}, fmt.Errorf("invalid hyperliquid price: %w", err)
	}
	size, err := parseTradeFloat(raw.Sz)
	if err != nil {
		return tradetape.TradeTapePrint{}, fmt.Errorf("invalid hyperliquid size: %w", err)
	}
	print := tradetape.TradeTapePrint{
		Venue:         "hyperliquid",
		Symbol:        raw.Coin,
		Timestamp:     raw.Time,
		Price:         price,
		Size:          size,
		Side:          side,
		AggressorSide: side,
		TradeID:       firstNonEmpty(raw.Hash, fmt.Sprint(raw.TID)),
	}
	if err := tradetape.ValidatePrint(print); err != nil {
		return tradetape.TradeTapePrint{}, err
	}
	return print, nil
}
