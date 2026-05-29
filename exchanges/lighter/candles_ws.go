package lighter

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"github.com/gorilla/websocket"

	"AlgoTrading2026/exchanges"
)

type LighterCandleWSMessage struct {
	Channel string            `json:"channel"`
	Type    string            `json:"type"`
	Candles []LighterWSCandle `json:"candles"`
}

type LighterWSCandle struct {
	Timestamp int64   `json:"t"`
	Open      float64 `json:"o"`
	High      float64 `json:"h"`
	Low       float64 `json:"l"`
	Close     float64 `json:"c"`
	Volume    float64 `json:"v"`
	USDVolume float64 `json:"V"`
}

func (c *Client) StreamCandles(symbol string, interval string, handler exchanges.CandleHandler) error {
	switch strings.ToUpper(symbol) {
	case "ETH", "ETHUSDT":
		return c.StreamCandlesByIndex(0, "ETH", interval, handler)
	case "BTC", "BTCUSDT":
		return c.StreamCandlesByIndex(1, "BTC", interval, handler)
	default:
		return fmt.Errorf("unsupported lighter candle symbol: %s", symbol)
	}
}

func (c *Client) StreamCandlesByIndex(
	marketID int,
	symbol string,
	interval string,
	handler exchanges.CandleHandler,
) error {
	conn, _, err := websocket.DefaultDialer.Dial(getWSURL(), nil)
	if err != nil {
		return err
	}
	defer conn.Close()

	subscribeMsg := fmt.Sprintf(`{"type":"subscribe","channel":"candle/%d/%s"}`, marketID, interval)

	if err := conn.WriteMessage(websocket.TextMessage, []byte(subscribeMsg)); err != nil {
		return err
	}

	log.Println("Connected to Lighter WS candles:", marketID, interval)

	go keepAlive(conn)

	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			return err
		}

		var candleMsg LighterCandleWSMessage
		if err := json.Unmarshal(msg, &candleMsg); err != nil {
			fmt.Println("Lighter candle decode error:", err)
			continue
		}

		if !strings.HasPrefix(candleMsg.Channel, "candle:") {
			continue
		}

		if len(candleMsg.Candles) == 0 {
			continue
		}

		k := candleMsg.Candles[len(candleMsg.Candles)-1]

		handler(exchanges.Candle{
			Venue:     c.Name(),
			Symbol:    symbol,
			Interval:  interval,
			Open:      fmt.Sprintf("%f", k.Open),
			High:      fmt.Sprintf("%f", k.High),
			Low:       fmt.Sprintf("%f", k.Low),
			Close:     fmt.Sprintf("%f", k.Close),
			Volume:    fmt.Sprintf("%f", k.Volume),
			StartTime: k.Timestamp,
			EndTime:   k.Timestamp + lighterIntervalToMillis(interval),
			Closed:    false,
		})
	}
}
