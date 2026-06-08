package aster

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

type AsterRawTrade struct {
	ID           int64  `json:"id"`
	TradeID      int64  `json:"t"`
	Price        string `json:"price"`
	PriceShort   string `json:"p"`
	Qty          string `json:"qty"`
	Quantity     string `json:"q"`
	Time         int64  `json:"time"`
	TradeTime    int64  `json:"T"`
	IsBuyerMaker bool   `json:"isBuyerMaker"`
	MakerShort   bool   `json:"m"`
}

func parseTradeFloat(value string) (float64, error) {
	return strconv.ParseFloat(strings.TrimSpace(value), 64)
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}

func GetRecentTrades(symbol string, limit int) ([]tradetape.TradeTapePrint, error) {
	return getRecentTradesAt(getBaseURL(), symbol, limit)
}

func getRecentTradesAt(baseURL string, symbol string, limit int) ([]tradetape.TradeTapePrint, error) {
	if limit <= 0 || limit > 1000 {
		limit = 100
	}
	params := url.Values{}
	params.Set("symbol", strings.ToUpper(symbol))
	params.Set("limit", strconv.Itoa(limit))
	resp, err := http.Get(baseURL + "/fapi/v1/trades?" + params.Encode())
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("aster recent trades bad status %d: %s", resp.StatusCode, string(body))
	}

	var raw []AsterRawTrade
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, err
	}
	out := make([]tradetape.TradeTapePrint, 0, len(raw))
	for _, trade := range raw {
		print, err := NormalizeAsterTrade(symbol, trade)
		if err != nil {
			return nil, err
		}
		out = append(out, print)
	}
	return out, nil
}

func NormalizeAsterTrade(symbol string, raw AsterRawTrade) (tradetape.TradeTapePrint, error) {
	priceText := firstNonEmpty(raw.Price, raw.PriceShort)
	sizeText := firstNonEmpty(raw.Qty, raw.Quantity)
	price, err := parseTradeFloat(priceText)
	if err != nil {
		return tradetape.TradeTapePrint{}, fmt.Errorf("invalid aster price: %w", err)
	}
	size, err := parseTradeFloat(sizeText)
	if err != nil {
		return tradetape.TradeTapePrint{}, fmt.Errorf("invalid aster size: %w", err)
	}
	isBuyerMaker := raw.IsBuyerMaker || raw.MakerShort
	side := "B"
	if isBuyerMaker {
		side = "A"
	}
	tradeID := raw.ID
	if tradeID == 0 {
		tradeID = raw.TradeID
	}
	timestamp := raw.Time
	if timestamp == 0 {
		timestamp = raw.TradeTime
	}
	print := tradetape.TradeTapePrint{
		Venue:         "aster",
		Symbol:        strings.ToUpper(symbol),
		Timestamp:     timestamp,
		Price:         price,
		Size:          size,
		Side:          side,
		AggressorSide: side,
		TradeID:       strconv.FormatInt(tradeID, 10),
	}
	if err := tradetape.ValidatePrint(print); err != nil {
		return tradetape.TradeTapePrint{}, err
	}
	return print, nil
}
