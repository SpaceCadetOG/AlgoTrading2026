package hyperliquid

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/gorilla/websocket"

	"AlgoTrading2026/exchanges"
)

type TradeMessage struct {
	Channel string    `json:"channel"`
	Data    []HLTrade `json:"data"`
}

type HLTrade struct {
	Coin string `json:"coin"`
	Side string `json:"side"`
	Px   string `json:"px"`
	Sz   string `json:"sz"`
	Time int64  `json:"time"`
}

func (c *Client) StreamTrades(
	symbol string,
	handler exchanges.TradeHandler,
) error {
	conn, _, err := websocket.DefaultDialer.Dial(getWSURL(), nil)
	if err != nil {
		return err
	}
	defer conn.Close()

	subscribeMsg := fmt.Sprintf(`{
		"method":"subscribe",
		"subscription":{
			"type":"trades",
			"coin":"%s"
		}
	}`, symbol)

	err = conn.WriteMessage(websocket.TextMessage, []byte(subscribeMsg))
	if err != nil {
		return err
	}

	log.Println("Connected to Hyperliquid WS trades:", symbol)

	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			return err
		}

		var tradeMsg TradeMessage
		err = json.Unmarshal(msg, &tradeMsg)
		if err != nil {
			fmt.Println(string(msg))
			continue
		}

		if tradeMsg.Channel != "trades" {
			continue
		}

		for _, t := range tradeMsg.Data {
			handler(exchanges.Trade{
				Venue:  "hyperliquid",
				Symbol: t.Coin,
				Side:   t.Side,
				Price:  t.Px,
				Size:   t.Sz,
				Time:   t.Time,
			})
		}
	}
}
