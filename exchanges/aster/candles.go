package aster

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"time"

	"AlgoTrading2026/exchanges"
)

const maxAsterKlineLimit = 1500

func (c *Client) GetCandles(
	symbol string,
	interval string,
	limit int,
) ([]exchanges.Candle, error) {
	if limit <= 0 {
		limit = 100
	}
	if limit > maxAsterKlineLimit {
		return c.getCandlesPaged(symbol, interval, limit)
	}

	return c.getCandles(symbol, interval, limit, 0, 0)
}

func (c *Client) getCandles(symbol string, interval string, limit int, startTime int64, endTime int64) ([]exchanges.Candle, error) {
	params := url.Values{}
	params.Set("symbol", symbol)
	params.Set("interval", interval)
	params.Set("limit", strconv.Itoa(limit))
	if startTime > 0 {
		params.Set("startTime", strconv.FormatInt(startTime, 10))
	}
	if endTime > 0 {
		params.Set("endTime", strconv.FormatInt(endTime, 10))
	}
	endpoint := getBaseURL() + "/fapi/v3/klines?" + params.Encode()

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
		return nil, fmt.Errorf("aster candles bad status %d: %s", resp.StatusCode, string(body))
	}

	var raw [][]any
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, err
	}

	candles := make([]exchanges.Candle, 0, len(raw))

	for _, row := range raw {
		if len(row) < 7 {
			continue
		}

		candles = append(candles, exchanges.Candle{
			Venue:     c.Name(),
			Symbol:    symbol,
			Interval:  interval,
			Open:      fmt.Sprint(row[1]),
			High:      fmt.Sprint(row[2]),
			Low:       fmt.Sprint(row[3]),
			Close:     fmt.Sprint(row[4]),
			Volume:    fmt.Sprint(row[5]),
			StartTime: anyToInt64(row[0]),
			EndTime:   anyToInt64(row[6]),
			Closed:    true,
		})
	}

	return candles, nil
}

func (c *Client) getCandlesPaged(symbol string, interval string, limit int) ([]exchanges.Candle, error) {
	msPerCandle := asterIntervalToMillis(interval)
	if msPerCandle == 0 {
		return nil, fmt.Errorf("unsupported aster interval: %s", interval)
	}

	endTime := time.Now().UnixMilli()
	startTime := endTime - int64(limit)*msPerCandle
	out := make([]exchanges.Candle, 0, limit)
	seen := make(map[int64]bool, limit)

	for len(out) < limit && startTime < endTime {
		remaining := limit - len(out)
		chunkLimit := remaining
		if chunkLimit > maxAsterKlineLimit {
			chunkLimit = maxAsterKlineLimit
		}

		chunk, err := c.getCandles(symbol, interval, chunkLimit, startTime, endTime)
		if err != nil {
			if len(out) > 0 {
				break
			}
			return nil, err
		}
		if len(chunk) == 0 {
			break
		}

		lastStart := startTime
		for _, candle := range chunk {
			if !seen[candle.StartTime] {
				out = append(out, candle)
				seen[candle.StartTime] = true
			}
			if candle.StartTime > lastStart {
				lastStart = candle.StartTime
			}
		}
		nextStart := lastStart + msPerCandle
		if nextStart <= startTime {
			break
		}
		startTime = nextStart
	}

	sort.Slice(out, func(i int, j int) bool {
		return out[i].StartTime < out[j].StartTime
	})
	if len(out) > limit {
		out = out[len(out)-limit:]
	}
	return out, nil
}

func asterIntervalToMillis(interval string) int64 {
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
	case "1d":
		return int64(24 * time.Hour / time.Millisecond)
	default:
		return 0
	}
}

func anyToInt64(v any) int64 {
	switch x := v.(type) {
	case float64:
		return int64(x)
	case int64:
		return x
	case int:
		return int64(x)
	case string:
		n, _ := strconv.ParseInt(x, 10, 64)
		return n
	default:
		return 0
	}
}
