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

type LighterCandlesResponse struct {
	Code       int             `json:"code"`
	Resolution string          `json:"r"`
	Candles    []LighterCandle `json:"c"`
}

type LighterCandle struct {
	Timestamp int64   `json:"t"`
	Open      float64 `json:"o"`
	High      float64 `json:"h"`
	Low       float64 `json:"l"`
	Close     float64 `json:"c"`
	Volume    float64 `json:"v"`
	USDVolume float64 `json:"V"`
}

func (c *Client) GetCandles(
	symbol string,
	interval string,
	limit int,
) ([]exchanges.Candle, error) {
	switch strings.ToUpper(symbol) {
	case "ETH", "ETHUSDT":
		return c.GetCandlesByIndex(0, "ETH", interval, limit)
	case "BTC", "BTCUSDT":
		return c.GetCandlesByIndex(1, "BTC", interval, limit)
	default:
		return nil, fmt.Errorf("unsupported lighter candle symbol: %s", symbol)
	}
}

func (c *Client) GetCandlesByIndex(
	marketID int,
	symbol string,
	interval string,
	limit int,
) ([]exchanges.Candle, error) {
	if limit <= 0 {
		limit = 100
	}

	msPerCandle := lighterIntervalToMillis(interval)
	if msPerCandle == 0 {
		return nil, fmt.Errorf("unsupported lighter interval: %s", interval)
	}

	endMS := time.Now().UnixMilli()
	startMS := endMS - int64(limit)*msPerCandle

	params := url.Values{}
	params.Set("market_id", strconv.Itoa(marketID))
	params.Set("resolution", interval)
	params.Set("start_timestamp", strconv.FormatInt(startMS, 10))
	params.Set("end_timestamp", strconv.FormatInt(endMS, 10))
	params.Set("count_back", strconv.Itoa(limit))
	params.Set("set_timestamp_to_end", "false")

	endpoint := getBaseURL() + "/api/v1/candles?" + params.Encode()

	resp, err := http.Get(endpoint)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("lighter candles bad status %d: %s", resp.StatusCode, string(body))
	}

	var parsed LighterCandlesResponse
	if err := json.Unmarshal(body, &parsed); err != nil {
		return nil, err
	}

	candles := make([]exchanges.Candle, 0, len(parsed.Candles))

	for _, k := range parsed.Candles {
		candles = append(candles, exchanges.Candle{
			Venue:     c.Name(),
			Symbol:    symbol,
			Interval:  parsed.Resolution,
			Open:      fmt.Sprintf("%f", k.Open),
			High:      fmt.Sprintf("%f", k.High),
			Low:       fmt.Sprintf("%f", k.Low),
			Close:     fmt.Sprintf("%f", k.Close),
			Volume:    fmt.Sprintf("%f", k.Volume),
			StartTime: k.Timestamp,
			EndTime:   k.Timestamp + msPerCandle,
			Closed:    true,
		})
	}

	return candles, nil
}

func lighterIntervalToMillis(interval string) int64 {
	switch interval {
	case "1m":
		return int64(time.Minute / time.Millisecond)
	case "5m":
		return int64(5 * time.Minute / time.Millisecond)
	case "15m":
		return int64(15 * time.Minute / time.Millisecond)
	case "30m":
		return int64(30 * time.Minute / time.Millisecond)
	case "1h":
		return int64(time.Hour / time.Millisecond)
	case "4h":
		return int64(4 * time.Hour / time.Millisecond)
	case "12h":
		return int64(12 * time.Hour / time.Millisecond)
	case "1d":
		return int64(24 * time.Hour / time.Millisecond)
	case "1w":
		return int64(7 * 24 * time.Hour / time.Millisecond)
	default:
		return 0
	}
}
