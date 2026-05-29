package lighter

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/gorilla/websocket"

	"AlgoTrading2026/exchanges"
)

const lighterWSBaseURL = "wss://mainnet.zklighter.elliot.ai/stream?readonly=true"

type LighterTickerMessage struct {
	Channel string `json:"channel"`
	Type    string `json:"type"`

	Ticker struct {
		Symbol string `json:"s"`

		Ask struct {
			Price string `json:"price"`
			Size  string `json:"size"`
		} `json:"a"`

		Bid struct {
			Price string `json:"price"`
			Size  string `json:"size"`
		} `json:"b"`

		LastUpdatedAt int64 `json:"last_updated_at"`
	} `json:"ticker"`

	Timestamp int64 `json:"timestamp"`
}

func (c *Client) StreamTicker(
	symbol string,
	handler exchanges.TickerHandler,
) error {
	switch strings.ToUpper(symbol) {
	case "ETH", "ETHUSDT":
		return c.StreamTickerByIndex(0, handler)
	case "BTC", "BTCUSDT":
		return c.StreamTickerByIndex(1, handler)
	default:
		return fmt.Errorf("unsupported lighter ticker symbol: %s", symbol)
	}
}

func (c *Client) StreamTickerByIndex(
	marketIndex int,
	handler exchanges.TickerHandler,
) error {
	conn, _, err := websocket.DefaultDialer.Dial(getWSURL(), nil)
	if err != nil {
		return err
	}
	defer conn.Close()

	subscribeMsg := fmt.Sprintf(
		`{"type":"subscribe","channel":"ticker/%d"}`,
		marketIndex,
	)

	err = conn.WriteMessage(websocket.TextMessage, []byte(subscribeMsg))
	if err != nil {
		return err
	}

	log.Println("Connected to Lighter WS ticker index:", marketIndex)

	go keepAlive(conn)

	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			return err
		}

		var tickerMsg LighterTickerMessage
		err = json.Unmarshal(msg, &tickerMsg)
		if err != nil {
			fmt.Println(string(msg))
			continue
		}

		if !strings.HasPrefix(tickerMsg.Channel, "ticker:") {
			continue
		}

		if tickerMsg.Ticker.Symbol == "" {
			continue
		}

		handler(exchanges.Ticker{
			Venue:   "lighter",
			Symbol:  tickerMsg.Ticker.Symbol,
			Bid:     tickerMsg.Ticker.Bid.Price,
			BidSize: tickerMsg.Ticker.Bid.Size,
			Ask:     tickerMsg.Ticker.Ask.Price,
			AskSize: tickerMsg.Ticker.Ask.Size,
			Time:    tickerMsg.Timestamp,
		})
	}
}

func keepAlive(conn *websocket.Conn) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		err := conn.WriteMessage(websocket.PingMessage, []byte("ping"))
		if err != nil {
			return
		}
	}
}
