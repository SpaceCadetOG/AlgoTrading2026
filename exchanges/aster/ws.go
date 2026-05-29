package aster

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"github.com/gorilla/websocket"

	"AlgoTrading2026/exchanges"
)

const wsBaseURL = "wss://fstream.asterdex.com/ws"

type AsterTradeMessage struct {
	EventType    string `json:"e"`
	EventTime    int64  `json:"E"`
	Symbol       string `json:"s"`
	TradeID      int64  `json:"t"`
	Price        string `json:"p"`
	Quantity     string `json:"q"`
	TradeTime    int64  `json:"T"`
	IsBuyerMaker bool   `json:"m"`
}

func (c *Client) StreamTrades(
	symbol string,
	handler exchanges.TradeHandler,
) error {
	stream := strings.ToLower(symbol) + "@trade"
	url := wsBaseURL + "/" + stream

	conn, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		return err
	}
	defer conn.Close()

	log.Println("Connected to Aster WS trades:", stream)

	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			return err
		}

		var tradeMsg AsterTradeMessage
		err = json.Unmarshal(msg, &tradeMsg)
		if err != nil {
			fmt.Println(string(msg))
			continue
		}

		side := "B"
		if tradeMsg.IsBuyerMaker {
			side = "A"
		}

		handler(exchanges.Trade{
			Venue:  "aster",
			Symbol: tradeMsg.Symbol,
			Side:   side,
			Price:  tradeMsg.Price,
			Size:   tradeMsg.Quantity,
			Time:   tradeMsg.TradeTime,
		})
	}
}
