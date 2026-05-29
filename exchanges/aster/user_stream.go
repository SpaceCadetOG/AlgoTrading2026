package aster

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

type AsterListenKeyResponse struct {
	ListenKey string `json:"listenKey"`
}

type AsterUserStreamEvent struct {
	EventType       string           `json:"e"`
	EventTime       int64            `json:"E"`
	TransactionTime int64            `json:"T,omitempty"`
	Order           *AsterOrderEvent `json:"o,omitempty"`
	Account         json.RawMessage  `json:"a,omitempty"`
	Raw             json.RawMessage  `json:"-"`
}

type AsterOrderEvent struct {
	Symbol               string `json:"s"`
	ClientOrderID        string `json:"c"`
	Side                 string `json:"S"`
	OrderType            string `json:"o"`
	TimeInForce          string `json:"f"`
	OriginalQty          string `json:"q"`
	OriginalPrice        string `json:"p"`
	AveragePrice         string `json:"ap"`
	StopPrice            string `json:"sp"`
	ExecutionType        string `json:"x"`
	OrderStatus          string `json:"X"`
	OrderID              int64  `json:"i"`
	LastFilledQty        string `json:"l"`
	FilledAccumulatedQty string `json:"z"`
	LastFilledPrice      string `json:"L"`
	CommissionAsset      string `json:"N,omitempty"`
	Commission           string `json:"n,omitempty"`
	TradeTime            int64  `json:"T"`
	TradeID              int64  `json:"t"`
	ReduceOnly           bool   `json:"R"`
	WorkingType          string `json:"wt"`
	OriginalOrderType    string `json:"ot"`
	PositionSide         string `json:"ps"`
	ClosePosition        bool   `json:"cp"`
	RealizedProfit       string `json:"rp"`
}

type AsterUserStreamHandler func(AsterUserStreamEvent)

func (c *Client) StartUserDataStream() (string, error) {
	var result AsterListenKeyResponse
	err := c.signedJSON(http.MethodPost, "/fapi/v3/listenKey", nil, &result)
	if err != nil {
		return "", err
	}
	if result.ListenKey == "" {
		return "", fmt.Errorf("aster listenKey response missing listenKey")
	}

	return result.ListenKey, nil
}

func (c *Client) KeepaliveUserDataStream() error {
	_, err := c.signedRequest(http.MethodPut, "/fapi/v3/listenKey", nil)
	return err
}

func (c *Client) CloseUserDataStream() error {
	_, err := c.signedRequest(http.MethodDelete, "/fapi/v3/listenKey", nil)
	return err
}

func (c *Client) StreamUserData(ctx context.Context, listenKey string, handler AsterUserStreamHandler) error {
	url := getWSBaseURL() + "/" + listenKey
	conn, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		return err
	}
	defer conn.Close()

	log.Println("Connected to Aster user data stream")

	errCh := make(chan error, 1)
	go func() {
		<-ctx.Done()
		errCh <- conn.Close()
	}()

	for {
		select {
		case <-ctx.Done():
			return nil
		case err := <-errCh:
			if ctx.Err() != nil {
				return nil
			}
			return err
		default:
		}

		_, msg, err := conn.ReadMessage()
		if err != nil {
			if ctx.Err() != nil {
				return nil
			}
			return err
		}

		var event AsterUserStreamEvent
		event.Raw = append(event.Raw[:0], msg...)
		if err := json.Unmarshal(msg, &event); err != nil {
			log.Printf("Aster user stream decode error: %v raw=%s", err, string(msg))
			continue
		}

		if handler != nil {
			handler(event)
		}
	}
}
