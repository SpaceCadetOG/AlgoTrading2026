package hyperliquid

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/gorilla/websocket"

	"AlgoTrading2026/exchanges"
)

type HLCandleWSMessage struct {
	Channel string     `json:"channel"`
	Data    HLCandleWS `json:"data"`
}

type HLCandleWS struct {
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

func (c *Client) StreamCandles(
	symbol string,
	interval string,
	handler exchanges.CandleHandler,
) error {
	conn, _, err := websocket.DefaultDialer.Dial(getWSURL(), nil)
	if err != nil {
		return err
	}
	defer conn.Close()

	subscribeMsg := fmt.Sprintf(`{
		"method":"subscribe",
		"subscription":{
			"type":"candle",
			"coin":"%s",
			"interval":"%s"
		}
	}`, symbol, interval)

	if err := conn.WriteMessage(websocket.TextMessage, []byte(subscribeMsg)); err != nil {
		return err
	}

	log.Println("Connected to Hyperliquid WS candles:", symbol, interval)

	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			return err
		}

		var candleMsg HLCandleWSMessage
		if err := json.Unmarshal(msg, &candleMsg); err != nil {
			fmt.Println(string(msg))
			continue
		}

		if candleMsg.Channel != "candle" {
			continue
		}

		k := candleMsg.Data

		handler(exchanges.Candle{
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
			Closed:    false,
		})
	}
}
