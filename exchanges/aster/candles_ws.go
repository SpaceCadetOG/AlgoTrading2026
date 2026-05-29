package aster

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"github.com/gorilla/websocket"

	"AlgoTrading2026/config"
	"AlgoTrading2026/exchanges"
)

func getWSBaseURL() string {
	if config.IsTestnet() {
		return "wss://fstream.asterdex-testnet.com/ws"
	}

	return "wss://fstream.asterdex.com/ws"
}

type FlexString string

func (f *FlexString) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err == nil {
		*f = FlexString(s)
		return nil
	}

	var n float64
	if err := json.Unmarshal(data, &n); err == nil {
		*f = FlexString(fmt.Sprintf("%f", n))
		return nil
	}

	return fmt.Errorf("unsupported value: %s", string(data))
}

func (f FlexString) String() string {
	return string(f)
}

type AsterKlineMessage struct {
	EventType string `json:"e"`
	EventTime int64  `json:"E"`
	Symbol    string `json:"s"`
	Kline     struct {
		StartTime int64      `json:"t"`
		EndTime   int64      `json:"T"`
		Interval  string     `json:"i"`
		Open      FlexString `json:"o"`
		Close     FlexString `json:"c"`
		High      FlexString `json:"h"`
		Low       FlexString `json:"l"`
		Volume    FlexString `json:"v"`
		IsClosed  bool       `json:"x"`
	} `json:"k"`
}

func (c *Client) StreamCandles(symbol string, interval string, handler exchanges.CandleHandler) error {
	stream := strings.ToLower(symbol) + "@kline_" + interval
	url := getWSBaseURL() + "/" + stream

	conn, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		return err
	}
	defer conn.Close()

	log.Println("Connected to Aster WS candles:", stream)

	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			return err
		}

		var klineMsg AsterKlineMessage
		if err := json.Unmarshal(msg, &klineMsg); err != nil {
			fmt.Println("Aster candle decode error:", err)
			continue
		}

		handler(exchanges.Candle{
			Venue:     c.Name(),
			Symbol:    klineMsg.Symbol,
			Interval:  klineMsg.Kline.Interval,
			Open:      klineMsg.Kline.Open.String(),
			High:      klineMsg.Kline.High.String(),
			Low:       klineMsg.Kline.Low.String(),
			Close:     klineMsg.Kline.Close.String(),
			Volume:    klineMsg.Kline.Volume.String(),
			StartTime: klineMsg.Kline.StartTime,
			EndTime:   klineMsg.Kline.EndTime,
			Closed:    klineMsg.Kline.IsClosed,
		})
	}
}