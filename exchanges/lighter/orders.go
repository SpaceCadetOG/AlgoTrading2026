package lighter

import (
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	lighterclient "github.com/elliottech/lighter-go/client"
	lighterhttp "github.com/elliottech/lighter-go/client/http"
	lightertypes "github.com/elliottech/lighter-go/types"
	"github.com/elliottech/lighter-go/types/txtypes"

	"AlgoTrading2026/execution"
)

const lighterChainID = uint32(304)

type SendTxResult struct {
	Code    int             `json:"code"`
	Message string          `json:"message"`
	TxHash  string          `json:"tx_hash"`
	Hash    string          `json:"hash"`
	Raw     json.RawMessage `json:"-"`
}

type LighterLifecycleResult struct {
	SendTx           SendTxResult `json:"send_tx"`
	ClientOrderIndex int64        `json:"client_order_index"`
	TxType           uint8        `json:"tx_type"`
	TxInfoLength     int          `json:"tx_info_length"`
}

type LighterCancelResult struct {
	SendTx       SendTxResult `json:"send_tx"`
	OrderIndex   int64        `json:"order_index"`
	TxType       uint8        `json:"tx_type"`
	TxInfoLength int          `json:"tx_info_length"`
}

func (c *Client) PlaceOrder(order execution.OrderRequest) (*execution.OrderResult, error) {
	if os.Getenv("ENABLE_LIVE_ORDERS") != "true" {
		fmt.Println("LIVE ORDERS DISABLED")

		return &execution.OrderResult{
			Success: true,
			Venue:   "lighter",
			Symbol:  order.Symbol,
			Status:  "DRY_RUN",
			Message: "live orders disabled",
		}, nil
	}

	cfg, err := LoadExecutionConfig()
	if err != nil {
		return nil, err
	}

	txClient, accountIndex, apiKeyIndex, err := newTxClient(cfg)
	if err != nil {
		return nil, err
	}

	market, err := c.ResolveMarket(order.Symbol)
	if err != nil {
		return nil, err
	}

	priceInt, baseAmount, err := market.WirePriceSize(order.Price, order.Size)
	if err != nil {
		return nil, err
	}

	isAsk := uint8(0)
	if order.Side == execution.Sell {
		isAsk = 1
	}

	clientOrderIndex := time.Now().UnixMilli()
	req := &lightertypes.CreateOrderTxReq{
		MarketIndex:      market.ID,
		ClientOrderIndex: clientOrderIndex,
		BaseAmount:       baseAmount,
		Price:            priceInt,
		IsAsk:            isAsk,
		Type:             txtypes.LimitOrder,
		TimeInForce:      txtypes.GoodTillTime,
		ReduceOnly:       boolToUint8(order.ReduceOnly),
		TriggerPrice:     txtypes.NilOrderTriggerPrice,
		OrderExpiry:      time.Now().Add(10 * time.Minute).UnixMilli(),
	}

	opts := transactOpts(accountIndex, apiKeyIndex)
	tx, err := txClient.GetCreateOrderTransaction(req, opts)
	if err != nil {
		return nil, fmt.Errorf("lighter create order tx: %w", err)
	}

	txInfo, err := tx.GetTxInfo()
	if err != nil {
		return nil, fmt.Errorf("lighter get create order tx info: %w", err)
	}

	txType := tx.GetTxType()
	debugLighterTx("createOrder", cfg, txType, len(txInfo), clientOrderIndex)

	if os.Getenv("ENABLE_LIGHTER_SENDTX") != "true" {
		return &execution.OrderResult{
			Success: true,
			Venue:   "lighter",
			Symbol:  order.Symbol,
			OrderID: strconv.FormatInt(clientOrderIndex, 10),
			Status:  "SIGNED_DRY_RUN",
			Message: fmt.Sprintf("lighter order tx signed/constructed tx_type=%d", txType),
			Raw: LighterLifecycleResult{
				ClientOrderIndex: clientOrderIndex,
				TxType:           txType,
				TxInfoLength:     len(txInfo),
			},
		}, nil
	}

	sendResult, err := sendTx(txType, txInfo)
	status := lighterStatus(sendResult)
	result := LighterLifecycleResult{
		SendTx:           sendResult,
		ClientOrderIndex: clientOrderIndex,
		TxType:           txType,
		TxInfoLength:     len(txInfo),
	}
	if err != nil {
		return &execution.OrderResult{
			Success: false,
			Venue:   "lighter",
			Symbol:  order.Symbol,
			OrderID: strconv.FormatInt(clientOrderIndex, 10),
			Status:  status,
			Message: err.Error(),
			Raw:     result,
		}, err
	}

	return &execution.OrderResult{
		Success: true,
		Venue:   "lighter",
		Symbol:  order.Symbol,
		OrderID: strconv.FormatInt(clientOrderIndex, 10),
		Status:  status,
		Message: sendResult.Message,
		Raw:     result,
	}, nil
}

func (c *Client) CancelOrder(symbol string, orderIndex int64) (*LighterCancelResult, error) {
	cfg, err := LoadExecutionConfig()
	if err != nil {
		return nil, err
	}

	txClient, accountIndex, apiKeyIndex, err := newTxClient(cfg)
	if err != nil {
		return nil, err
	}

	market, err := c.ResolveMarket(symbol)
	if err != nil {
		return nil, err
	}

	req := &lightertypes.CancelOrderTxReq{
		MarketIndex: market.ID,
		Index:       orderIndex,
	}
	tx, err := txClient.GetCancelOrderTransaction(req, transactOpts(accountIndex, apiKeyIndex))
	if err != nil {
		return nil, fmt.Errorf("lighter cancel order tx: %w", err)
	}

	txInfo, err := tx.GetTxInfo()
	if err != nil {
		return nil, fmt.Errorf("lighter get cancel tx info: %w", err)
	}

	txType := tx.GetTxType()
	debugLighterTx("cancelOrder", cfg, txType, len(txInfo), orderIndex)

	sendResult, err := sendTx(txType, txInfo)
	result := &LighterCancelResult{
		SendTx:       sendResult,
		OrderIndex:   orderIndex,
		TxType:       txType,
		TxInfoLength: len(txInfo),
	}
	if err != nil {
		return result, err
	}

	return result, nil
}

func (c *Client) CancelAllOpenOrders() (*LighterCancelResult, error) {
	cfg, err := LoadExecutionConfig()
	if err != nil {
		return nil, err
	}

	txClient, accountIndex, apiKeyIndex, err := newTxClient(cfg)
	if err != nil {
		return nil, err
	}

	req := &lightertypes.CancelAllOrdersTxReq{
		TimeInForce: txtypes.ImmediateCancelAll,
		Time:        txtypes.NilOrderExpiry,
	}
	tx, err := txClient.GetCancelAllOrdersTransaction(req, transactOpts(accountIndex, apiKeyIndex))
	if err != nil {
		return nil, fmt.Errorf("lighter cancel all orders tx: %w", err)
	}

	txInfo, err := tx.GetTxInfo()
	if err != nil {
		return nil, fmt.Errorf("lighter get cancel all tx info: %w", err)
	}

	txType := tx.GetTxType()
	debugLighterTx("cancelAllOrders", cfg, txType, len(txInfo), 0)

	sendResult, err := sendTx(txType, txInfo)
	result := &LighterCancelResult{
		SendTx:       sendResult,
		TxType:       txType,
		TxInfoLength: len(txInfo),
	}
	if err != nil {
		return result, err
	}

	return result, nil
}

func newTxClient(cfg *LighterExecutionConfig) (*lighterclient.TxClient, int64, uint8, error) {
	accountIndex := int64(cfg.AccountIndex)
	apiKeyIndex := uint8(cfg.APIKeyIndex)
	httpClient := lighterhttp.NewClient(getBaseURL())

	txClient, err := lighterclient.NewTxClient(
		httpClient,
		cfg.APIPrivateKey,
		accountIndex,
		apiKeyIndex,
		lighterChainID,
	)
	if err != nil {
		return nil, 0, 0, fmt.Errorf("new lighter tx client: %w", err)
	}

	return txClient, accountIndex, apiKeyIndex, nil
}

func transactOpts(accountIndex int64, apiKeyIndex uint8) *lightertypes.TransactOpts {
	return &lightertypes.TransactOpts{
		FromAccountIndex: &accountIndex,
		ApiKeyIndex:      &apiKeyIndex,
		ExpiredAt:        time.Now().Add(10 * time.Minute).UnixMilli(),
	}
}

func sendTx(txType uint8, txInfo string) (SendTxResult, error) {
	form := url.Values{}
	form.Set("tx_type", strconv.Itoa(int(txType)))
	form.Set("tx_info", txInfo)
	form.Set("price_protection", "true")

	sendReq, err := http.NewRequest(
		http.MethodPost,
		getBaseURL()+"/api/v1/sendTx",
		strings.NewReader(form.Encode()),
	)
	if err != nil {
		return SendTxResult{}, err
	}

	sendReq.Header.Set("accept", "application/json")
	sendReq.Header.Set("content-type", "application/x-www-form-urlencoded")

	res, err := http.DefaultClient.Do(sendReq)
	if err != nil {
		return SendTxResult{}, err
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return SendTxResult{}, err
	}

	result := SendTxResult{Raw: append(json.RawMessage(nil), body...)}
	_ = json.Unmarshal(body, &result)
	debugLighterSendTx(res.StatusCode, body)

	if res.StatusCode < 200 || res.StatusCode >= 300 {
		if result.Message == "" {
			result.Message = string(body)
		}
		return result, fmt.Errorf("lighter sendTx bad status %d: %s", res.StatusCode, string(body))
	}

	if result.Message == "" {
		result.Message = string(body)
	}

	return result, nil
}

func lighterStatus(result SendTxResult) string {
	if result.Code != 0 {
		return fmt.Sprintf("code_%d", result.Code)
	}
	if result.Message != "" {
		return result.Message
	}
	return "submitted"
}

func boolToUint8(value bool) uint8 {
	if value {
		return 1
	}
	return 0
}

func debugLighterTx(action string, cfg *LighterExecutionConfig, txType uint8, txInfoLength int, orderIndex int64) {
	if os.Getenv("LIGHTER_DEBUG_SIGNING") != "true" {
		return
	}

	cleanKey := strings.TrimPrefix(strings.TrimSpace(cfg.APIPrivateKey), "0x")
	fmt.Printf(
		"LIGHTER debug action=%s accountIndex=%d apiKeyIndex=%d apiKeyChars=%d txType=%d txInfoLength=%d orderIndex=%d endpoint=%s\n",
		action,
		cfg.AccountIndex,
		cfg.APIKeyIndex,
		len(cleanKey),
		txType,
		txInfoLength,
		orderIndex,
		getBaseURL()+"/api/v1/sendTx",
	)
}

func debugLighterSendTx(statusCode int, body []byte) {
	if os.Getenv("LIGHTER_DEBUG_SIGNING") != "true" {
		return
	}

	fmt.Printf("LIGHTER debug sendTxStatus=%d sendTxResponse=%s\n", statusCode, string(body))
}

func pow10Int(decimals int) float64 {
	return math.Pow10(decimals)
}
