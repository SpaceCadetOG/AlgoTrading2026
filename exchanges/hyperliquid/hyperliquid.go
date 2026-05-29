package hyperliquid

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"strconv"
	"strings"

	"AlgoTrading2026/exchanges"
)

type Client struct {
	Address    string
	PrivateKey string
}

type MarginSummary struct {
	AccountValue    string `json:"accountValue"`
	TotalNtlPos     string `json:"totalNtlPos"`
	TotalRawUsd     string `json:"totalRawUsd"`
	TotalMarginUsed string `json:"totalMarginUsed"`
}

type PerpAccountStatus struct {
	MarginSummary      MarginSummary `json:"marginSummary"`
	CrossMarginSummary MarginSummary `json:"crossMarginSummary"`
	Withdrawable       string        `json:"withdrawable"`
	AssetPositions     []HLPosition  `json:"assetPositions"`
	Time               int64         `json:"time"`
}

type HLPosition struct {
	Type     string `json:"type"`
	Position struct {
		Coin     string  `json:"coin"`
		EntryPx  *string `json:"entryPx"`
		Leverage struct {
			Type  string `json:"type"`
			Value int    `json:"value"`
		} `json:"leverage"`
		LiquidationPx  *string `json:"liquidationPx"`
		MarginUsed     string  `json:"marginUsed"`
		PositionValue  string  `json:"positionValue"`
		ReturnOnEquity string  `json:"returnOnEquity"`
		Szi            string  `json:"szi"`
		UnrealizedPnl  string  `json:"unrealizedPnl"`
	} `json:"position"`
}

type SpotBalances struct {
	Balances []SpotBalance `json:"balances"`
}

type SpotBalance struct {
	Coin     string `json:"coin"`
	Total    string `json:"total"`
	Hold     string `json:"hold"`
	EntryNtl string `json:"entryNtl"`
}

func NewClient(address string, privateKey ...string) *Client {
	key := ""
	if len(privateKey) > 0 {
		key = privateKey[0]
	}

	return &Client{
		Address:    strings.TrimSpace(address),
		PrivateKey: strings.TrimPrefix(strings.TrimSpace(key), "0x"),
	}
}

func (c *Client) Name() string {
	return "hyperliquid"
}

func (c *Client) GetAccountSnapshot() (*exchanges.AccountSnapshot, error) {
	perps, err := c.GetPerpAccountStatus()
	if err != nil {
		return nil, err
	}

	balances, err := c.GetBalances()
	if err != nil {
		return nil, err
	}

	positions := normalizeHLPositions(perps.AssetPositions)

	snapshot := &exchanges.AccountSnapshot{
		Venue:       c.Name(),
		AccountID:   c.Address,
		WalletValue: perps.MarginSummary.AccountValue,
		Available:   perps.Withdrawable,
		OpenPnL:     sumPositionPnL(positions),
		Positions:   positions,
		Balances:    balances,
	}

	return snapshot, nil
}

func postInfoJSON(jsonBody string, out any) error {
	return postInfoJSONAt(getBaseURL(), jsonBody, out)
}

func postInfoJSONAt(baseURL string, jsonBody string, out any) error {
	resp, err := http.Post(
		baseURL+"/info",
		"application/json",
		bytes.NewBuffer([]byte(jsonBody)),
	)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("bad status %d: %s", resp.StatusCode, string(body))
	}

	debugHLREST(baseURL+"/info", jsonBody, body)

	return json.Unmarshal(body, out)
}

func (c *Client) GetPerpAccountStatus() (*PerpAccountStatus, error) {
	return GetPerpAccountStatus(c.Address)
}

func GetPerpAccountStatus(address string) (*PerpAccountStatus, error) {
	var result PerpAccountStatus

	jsonBody := fmt.Sprintf(`{"type":"clearinghouseState","user":"%s"}`, address)
	err := postInfoJSON(jsonBody, &result)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

func (c *Client) GetSpotBalances() (*SpotBalances, error) {
	return GetSpotBalances(c.Address)
}

func GetSpotBalances(address string) (*SpotBalances, error) {
	var result SpotBalances

	jsonBody := fmt.Sprintf(`{"type":"spotClearinghouseState","user":"%s"}`, address)
	err := postInfoJSON(jsonBody, &result)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

func (c *Client) GetRawPositions() ([]HLPosition, error) {
	perps, err := c.GetPerpAccountStatus()
	if err != nil {
		return nil, err
	}

	positions := make([]HLPosition, 0, len(perps.AssetPositions))
	for _, p := range perps.AssetPositions {
		if p.Position.Szi != "0" {
			positions = append(positions, p)
		}
	}

	return positions, nil
}

func (c *Client) GetPositions() ([]exchanges.Position, error) {
	positions, err := c.GetRawPositions()
	if err != nil {
		return nil, err
	}

	return normalizeHLPositions(positions), nil
}

func (c *Client) GetBalances() ([]exchanges.Balance, error) {
	spot, err := c.GetSpotBalances()
	if err != nil {
		return nil, err
	}

	out := make([]exchanges.Balance, 0, len(spot.Balances))
	for _, b := range spot.Balances {
		out = append(out, exchanges.Balance{
			Asset:     b.Coin,
			Total:     b.Total,
			Available: b.Total,
		})
	}

	return out, nil
}

func (c *Client) GetAllMids() (map[string]string, error) {
	var result map[string]string
	err := postInfoJSON(`{"type":"allMids"}`, &result)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func normalizeHLPositions(positions []HLPosition) []exchanges.Position {
	out := make([]exchanges.Position, 0, len(positions))
	for _, p := range positions {
		if p.Position.Szi == "0" {
			continue
		}

		entry := ""
		if p.Position.EntryPx != nil {
			entry = *p.Position.EntryPx
		}

		out = append(out, exchanges.Position{
			Venue:  "hyperliquid",
			Symbol: p.Position.Coin,
			Side:   hlPositionSide(p.Position.Szi),
			Size:   p.Position.Szi,
			Entry:  entry,
			PnL:    p.Position.UnrealizedPnl,
			Lev:    fmt.Sprintf("%d", p.Position.Leverage.Value),
		})
	}

	return out
}

func hlPositionSide(size string) string {
	value, err := strconv.ParseFloat(size, 64)
	if err != nil {
		return "BOTH"
	}
	if value > 0 {
		return "LONG"
	}
	if value < 0 {
		return "SHORT"
	}

	return "FLAT"
}

func sumPositionPnL(positions []exchanges.Position) string {
	total := 0.0
	for _, p := range positions {
		value, err := strconv.ParseFloat(p.PnL, 64)
		if err == nil && !math.IsNaN(value) && !math.IsInf(value, 0) {
			total += value
		}
	}

	return strconv.FormatFloat(total, 'f', -1, 64)
}
