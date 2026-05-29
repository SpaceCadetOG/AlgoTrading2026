package aster

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/big"
	"net/http"
	"net/url"
	"os"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"

	"AlgoTrading2026/config"
	"AlgoTrading2026/exchanges"
)

func getBaseURL() string {
	if config.IsTestnet() {
		return "https://fapi.asterdex-testnet.com"
	}

	return "https://fapi.asterdex.com"
}

type Client struct {
	User       string
	Signer     string
	PrivateKey string

	mu       sync.Mutex
	lastSec  int64
	nonceInc int64
}

type AsterAccount struct {
	TotalWalletBalance    string          `json:"totalWalletBalance"`
	TotalUnrealizedProfit string          `json:"totalUnrealizedProfit"`
	TotalMarginBalance    string          `json:"totalMarginBalance"`
	AvailableBalance      string          `json:"availableBalance"`
	Positions             []AsterPosition `json:"positions"`
}

type AsterPosition struct {
	Symbol           string `json:"symbol"`
	PositionAmt      string `json:"positionAmt"`
	EntryPrice       string `json:"entryPrice"`
	UnrealizedProfit string `json:"unrealizedProfit"`
	Leverage         string `json:"leverage"`
	Isolated         bool   `json:"isolated"`
}

type AsterBalance struct {
	AccountAlias       string `json:"accountAlias"`
	Asset              string `json:"asset"`
	Balance            string `json:"balance"`
	CrossWalletBalance string `json:"crossWalletBalance"`
	AvailableBalance   string `json:"availableBalance"`
}

func NewClient(user, signer, privateKey string) *Client {
	return &Client{
		User:       strings.TrimSpace(user),
		Signer:     strings.TrimSpace(signer),
		PrivateKey: strings.TrimPrefix(strings.TrimSpace(privateKey), "0x"),
	}
}

func (c *Client) Name() string {
	return "aster"
}

func (c *Client) GetAccountSnapshot() (*exchanges.AccountSnapshot, error) {
	account, err := c.GetFuturesAccountStatus()
	if err != nil {
		return nil, err
	}

	balances, err := c.GetBalances()
	if err != nil {
		return nil, err
	}

	positions := normalizeAsterPositions(account.OpenPositions())

	snapshot := &exchanges.AccountSnapshot{
		Venue:       c.Name(),
		AccountID:   c.User,
		WalletValue: account.TotalWalletBalance,
		Available:   account.AvailableBalance,
		OpenPnL:     account.TotalUnrealizedProfit,
		Positions:   positions,
		Balances:    balances,
	}

	return snapshot, nil
}

func (c *Client) GetPositions() ([]exchanges.Position, error) {
	account, err := c.GetFuturesAccountStatus()
	if err != nil {
		return nil, err
	}

	return normalizeAsterPositions(account.OpenPositions()), nil
}

func (c *Client) GetBalances() ([]exchanges.Balance, error) {
	balances, err := c.GetFuturesBalance()
	if err != nil {
		return nil, err
	}

	out := make([]exchanges.Balance, 0, len(balances))
	for _, b := range balances {
		out = append(out, exchanges.Balance{
			Asset:     b.Asset,
			Total:     b.Balance,
			Available: b.AvailableBalance,
		})
	}

	return out, nil
}

func (c *Client) GetFuturesAccountStatus() (*AsterAccount, error) {
	var result AsterAccount

	err := c.signedGetJSON("/fapi/v3/account", nil, &result)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

func (c *Client) GetFuturesBalance() ([]AsterBalance, error) {
	var result []AsterBalance

	err := c.signedGetJSON("/fapi/v3/balance", nil, &result)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (c *Client) GetPositionInformation(symbol string) ([]AsterPosition, error) {
	params := url.Values{}
	if symbol != "" {
		params.Set("symbol", symbol)
	}

	var result []AsterPosition
	err := c.signedJSON(http.MethodGet, "/fapi/v3/positionRisk", params, &result)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (a *AsterAccount) OpenPositions() []AsterPosition {
	var open []AsterPosition

	for _, p := range a.Positions {
		amt, err := strconv.ParseFloat(p.PositionAmt, 64)
		if err != nil {
			continue
		}

		if amt != 0 {
			open = append(open, p)
		}
	}

	return open
}

func normalizeAsterPositions(positions []AsterPosition) []exchanges.Position {
	out := make([]exchanges.Position, 0, len(positions))
	for _, p := range positions {
		out = append(out, exchanges.Position{
			Venue:  "aster",
			Symbol: p.Symbol,
			Side:   "BOTH",
			Size:   p.PositionAmt,
			Entry:  p.EntryPrice,
			PnL:    p.UnrealizedProfit,
			Lev:    p.Leverage,
		})
	}

	return out
}

func (c *Client) signedGetJSON(path string, params url.Values, out any) error {
	return c.signedJSON(http.MethodGet, path, params, out)
}

func (c *Client) signedJSON(method string, path string, params url.Values, out any) error {
	body, err := c.signedRequest(method, path, params)
	if err != nil {
		return err
	}

	return json.Unmarshal(body, out)
}

func (c *Client) signedGet(path string, params url.Values) (string, error) {
	body, err := c.signedRequest(http.MethodGet, path, params)
	if err != nil {
		return "", err
	}

	return string(body), nil
}

func (c *Client) signedRequest(method string, path string, params url.Values) ([]byte, error) {
	if params == nil {
		params = url.Values{}
	}

	params.Set("nonce", c.nextNonce())
	params.Set("signer", c.Signer)

	if c.User != "" {
		params.Set("user", c.User)
	}

	queryString := params.Encode()
	if config.IsTestnet() {
		queryString = asterTestnetQueryString(params)
	}

	signature, err := c.signAsterMessage(queryString)
	if err != nil {
		return nil, err
	}

	fullURL := getBaseURL() + path + "?" + queryString + "&signature=" + url.QueryEscape(signature)

	req, err := http.NewRequest(method, fullURL, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("User-Agent", "GoAsterLearning/1.0")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("bad status %d: %s", resp.StatusCode, string(body))
	}

	return body, nil
}

func (c *Client) nextNonce() string {
	c.mu.Lock()
	defer c.mu.Unlock()

	nowSec := time.Now().Unix()

	if nowSec == c.lastSec {
		c.nonceInc++
	} else {
		c.lastSec = nowSec
		c.nonceInc = 0
	}

	nonce := nowSec*1_000_000 + c.nonceInc
	return fmt.Sprintf("%d", nonce)
}

func (c *Client) signAsterMessage(queryString string) (string, error) {
	if c.PrivateKey == "" {
		return "", fmt.Errorf("missing Aster private key for %s", config.TradingEnv())
	}

	privateKey, err := crypto.HexToECDSA(c.PrivateKey)
	if err != nil {
		return "", fmt.Errorf("invalid Aster private key for %s: %w", config.TradingEnv(), err)
	}

	digest := asterTypedDataDigest(queryString)

	sig, err := crypto.Sign(digest, privateKey)
	if err != nil {
		return "", err
	}

	sig[64] += 27
	c.debugAsterSignature(queryString, digest, sig)

	return "0x" + hex.EncodeToString(sig), nil
}

func (c *Client) debugAsterSignature(queryString string, digest []byte, sig []byte) {
	if os.Getenv("ASTER_DEBUG_SIGNING") != "true" {
		return
	}

	sigForRecovery := append([]byte(nil), sig...)
	if sigForRecovery[64] >= 27 {
		sigForRecovery[64] -= 27
	}

	recovered := "<unavailable>"
	pub, err := crypto.SigToPub(digest, sigForRecovery)
	if err == nil {
		recovered = crypto.PubkeyToAddress(*pub).Hex()
	}

	log.Printf("ASTER signing env=%s chainId=%d", config.TradingEnv(), asterChainID())
	log.Printf("ASTER signing query=%s", queryString)
	log.Printf("ASTER signing domainSeparator=0x%s", hex.EncodeToString(asterDomainSeparator()))
	log.Printf("ASTER signing messageHash=0x%s", hex.EncodeToString(asterMessageHash(queryString)))
	log.Printf("ASTER signing digest=0x%s", hex.EncodeToString(digest))
	log.Printf("ASTER signing configuredSigner=%s recoveredSigner=%s", c.Signer, recovered)
}

func asterTestnetQueryString(params url.Values) string {
	orderedKeys := []string{
		"symbol",
		"side",
		"type",
		"quantity",
		"price",
		"timeInForce",
		"leverage",
		"orderId",
		"stopPrice",
		"nonce",
		"user",
		"signer",
	}

	used := make(map[string]bool, len(orderedKeys))
	var parts []string

	for _, key := range orderedKeys {
		values, ok := params[key]
		if !ok {
			continue
		}
		used[key] = true
		appendQueryParts(&parts, key, values)
	}

	var rest []string
	for key := range params {
		if !used[key] {
			rest = append(rest, key)
		}
	}
	sort.Strings(rest)

	for _, key := range rest {
		appendQueryParts(&parts, key, params[key])
	}

	return strings.Join(parts, "&")
}

func appendQueryParts(parts *[]string, key string, values []string) {
	sort.Strings(values)

	for _, value := range values {
		*parts = append(*parts, url.QueryEscape(key)+"="+url.QueryEscape(value))
	}
}

func asterTypedDataDigest(msg string) []byte {
	domainSeparator := asterDomainSeparator()
	messageHash := asterMessageHash(msg)

	raw := []byte{0x19, 0x01}
	raw = append(raw, domainSeparator...)
	raw = append(raw, messageHash...)

	return crypto.Keccak256(raw)
}

func asterDomainSeparator() []byte {
	domainType := "EIP712Domain(string name,string version,uint256 chainId,address verifyingContract)"

	typeHash := crypto.Keccak256([]byte(domainType))
	nameHash := crypto.Keccak256([]byte("AsterSignTransaction"))
	versionHash := crypto.Keccak256([]byte("1"))

	chainID := new(big.Int).SetInt64(asterChainID())
	verifyingContract := common.HexToAddress("0x0000000000000000000000000000000000000000")

	var encoded []byte
	encoded = append(encoded, typeHash...)
	encoded = append(encoded, nameHash...)
	encoded = append(encoded, versionHash...)
	encoded = append(encoded, leftPad32(chainID.Bytes())...)
	encoded = append(encoded, leftPad32(verifyingContract.Bytes())...)

	return crypto.Keccak256(encoded)
}

func asterChainID() int64 {
	if config.IsTestnet() {
		return 714
	}

	return 1666
}

func asterMessageHash(msg string) []byte {
	messageType := "Message(string msg)"

	typeHash := crypto.Keccak256([]byte(messageType))
	msgHash := crypto.Keccak256([]byte(msg))

	var encoded []byte
	encoded = append(encoded, typeHash...)
	encoded = append(encoded, msgHash...)

	return crypto.Keccak256(encoded)
}

func leftPad32(input []byte) []byte {
	if len(input) >= 32 {
		return input
	}

	out := make([]byte, 32)
	copy(out[32-len(input):], input)

	return out
}
