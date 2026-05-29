package hyperliquid

import (
	"fmt"
	"time"

	"AlgoTrading2026/exchanges"
)

type HLCandle struct {
	T      int64  `json:"t"`
	TClose int64  `json:"T"`
	S      string `json:"s"`
	I      string `json:"i"`
	O      string `json:"o"`
	C      string `json:"c"`
	H      string `json:"h"`
	L      string `json:"l"`
	V      string `json:"v"`
	N      int64  `json:"n"`
}

func (c *Client) GetCandles(
	symbol string,
	interval string,
	limit int,
) ([]exchanges.Candle, error) {
	if limit <= 0 {
		limit = 100
	}

	msPerCandle := intervalToMillis(interval)
	if msPerCandle == 0 {
		return nil, fmt.Errorf("unsupported interval: %s", interval)
	}

	endTime := time.Now().UnixMilli()
	startTime := endTime - int64(limit)*msPerCandle

	body := fmt.Sprintf(
		`{"type":"candleSnapshot","req":{"coin":"%s","interval":"%s","startTime":%d,"endTime":%d}}`,
		symbol,
		interval,
		startTime,
		endTime,
	)

	var raw []HLCandle
	if err := postInfoJSON(body, &raw); err != nil {
		return nil, err
	}

	candles := make([]exchanges.Candle, 0, len(raw))

	for _, k := range raw {
		candles = append(candles, exchanges.Candle{
			Venue:     c.Name(),
			Symbol:    k.S,
			Interval:  k.I,
			Open:      k.O,
			High:      k.H,
			Low:       k.L,
			Close:     k.C,
			Volume:    k.V,
			StartTime: k.T,
			EndTime:   k.TClose,
			Closed:    true,
		})
	}

	return candles, nil
}

func intervalToMillis(interval string) int64 {
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
