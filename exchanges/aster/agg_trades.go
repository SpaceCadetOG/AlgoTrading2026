package aster

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"

	"AlgoTrading2026/tradetape"
)

const (
	asterAggTradeLimit     = 1000
	asterAggTradeWindowMS  = int64(59*60*1000 + 59*1000)
	asterAggTradesEndpoint = "/fapi/v3/aggTrades"
)

type AsterAggTrade struct {
	AggregateTradeID int64  `json:"a"`
	Price            string `json:"p"`
	Quantity         string `json:"q"`
	FirstTradeID     int64  `json:"f"`
	LastTradeID      int64  `json:"l"`
	Timestamp        int64  `json:"T"`
	BuyerMaker       bool   `json:"m"`
}

type AggTradeProvider struct{}

func (AggTradeProvider) Venue() string { return "aster" }

func (AggTradeProvider) BackfillTrades(symbol string, startMS int64, endMS int64) ([]tradetape.TradeTapePrint, error) {
	return BackfillAsterAggTrades(symbol, startMS, endMS)
}

func BackfillAsterAggTrades(symbol string, startMS int64, endMS int64) ([]tradetape.TradeTapePrint, error) {
	return backfillAsterAggTradesAt(getBaseURL(), symbol, startMS, endMS)
}

func BackfillAsterAggTradeWindows(startMS int64, endMS int64) [][2]int64 {
	windows := make([][2]int64, 0)
	if startMS <= 0 || endMS <= startMS {
		return windows
	}
	for start := startMS; start < endMS; {
		end := start + asterAggTradeWindowMS
		if end > endMS {
			end = endMS
		}
		windows = append(windows, [2]int64{start, end})
		start = end + 1
	}
	return windows
}

func backfillAsterAggTradesAt(baseURL string, symbol string, startMS int64, endMS int64) ([]tradetape.TradeTapePrint, error) {
	symbol = strings.ToUpper(strings.TrimSpace(symbol))
	if symbol == "" {
		return nil, fmt.Errorf("symbol is required")
	}
	windows := BackfillAsterAggTradeWindows(startMS, endMS)
	out := make([]tradetape.TradeTapePrint, 0)
	for _, window := range windows {
		windowStart := window[0]
		for windowStart <= window[1] {
			raw, err := getAsterAggTradesAt(baseURL, symbol, windowStart, window[1], asterAggTradeLimit)
			if err != nil {
				return nil, err
			}
			if len(raw) == 0 {
				break
			}
			var lastTimestamp int64
			for _, trade := range raw {
				print, err := NormalizeAsterAggTrade(symbol, trade)
				if err != nil {
					return nil, err
				}
				out = append(out, print)
				if trade.Timestamp > lastTimestamp {
					lastTimestamp = trade.Timestamp
				}
			}
			if len(raw) < asterAggTradeLimit || lastTimestamp <= windowStart {
				break
			}
			windowStart = lastTimestamp + 1
		}
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Timestamp == out[j].Timestamp {
			return out[i].TradeID < out[j].TradeID
		}
		return out[i].Timestamp < out[j].Timestamp
	})
	return out, nil
}

func getAsterAggTradesAt(baseURL string, symbol string, startMS int64, endMS int64, limit int) ([]AsterAggTrade, error) {
	if limit <= 0 || limit > asterAggTradeLimit {
		limit = asterAggTradeLimit
	}
	params := url.Values{}
	params.Set("symbol", symbol)
	params.Set("limit", strconv.Itoa(limit))
	if startMS > 0 {
		params.Set("startTime", strconv.FormatInt(startMS, 10))
	}
	if endMS > 0 {
		params.Set("endTime", strconv.FormatInt(endMS, 10))
	}
	resp, err := http.Get(baseURL + asterAggTradesEndpoint + "?" + params.Encode())
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("aster aggTrades bad status %d: %s", resp.StatusCode, string(body))
	}
	var raw []AsterAggTrade
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, err
	}
	return raw, nil
}

func NormalizeAsterAggTrade(symbol string, raw AsterAggTrade) (tradetape.TradeTapePrint, error) {
	price, err := parseTradeFloat(raw.Price)
	if err != nil {
		return tradetape.TradeTapePrint{}, fmt.Errorf("invalid aster aggTrade price: %w", err)
	}
	size, err := parseTradeFloat(raw.Quantity)
	if err != nil {
		return tradetape.TradeTapePrint{}, fmt.Errorf("invalid aster aggTrade size: %w", err)
	}
	// Aster/Binance-style buyer-maker semantics: m=true means the buyer was maker,
	// so the seller crossed the spread and the aggressor side is SELL. m=false
	// means the buyer was taker/aggressor, so the aggressor side is BUY.
	side := "BUY"
	if raw.BuyerMaker {
		side = "SELL"
	}
	print := tradetape.TradeTapePrint{
		Venue:         "aster",
		Symbol:        strings.ToUpper(symbol),
		VenueSymbol:   strings.ToUpper(symbol),
		MarketID:      strings.ToUpper(symbol),
		Timestamp:     raw.Timestamp,
		Price:         price,
		Size:          size,
		Side:          side,
		AggressorSide: side,
		TradeID:       strconv.FormatInt(raw.AggregateTradeID, 10),
	}
	print = tradetape.NormalizePrintSymbols(print)
	if err := tradetape.ValidatePrint(print); err != nil {
		return tradetape.TradeTapePrint{}, err
	}
	return print, nil
}
