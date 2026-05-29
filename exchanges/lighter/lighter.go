package lighter

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"AlgoTrading2026/exchanges"
)

type Client struct {
	Address string
}

type AccountsByL1Address struct {
	Code        int              `json:"code"`
	L1Address   string           `json:"l1_address"`
	SubAccounts []LighterAccount `json:"sub_accounts"`
}

type LighterAccount struct {
	Code             int    `json:"code"`
	AccountType      int    `json:"account_type"`
	Index            int    `json:"index"`
	L1Address        string `json:"l1_address"`
	AvailableBalance string `json:"available_balance"`
	Status           int    `json:"status"`
	Collateral       string `json:"collateral"`
}

func NewClient(address string) *Client {
	return &Client{
		Address: strings.TrimSpace(address),
	}
}

func (c *Client) Name() string {
	return "lighter"
}

func (c *Client) GetAccountSnapshot() (*exchanges.AccountSnapshot, error) {
	balances, err := c.GetBalances()
	if err != nil {
		return nil, err
	}

	positions, err := c.GetPositions()
	if err != nil {
		return nil, err
	}

	snapshot := &exchanges.AccountSnapshot{
		Venue:     c.Name(),
		AccountID: c.Address,
		OpenPnL:   "0",
		Balances:  balances,
		Positions: positions,
	}

	for _, balance := range balances {
		if snapshot.WalletValue == "" {
			snapshot.WalletValue = balance.Total
		}
		if snapshot.Available == "" {
			if balance.Available != "" {
				snapshot.Available = balance.Available
			} else {
				snapshot.Available = balance.Total
			}
		}
	}

	return snapshot, nil
}

func (c *Client) GetBalances() ([]exchanges.Balance, error) {
	accounts, err := GetAccountsByL1Address(c.Address)
	if err != nil {
		return nil, err
	}

	balances := make([]exchanges.Balance, 0, len(accounts.SubAccounts))
	for _, acct := range accounts.SubAccounts {
		balances = append(balances, exchanges.Balance{
			Asset:     "USDC",
			Total:     acct.Collateral,
			Available: acct.AvailableBalance,
		})
	}

	return balances, nil
}

func GetAccountsByL1Address(address string) (*AccountsByL1Address, error) {
	endpoint := getBaseURL() + "/api/v1/accountsByL1Address"

	params := url.Values{}
	params.Set("l1_address", strings.TrimSpace(address))

	resp, err := http.Get(endpoint + "?" + params.Encode())
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

	var result AccountsByL1Address
	err = json.Unmarshal(body, &result)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

func (c *Client) GetConfiguredAccount() (*LighterAccount, error) {
	cfg, err := LoadExecutionConfig()
	if err != nil {
		return nil, err
	}

	params := url.Values{}
	params.Set("index", strconv.Itoa(cfg.AccountIndex))

	resp, err := http.Get(getBaseURL() + "/api/v1/account?" + params.Encode())
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("lighter account bad status %d: %s", resp.StatusCode, string(body))
	}

	var account LighterAccount
	if err := json.Unmarshal(body, &account); err == nil && account.Index != 0 {
		return &account, nil
	}

	var wrapped struct {
		Account LighterAccount `json:"account"`
	}
	if err := json.Unmarshal(body, &wrapped); err != nil {
		return nil, err
	}

	return &wrapped.Account, nil
}
